package runtime

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// Resolver represents logic, such as a SQL query, that produces output data.
// Resolvers are used to evaluate API requests, alerts, reports, etc.
//
// A resolver has two levels of configuration: static properties and dynamic arguments.
// For example, a SQL resolver has a static property for the SQL query and dynamic arguments for the query parameters.
// The static properties are usually declared in advance, such as in the YAML for a custom API, whereas the dynamic arguments are provided just prior to execution, such as in an API request.
type Resolver interface {
	// Close is called when done with the resolver.
	// Note that the Resolve method may not have been called when Close is called (in case of cache hits or validation failures).
	Close() error
	// Key that can be used for caching. It can be a large string since the value will be hashed.
	// The key should include all the properties and args that affect the output.
	// It does not need to include the instance ID or resolver name, as those are added separately to the cache key.
	Key() string
	// Refs access by the resolver. The output may be approximate, i.e. some of the refs may not exist.
	// The output should avoid duplicates and be stable between invocations.
	Refs() []*runtimev1.ResourceName
	// Validate the properties and args without running any expensive operations.
	Validate(ctx context.Context) error
	// ResolveInteractive resolves data for interactive use (e.g. API requests or alerts).
	ResolveInteractive(ctx context.Context) (ResolverResult, error)
	// ResolveExport resolve data for export (e.g. downloads or reports).
	ResolveExport(ctx context.Context, w io.Writer, opts *ResolverExportOptions) error
}

// ResolverResult is the result of a resolver's execution.
type ResolverResult interface {
	// Schema is the schema for the Data
	Schema() *runtimev1.StructType
	// Cache indicates whether the result can be cached
	Cache() bool
	// MarshalJSON is a convenience method to serialize the result to JSON.
	MarshalJSON() ([]byte, error)
	// Close should be called to release resources
	Close() error
}

func NewResolverResult(result *drivers.Result, rowLimit int64, cache bool) ResolverResult {
	return &resolverResult{
		rows:     result,
		rowLimit: rowLimit,
		cache:    cache,
	}
}

type resolverResult struct {
	rows     *drivers.Result
	cache    bool
	rowLimit int64
}

// Cache implements ResolverResult.
func (r *resolverResult) Cache() bool {
	return r.cache
}

// Close implements ResolverResult.
func (r *resolverResult) Close() error {
	return r.rows.Close()
}

// MarshalJSON implements ResolverResult.
func (r *resolverResult) MarshalJSON() ([]byte, error) {
	// close is idempotent so we close rows in this function itself
	defer r.rows.Close()
	var out []map[string]any
	for r.rows.Next() {
		if int64(len(out)) >= r.rowLimit {
			return nil, fmt.Errorf("sql resolver: query limit exceeded: returned more than %d rows", r.rowLimit)
		}

		row := make(map[string]any)
		err := r.rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		for _, field := range r.rows.Schema.Fields {
			if row[field.Name] == nil {
				continue
			}
			switch field.Type.Code {
			case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
				// big.Int is marshalled as Number in JSON which can lead to loss of precision. We fix this by setting as string.
				switch v := row[field.Name].(type) {
				case big.Int:
					row[field.Name] = v.Text(10)
				case *big.Int:
					row[field.Name] = v.Text(10)
				}
			case runtimev1.Type_CODE_DATE:
				switch v := row[field.Name].(type) {
				case time.Time:
					row[field.Name] = v.Format(time.DateOnly)
				case *time.Time:
					row[field.Name] = v.Format(time.DateOnly)
				}
			case runtimev1.Type_CODE_UINT64:
				switch v := row[field.Name].(type) {
				case uint64:
					row[field.Name] = strconv.FormatUint(v, 10)
				case *uint64:
					row[field.Name] = strconv.FormatUint(*v, 10)
				}
			case runtimev1.Type_CODE_DECIMAL:
				if v, ok := row[field.Name].(duckdb.Decimal); ok {
					row[field.Name] = duckDBDecimalToString(v)
				}
			}
		}
		out = append(out, row)
	}
	return json.Marshal(out)
}

// Schema implements ResolverResult.
func (r *resolverResult) Schema() *runtimev1.StructType {
	return r.rows.Schema
}

var _ ResolverResult = &resolverResult{}

// ResolverExportOptions are the options passed to a resolver's ResolveExport method.
type ResolverExportOptions struct {
	// Format is the format to export the result in.
	Format runtimev1.ExportFormat
	// PreWriteHook is a function that is called after the export has been prepared, but before the first bytes are output to the io.Writer.
	PreWriteHook func(filename string) error
}

// ResolverOptions are the options passed to a resolver initializer.
type ResolverOptions struct {
	Runtime        *Runtime
	InstanceID     string
	Properties     map[string]any
	Args           map[string]any
	UserAttributes map[string]any
	ForExport      bool
}

// ResolverInitializer is a function that initializes a resolver.
type ResolverInitializer func(ctx context.Context, opts *ResolverOptions) (Resolver, error)

// ResolverInitializers tracks resolver initializers by name.
var ResolverInitializers = make(map[string]ResolverInitializer)

// RegisterResolverInitializer registers a resolver initializer by name.
func RegisterResolverInitializer(name string, initializer ResolverInitializer) {
	if ResolverInitializers[name] != nil {
		panic(fmt.Errorf("resolver already registered for name %q", name))
	}
	ResolverInitializers[name] = initializer
}

// ResolveOptions are the options passed to the runtime's Resolve method.
type ResolveOptions struct {
	InstanceID         string
	Resolver           string
	ResolverProperties map[string]any
	Args               map[string]any
	UserAttributes     map[string]any
}

// ResolveResult is subset of ResolverResult that is cached
type ResolveResult struct {
	Data   []byte
	Schema *runtimev1.StructType
}

// Resolve resolves a query using the given options.
func (r *Runtime) Resolve(ctx context.Context, opts *ResolveOptions) (ResolveResult, error) {
	// Initialize the resolver
	initializer, ok := ResolverInitializers[opts.Resolver]
	if !ok {
		return ResolveResult{}, fmt.Errorf("no resolver found for name %q", opts.Resolver)
	}
	resolver, err := initializer(ctx, &ResolverOptions{
		Runtime:        r,
		InstanceID:     opts.InstanceID,
		Properties:     opts.ResolverProperties,
		Args:           opts.Args,
		UserAttributes: opts.UserAttributes,
		ForExport:      false,
	})
	if err != nil {
		return ResolveResult{}, err
	}
	defer resolver.Close()

	// Build cache key based on the resolver's key and refs
	ctrl, err := r.Controller(ctx, opts.InstanceID)
	if err != nil {
		return ResolveResult{}, err
	}
	hash := md5.New()
	if _, err := hash.Write([]byte(resolver.Key())); err != nil {
		return ResolveResult{}, err
	}
	for _, ref := range resolver.Refs() {
		res, err := ctrl.Get(ctx, ref, false)
		if err != nil {
			// Refs are approximate, not exact (see docstring for Refs()), so they may not all exist
			continue
		}

		if _, err := hash.Write([]byte(res.Meta.Name.Kind)); err != nil {
			return ResolveResult{}, err
		}
		if _, err := hash.Write([]byte(res.Meta.Name.Name)); err != nil {
			return ResolveResult{}, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Seconds); err != nil {
			return ResolveResult{}, err
		}
		if err := binary.Write(hash, binary.BigEndian, res.Meta.StateUpdatedOn.Nanos); err != nil {
			return ResolveResult{}, err
		}
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	key := fmt.Sprintf("inst:%s:resolver:%s:hash:%s", opts.InstanceID, opts.Resolver, sum)

	// Try to get from cache
	if val, ok := r.queryCache.cache.Get(key); ok {
		return val.(ResolveResult), nil
	}

	// Load with singleflight
	val, err := r.queryCache.singleflight.Do(ctx, key, func(ctx context.Context) (any, error) {
		// Try cache again
		if val, ok := r.queryCache.cache.Get(key); ok {
			return val, nil
		}

		res, err := resolver.ResolveInteractive(ctx)
		if err != nil {
			return ResolveResult{}, err
		}
		defer res.Close()

		data, err := res.MarshalJSON()
		if err != nil {
			return ResolveResult{}, err
		}

		cRes := ResolveResult{
			Data:   data,
			Schema: res.Schema(),
		}
		if res.Cache() {
			r.queryCache.cache.Set(key, cRes, int64(len(data)))
		}
		return cRes, nil
	})
	if err != nil {
		return ResolveResult{}, err
	}
	return val.(ResolveResult), nil
}

func duckDBDecimalToString(d duckdb.Decimal) string {
	scale := big.NewInt(int64(d.Scale))
	factor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), scale, nil))
	value := new(big.Float).SetInt(d.Value)
	value.Quo(value, factor)
	return value.Text('f', int(d.Scale))
}
