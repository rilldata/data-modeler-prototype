package cmdutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var ignoreFileList = []string{"/.env"}

// UploadRepo uploads a local project files to rill managed store.
// Internally it creates an asset object on admin service and returns its id which can be supplied while creating/updating project.
func UploadRepo(ctx context.Context, repo drivers.RepoStore, ch *Helper, org, name string) (string, error) {
	// list files
	entries, err := repo.ListRecursive(ctx, "**", false)
	if err != nil {
		return "", err
	}

	// generate a tar ball
	b := &bytes.Buffer{}
	if err := createTarball(b, entries, repo.Root()); err != nil {
		return "", err
	}

	// upload the tar ball
	assetID, err := uploadTarBall(ctx, ch, org, name, b)
	if err != nil {
		return "", err
	}
	return assetID, nil
}

// borrowed from https://github.com/goreleaser/goreleaser/blob/main/pkg/archive/tar/tar.go with minor changes
func createTarball(writer io.Writer, files []drivers.DirEntry, root string) error {
	gw, err := gzip.NewWriterLevel(writer, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, entry := range files {
		if drivers.IsIgnored(entry.Path, ignoreFileList) {
			continue
		}
		fullPath := filepath.Join(root, entry.Path)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: repo contains symlinks", entry.Path)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		header.Name = entry.Path
		if err = tw.WriteHeader(header); err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if info.IsDir() {
			continue
		}

		file, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			file.Close()
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		file.Close()
	}
	return nil
}

func uploadTarBall(ctx context.Context, ch *Helper, org, name string, body io.Reader) (string, error) {
	adminClient, err := ch.Client()
	if err != nil {
		return "", err
	}

	// generate a upload URL
	asset, err := adminClient.CreateAsset(ctx, &adminv1.CreateAssetRequest{
		OrganizationName: org,
		Type:             "deploy",
		Name:             fmt.Sprintf("%s__%s", org, name),
		Extension:        "tar.gz",
	})
	if err != nil {
		return "", err
	}

	// Create a put request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, asset.SignedUrl, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range asset.SigningHeaders {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload file: status code %d, response %s", resp.StatusCode, string(body))
	}
	return asset.AssetId, nil
}
