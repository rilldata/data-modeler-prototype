package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const _concurrentBlobDownloadLimit = 8

// map of supoprted extensions for partial downloads vs readers
// zipped csv files can't be partialled downloaded
// parquet files with compression has extension in format .<compression>.parquet eg: .gz.parquet
var _partialDownloadReaders = map[string]string{
	".parquet": "parquet",
	".csv":     "csv",
	".tsv":     "csv",
	".txt":     "csv",
	".ndjson":  "json",
	".json":    "json",
}

// implements connector.FileIterator
type blobIterator struct {
	ctx        context.Context
	bucket     *blob.Bucket
	objects    []*objectWithPlan
	index      int
	localFiles []string
	// all localfiles are created in this dir
	tempDir          string
	opts             *Options
	downloadMeasures *connectors.DownloadMeasures
}

type Options struct {
	GlobMaxTotalSize      int64
	GlobMaxObjectsMatched int
	GlobMaxObjectsListed  int64
	GlobPageSize          int
	ExtractPolicy         *runtimev1.Source_ExtractPolicy
	GlobPattern           string
}

// sets defaults if not set by user
func (opts *Options) validate() {
	if opts.GlobMaxObjectsMatched == 0 {
		opts.GlobMaxObjectsMatched = 1000
	}
	if opts.GlobMaxObjectsListed == 0 {
		opts.GlobMaxObjectsListed = 1000 * 1000
	}
	if opts.GlobMaxTotalSize == 0 {
		// 10 GB
		opts.GlobMaxTotalSize = 10 * 1024 * 1024 * 1024
	}
	if opts.GlobPageSize == 0 {
		opts.GlobPageSize = 1000
	}
}

func (opts *Options) validateLimits(size int64, matchCount int, fetched int64) error {
	if size > opts.GlobMaxTotalSize {
		return fmt.Errorf("glob pattern exceeds limits: would fetch more than %d bytes", opts.GlobMaxTotalSize)
	}
	if matchCount > opts.GlobMaxObjectsMatched {
		return fmt.Errorf("glob pattern exceeds limits: matched more than %d files", opts.GlobMaxObjectsMatched)
	}
	if fetched > opts.GlobMaxObjectsListed {
		return fmt.Errorf("glob pattern exceeds limits: listed more than %d files", opts.GlobMaxObjectsListed)
	}
	return nil
}

// NewIterator returns new instance of blobIterator
// the iterator keeps list of blob objects eagerly planned as per user's glob pattern and extract policies
// clients should call close once done to release all resources like closing the bucket
// the iterator takes responsibility of closing the bucket
func NewIterator(ctx context.Context, bucket *blob.Bucket, opts Options) (connectors.FileIterator, error) {
	opts.validate()

	measures, err := connectors.InitDownloadMeasures()
	if err != nil {
		return nil, err
	}

	it := &blobIterator{
		ctx:              ctx,
		bucket:           bucket,
		opts:             &opts,
		downloadMeasures: measures,
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "blob_ingestion")
	if err != nil {
		it.Close()
		return nil, err
	}
	it.tempDir = tempDir

	objects, err := it.plan()
	if err != nil {
		it.Close()
		return nil, err
	}
	it.objects = objects

	return it, nil
}

// Close frees the resources
func (it *blobIterator) Close() error {
	// remove temp dir since recursive paths created have to be removed as well
	err := os.RemoveAll(it.tempDir)
	if bucketCloseErr := it.bucket.Close(); bucketCloseErr != nil {
		return bucketCloseErr
	}
	return err
}

// HasNext returns true if iterator has more data
func (it *blobIterator) HasNext() bool {
	return it.index < len(it.objects)
}

// NextBatch downloads next n files and copies to local directory
func (it *blobIterator) NextBatch(n int) ([]string, error) {
	if !it.HasNext() {
		return nil, io.EOF
	}

	// delete files created in last iteration
	fileutil.ForceRemoveFiles(it.localFiles)
	start := it.index
	end := it.index + n
	if end > len(it.objects) {
		end = len(it.objects)
	}
	it.index = end

	// new slice creation is not necessary on every iteration
	// but there may be cases where n in first batch is different from n in next batch
	// to keep things easy creating a new slice every time
	it.localFiles = make([]string, end-start)
	g, grpCtx := errgroup.WithContext(it.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)
	for i, obj := range it.objects[start:end] {
		obj := obj
		index := start + i // with repect to object slice
		g.Go(func() error {
			// need to create file by maintaining same dir path as in glob for hivepartition support
			filename := filepath.Join(it.tempDir, obj.obj.Key)
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return err
			}

			file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer file.Close()

			it.localFiles[index-start] = file.Name()
			ext := filepath.Ext(obj.obj.Key)
			partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]

			startTime := time.Now()
			defer func(t *time.Time) {
				duration := time.Since(startTime)
				it.downloadMeasures.Collect(grpCtx, obj.obj.Size, duration, ext, "blob", isPartialDownloadSupported)
			}(&startTime)

			if obj.full || !isPartialDownloadSupported {
				// download full file
				return downloadObject(grpCtx, it.bucket, obj.obj.Key, file)
			}
			// download partial file
			// check if, for smaller size we can download entire file
			switch partialReader {
			case "parquet":
				return downloadParquet(grpCtx, it.bucket, obj.obj, obj.extractOption, file)
			case "csv":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: true}, file)
			case "json":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: false}, file)
			default:
				// should not reach here
				panic(fmt.Errorf("partial download not supported for extension %q", ext))
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result := make([]string, end-start)
	// clients can make changes to slice if passing the same slice that iterator holds
	// creating a copy since we want to delete all these files on next batch/close
	copy(result, it.localFiles)
	return result, nil
}

// todo :: ideally planner should take ownership of the bucket and return an iterator with next returning objectWithPlan
func (it *blobIterator) plan() ([]*objectWithPlan, error) {
	var (
		size, fetched int64
		matchCount    int
	)
	planner, err := newPlanner(it.opts.ExtractPolicy)
	if err != nil {
		return nil, err
	}

	listOpts := listOptions(it.opts.GlobPattern)
	token := blob.FirstPageToken
	for token != nil && !planner.done() {
		objs, nextToken, err := it.bucket.ListPage(it.ctx, token, it.opts.GlobPageSize, listOpts)
		if err != nil {
			return nil, err
		}

		token = nextToken
		fetched += int64(len(objs))
		for _, obj := range objs {
			if matched, _ := doublestar.Match(it.opts.GlobPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				if !planner.add(obj) {
					break
				}
			}
		}
		if err := it.opts.validateLimits(size, matchCount, fetched); err != nil {
			return nil, err
		}
	}

	items := planner.items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", it.opts.GlobPattern)
	}
	return items, nil
}

// listOptions for page listing api
func listOptions(globPattern string) *blob.ListOptions {
	listOptions := &blob.ListOptions{BeforeList: func(as func(interface{}) bool) error {
		// Access storage.Query via q here.
		var q *storage.Query
		if as(&q) {
			// we only need name and size, adding only required attributes to reduce data fetched
			_ = q.SetAttrSelection([]string{"Name", "Size"})
		}
		return nil
	}}

	prefix, glob := doublestar.SplitPattern(globPattern)
	if !fileutil.IsGlob(glob) {
		// single file
		listOptions.Prefix = globPattern
	} else if prefix != "." {
		listOptions.Prefix = prefix
	}

	return listOptions
}

// download full object
func downloadObject(ctx context.Context, bucket *blob.Bucket, objpath string, file *os.File) error {
	rc, err := bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()

	_, err = io.Copy(file, rc)
	return err
}
