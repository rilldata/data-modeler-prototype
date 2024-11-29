package rduckdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
	"golang.org/x/sync/errgroup"
)

// pullFromRemote updates local data with the latest data from remote.
// This is not safe for concurrent calls.
func (d *db) pullFromRemote(ctx context.Context) error {
	if !d.localDirty {
		// optimisation to skip sync if write was already synced
		return nil
	}
	d.logger.Debug("syncing from remote")
	// Create an errgroup for background downloads with limited concurrency.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	objects := d.remote.List(&blob.ListOptions{
		Delimiter: "/", // only list directories with a trailing slash and IsDir set to true
	})

	tblMetas := make(map[string]*tableMeta)
	for {
		// Stop the loop if the ctx was cancelled
		var stop bool
		select {
		case <-gctx.Done():
			stop = true
		default:
			// don't break
		}
		if stop {
			break // can't use break inside the select
		}

		obj, err := objects.Next(gctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if !obj.IsDir {
			continue
		}

		table := strings.TrimSuffix(obj.Key, "/")
		d.logger.Debug("SyncWithObjectStorage: discovered table", slog.String("table", table))

		// get version of the table
		var b []byte
		err = retry(gctx, func() error {
			res, err := d.remote.ReadAll(gctx, path.Join(table, "meta.json"))
			if err != nil {
				return err
			}
			b = res
			return nil
		})
		if err != nil {
			if gcerrors.Code(err) == gcerrors.NotFound {
				// invalid table directory
				d.logger.Debug("SyncWithObjectStorage: invalid table directory", slog.String("table", table))
				continue
			}
			return err
		}
		backedUpMeta := &tableMeta{}
		err = json.Unmarshal(b, backedUpMeta)
		if err != nil {
			d.logger.Debug("SyncWithObjectStorage: failed to unmarshal table metadata", slog.String("table", table), slog.Any("error", err))
			continue
		}

		// check if table in catalog is already upto date
		meta, _ := d.catalog.tableMeta(table)
		if meta != nil && meta.Version == backedUpMeta.Version {
			d.logger.Debug("SyncWithObjectStorage: table is already up to date", slog.String("table", table))
			continue
		}
		tblMetas[table] = backedUpMeta

		// check if table is locally present but not added to catalog yet
		meta, _ = d.tableMeta(table)
		if meta != nil && meta.Version == backedUpMeta.Version {
			d.logger.Debug("SyncWithObjectStorage: local table is not present in catalog", slog.String("table", table))
			tblMetas[table] = backedUpMeta
			continue
		}
		if err := d.initLocalTable(table, backedUpMeta.Version); err != nil {
			return err
		}

		tblIter := d.remote.List(&blob.ListOptions{Prefix: path.Join(table, backedUpMeta.Version)})
		// download all objects in the table and current version
		for {
			obj, err := tblIter.Next(gctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			g.Go(func() error {
				return retry(gctx, func() error {
					file, err := os.Create(filepath.Join(d.localPath, obj.Key))
					if err != nil {
						return err
					}
					defer file.Close()

					rdr, err := d.remote.NewReader(gctx, obj.Key, nil)
					if err != nil {
						return err
					}
					defer rdr.Close()

					_, err = io.Copy(file, rdr)
					return err
				})
			})
		}
	}

	// Wait for all outstanding downloads to complete
	err := g.Wait()
	if err != nil {
		return err
	}

	// Update table versions
	for table, meta := range tblMetas {
		err = d.writeTableMeta(table, meta)
		if err != nil {
			return err
		}
		d.catalog.addTableVersion(table, meta)
	}

	// mark tables that are not in remote for delete later
	entries, err := os.ReadDir(d.localPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, ok := tblMetas[entry.Name()]; ok {
			continue
		}
		d.catalog.removeTable(entry.Name())
	}
	return nil
}

// pushToRemote syncs the remote location with the local path for given table.
// If oldVersion is specified, it is deleted after successful sync.
func (d *db) pushToRemote(ctx context.Context, table string, oldMeta, meta *tableMeta) error {
	if meta.Type == "TABLE" {
		localPath := d.localTableDir(table, meta.Version)
		entries, err := os.ReadDir(localPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			d.logger.Debug("replicating file", slog.String("file", entry.Name()), slog.String("path", localPath))
			// no directory should exist as of now
			if entry.IsDir() {
				d.logger.Debug("found directory in path which should not exist", slog.String("file", entry.Name()), slog.String("path", localPath))
				continue
			}

			wr, err := os.Open(filepath.Join(localPath, entry.Name()))
			if err != nil {
				return err
			}

			// upload to cloud storage
			err = retry(ctx, func() error {
				return d.remote.Upload(ctx, path.Join(table, meta.Version, entry.Name()), wr, &blob.WriterOptions{
					ContentType: "application/octet-stream",
				})
			})
			_ = wr.Close()
			if err != nil {
				return err
			}
		}
	}

	// update table meta
	// todo :: also use etag to avoid concurrent writer conflicts
	d.localDirty = true
	m, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal table metadata: %w", err)
	}
	err = retry(ctx, func() error {
		return d.remote.WriteAll(ctx, path.Join(table, "meta.json"), m, nil)
	})
	if err != nil {
		d.logger.Error("failed to update meta.json in remote", slog.String("table", table), slog.Any("error", err))
	}

	// success -- remove old version
	if oldMeta != nil {
		_ = d.deleteRemote(ctx, table, oldMeta.Version)
	}
	return err
}

// deleteRemote deletes remote.
// If table is specified, only that table is deleted.
// If table and version is specified, only that version of the table is deleted.
func (d *db) deleteRemote(ctx context.Context, table, version string) error {
	if table == "" && version != "" {
		return fmt.Errorf("table must be specified if version is specified")
	}
	var prefix string
	if table != "" {
		if version != "" {
			prefix = path.Join(table, version) + "/"
		} else {
			prefix = table + "/"
			// delete meta.json first
			err := retry(ctx, func() error { return d.remote.Delete(ctx, "meta.json") })
			if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
				d.logger.Error("failed to delete meta.json in remote", slog.String("table", table), slog.Any("error", err))
				return err
			}
		}
	}
	// ignore errors since meta.json is already removed

	iter := d.remote.List(&blob.ListOptions{Prefix: prefix})
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			d.logger.Debug("failed to list object", slog.String("table", table), slog.Any("error", err))
		}
		err = retry(ctx, func() error { return d.remote.Delete(ctx, obj.Key) })
		if err != nil {
			d.logger.Debug("failed to delete object", slog.String("table", table), slog.String("object", obj.Key), slog.Any("error", err))
		}
	}
	return nil
}

func retry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < _maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // success
		}
		if !strings.Contains(err.Error(), "stream error: stream ID") {
			break // break and return error
		}

		timer := time.NewTimer(_retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err() // return on context cancellation
		case <-time.After(_retryDelay):
			timer.Stop()
		}
	}
	return err
}

const (
	_maxRetries = 5
	_retryDelay = 10 * time.Second
)
