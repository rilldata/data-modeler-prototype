package github

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/runtime/drivers"
)

var limit = 500

// Driver implements drivers.RepoStore.
func (c *connection) Driver() string {
	return "github"
}

// Root implements drivers.RepoStore.
func (c *connection) Root() string {
	return c.projectdir
}

// CommitHash implements drivers.RepoStore.
func (c *connection) CommitHash(ctx context.Context) (string, error) {
	err := c.cloneOrPull(ctx, true)
	if err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(c.tempdir)
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	if ref.Hash().IsZero() {
		return "", nil
	}

	return ref.Hash().String(), nil
}

// ListRecursive implements drivers.RepoStore.
func (c *connection) ListRecursive(ctx context.Context, glob string) ([]string, error) {
	err := c.cloneOrPull(ctx, true)
	if err != nil {
		return nil, err
	}

	fsRoot := os.DirFS(c.projectdir)
	glob = path.Clean(path.Join("./", glob))

	var paths []string
	err = doublestar.GlobWalk(fsRoot, glob, func(p string, d fs.DirEntry) error {
		// Don't track directories
		if d.IsDir() {
			return nil
		}

		// Exit if we reached the limit
		if len(paths) == limit {
			return fmt.Errorf("glob exceeded limit of %d matched files", limit)
		}

		// Track file (p is already relative to the FS root)
		p = filepath.Join("/", p)
		paths = append(paths, p)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// Get implements drivers.RepoStore.
func (c *connection) Get(ctx context.Context, filePath string) (string, error) {
	err := c.cloneOrPull(ctx, true)
	if err != nil {
		return "", err
	}

	filePath = filepath.Join(c.projectdir, filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Stat implements drivers.RepoStore.
func (c *connection) Stat(ctx context.Context, filePath string) (*drivers.RepoObjectStat, error) {
	err := c.cloneOrPull(ctx, true)
	if err != nil {
		return nil, err
	}

	filePath = filepath.Join(c.projectdir, filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &drivers.RepoObjectStat{
		LastUpdated: info.ModTime(),
	}, nil
}

// Put implements drivers.RepoStore.
func (c *connection) Put(ctx context.Context, filePath string, reader io.Reader) error {
	return fmt.Errorf("Put operation is unsupported")
}

// Rename implements drivers.RepoStore.
func (c *connection) Rename(ctx context.Context, fromPath, toPath string) error {
	return fmt.Errorf("Rename operation is unsupported")
}

// Delete implements drivers.RepoStore.
func (c *connection) Delete(ctx context.Context, filePath string) error {
	return fmt.Errorf("Delete operation is unsupported")
}

// Sync implements drivers.RepoStore.
func (c *connection) Sync(ctx context.Context) error {
	return c.cloneOrPull(ctx, false)
}

func (c *connection) Watch(ctx context.Context, callback drivers.WatchCallback) error {
	return fmt.Errorf("cannot watch %s repository is not supported", c.Driver())
}
