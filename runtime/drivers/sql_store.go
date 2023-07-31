package drivers

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrIteratorDone = errors.New("empty iterator")

// SQLStore is implemented by drivers capable of running sql queries and generating an iterator to consume results.
// In future the results can be produced in other formats like arrow as well.
// May be call it DataWarehouse to differentiate from OLAP or postgres?
type SQLStore interface {
	Query(ctx context.Context, props map[string]any, sql string) (RowIterator, error)
}

// RowIterator returns an iterator to iterate over result of a sql query
type RowIterator interface {
	// Schema of the underlying data
	Schema(ctx context.Context) (*runtimev1.StructType, error)
	// Next fetches next row
	Next(ctx context.Context) ([]any, error)
	// Close closes the iterator and frees resources
	Close() error
	// Size returns total size of data downloaded in unit.
	// Returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (uint64, bool)
}
