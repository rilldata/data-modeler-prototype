package server

import (
	"context"
	"fmt"
	"regexp"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

// Table level profiling APIs
func (s *Server) RenameDatabaseObject(ctx context.Context, req *api.RenameDatabaseObjectRequest) (*api.RenameDatabaseObjectResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf("alter %s \"%s\" rename to \"%s\"", req.Type.String(), req.Name, req.Newname),
	})
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &api.RenameDatabaseObjectResponse{}, nil
}

func (s *Server) TableCardinality(ctx context.Context, req *api.CardinalityRequest) (*api.CardinalityResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(req.TableName),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	return &api.CardinalityResponse{
		Cardinality: count,
	}, nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

var DoubleQuotesRegexp *regexp.Regexp = regexp.MustCompile("\"")

func EscapeDoubleQuotes(column string) string {
	return DoubleQuotesRegexp.ReplaceAllString(column, "\"\"")
}

func (s *Server) ProfileColumns(ctx context.Context, req *api.ProfileColumnsRequest) (*api.ProfileColumnsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf(`select column_name as name, data_type as type from information_schema.columns 
		where table_name = '%s' and table_schema = current_schema()`, req.TableName),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pcs []*api.ProfileColumn
	i := 0
	for rows.Next() {
		pc := api.ProfileColumn{}
		if err := rows.StructScan(&pc); err != nil {
			return nil, err
		}
		pcs = append(pcs, &pc)
		i++
	}

	for _, pc := range pcs[0:i] {
		columnName := EscapeDoubleQuotes(pc.Name)
		rows, err = s.query(ctx, req.InstanceId, &drivers.Statement{
			Query: fmt.Sprintf(`select max(length("%s")) as max from %s`, columnName, req.TableName),
		})
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			if err := rows.Scan(&pc.LargestStringLength); err != nil {
				return nil, err
			}
		}
		rows.Close()
	}

	return &api.ProfileColumnsResponse{
		ProfileColumns: pcs[0:i],
	}, nil
}

func (s *Server) TableRows(ctx context.Context, req *api.RowsRequest) (*api.RowsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf("select * from %s limit %d", req.TableName, req.Limit),
	})
	if err != nil {
		return nil, err
	}
	var data []*structpb.Struct
	if data, err = rowsToData(rows); err != nil {
		return nil, err
	}

	return &api.RowsResponse{
		Data: data,
	}, nil
}
