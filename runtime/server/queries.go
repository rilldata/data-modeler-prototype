package server

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"
	"unicode/utf8"

	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/sql"
	"github.com/rilldata/rill/runtime/sql/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Query implements RuntimeService
func (s *Server) Query(ctx context.Context, req *api.QueryRequest) (*api.QueryResponse, error) {
	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	res, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    req.Sql,
		Args:     args,
		DryRun:   req.DryRun,
		Priority: int(req.Priority),
	})
	if err != nil {
		// TODO: Parse error to determine error code
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.DryRun {
		// TODO: Return a meta object for dry-run queries
		// NOTE: Currently, instance.Query return nil rows for succesful dry-run queries
		return &api.QueryResponse{}, nil
	}

	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.QueryResponse{
		Meta: res.Schema,
		Data: data,
	}

	return resp, nil
}

func (s *Server) query(ctx context.Context, instanceID string, stmt *drivers.Statement) (*drivers.Result, error) {
	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, instanceID)
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	// Connect to olap
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Open catalog
	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	catalogObjects := s.catalogCache.allObjects(ctx, instanceID, catalog)

	// Get target dialect
	var dialect rpc.Dialect
	switch inst.Driver {
	case "duckdb":
		dialect = rpc.Dialect_DUCKDB
	case "druid":
		dialect = rpc.Dialect_DRUID
	default:
		panic(fmt.Errorf("unexpected instance driver: %s", inst.Driver))
	}

	// Transpile query
	stmt.Query, err = sql.Transpile(stmt.Query, dialect, catalogObjects)
	if err != nil {
		return nil, err
	}

	// Run
	return olap.Execute(ctx, stmt)
}

// QueryDirect implements RuntimeService
func (s *Server) QueryDirect(ctx context.Context, req *api.QueryDirectRequest) (*api.QueryDirectResponse, error) {
	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	res, err := s.queryDirect(ctx, req.InstanceId, &drivers.Statement{
		Query:    req.Sql,
		Args:     args,
		DryRun:   req.DryRun,
		Priority: int(req.Priority),
	})
	if err != nil {
		// TODO: Parse error to determine error code
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.DryRun {
		// TODO: Return a meta object for dry-run queries
		// NOTE: Currently, instance.Query return nil rows for succesful dry-run queries
		return &api.QueryDirectResponse{}, nil
	}

	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.QueryDirectResponse{
		Meta: res.Schema,
		Data: data,
	}

	return resp, nil
}

func (s *Server) queryDirect(ctx context.Context, instanceID string, stmt *drivers.Statement) (*drivers.Result, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, instanceID)
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	return olap.Execute(ctx, stmt)
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := mapToPB(rowMap)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// valToPB converts any value to a google.protobuf.Value. It's similar to
// structpb.NewValue, but adds support for a few extra primitive types.
func valToPB(v any) (*structpb.Value, error) {
	switch v := v.(type) {
	// In addition to the extra supported types, we also override handling for
	// maps and lists since we need to use valToPB on nested fields.
	case map[string]interface{}:
		v2, err := mapToPB(v)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(v2), nil
	case []interface{}:
		v2, err := sliceToPB(v)
		if err != nil {
			return nil, err
		}
		return structpb.NewListValue(v2), nil
	// Handle types not handled by structpb.NewValue
	case int8:
		return structpb.NewNumberValue(float64(v)), nil
	case int16:
		return structpb.NewNumberValue(float64(v)), nil
	case uint8:
		return structpb.NewNumberValue(float64(v)), nil
	case uint16:
		return structpb.NewNumberValue(float64(v)), nil
	case time.Time:
		s := v.Format(time.RFC3339Nano)
		return structpb.NewStringValue(s), nil
	case float32:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return structpb.NewNullValue(), nil
		}
		return structpb.NewNumberValue(float64(v)), nil
	case float64:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return structpb.NewNullValue(), nil
		}
		return structpb.NewNumberValue(float64(v)), nil
	case *big.Int:
		// Evil cast to float until frontend can deal with bigs:
		v2, _ := new(big.Float).SetInt(v).Float64()
		return structpb.NewNumberValue(v2), nil
		// This is what we should do when frontend supports it:
		// s := v.String()
		// return structpb.NewStringValue(s), nil
	case duckdb.Interval:
		m := map[string]any{"months": v.Months, "days": v.Days, "micros": v.Micros}
		v2, err := mapToPB(m)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(v2), nil
	default:
		// Default handling for basic types (ints, string, etc.)
		return structpb.NewValue(v)
	}
}

// mapToPB converts a map to a google.protobuf.Struct. It's similar to
// structpb.NewStruct(), but it recurses on valToPB instead of structpb.NewValue
// to add support for more types.
func mapToPB(v map[string]any) (*structpb.Struct, error) {
	x := &structpb.Struct{Fields: make(map[string]*structpb.Value, len(v))}
	for k, v := range v {
		if !utf8.ValidString(k) {
			return nil, fmt.Errorf("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Fields[k], err = valToPB(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// sliceToPB converts a map to a google.protobuf.List. It's similar to
// structpb.NewList(), but it recurses on valToPB instead of structpb.NewList
// to add support for more types.
func sliceToPB(v []interface{}) (*structpb.ListValue, error) {
	x := &structpb.ListValue{Values: make([]*structpb.Value, len(v))}
	for i, v := range v {
		var err error
		x.Values[i], err = valToPB(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}
