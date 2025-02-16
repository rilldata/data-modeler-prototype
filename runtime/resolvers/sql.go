package resolvers

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	runtime.RegisterResolverInitializer("sql", newSQL)
}

type sqlResolver struct {
	sql                 string
	refs                []*runtimev1.ResourceName
	olap                drivers.OLAPStore
	olapRelease         func()
	interactiveRowLimit int64
	priority            int
}

type sqlProps struct {
	Connector string `mapstructure:"connector"`
	SQL       string `mapstructure:"sql"`
	Limit     int64  `mapstructure:"limit"`
}

type sqlArgs struct {
	Priority int `mapstructure:"priority"`
	// NOTE: Not exhaustive. Any other args are passed to the "args" property when resolving the SQL template.
}

// newSQL creates a resolver that executes a SQL query.
// It supports the use of templating in the SQL string to inject user attributes and args into the SQL query.
func newSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &sqlProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	args := &sqlArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	inst, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	cfg, err := inst.Config()
	if err != nil {
		return nil, err
	}

	interactiveRowLimit := cfg.InteractiveSQLRowLimit
	if props.Limit != 0 {
		interactiveRowLimit = props.Limit
	}

	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID, props.Connector)
	if err != nil {
		return nil, err
	}

	resolvedSQL, refs, err := buildSQL(props.SQL, olap.Dialect(), opts.Args, inst, opts.Claims.UserAttributes, opts.ForExport)
	if err != nil {
		return nil, err
	}

	return &sqlResolver{
		sql:                 resolvedSQL,
		refs:                refs,
		olap:                olap,
		olapRelease:         release,
		interactiveRowLimit: interactiveRowLimit,
		priority:            args.Priority,
	}, nil
}

func (r *sqlResolver) Close() error {
	r.olapRelease()
	return nil
}

func (r *sqlResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	if r.olap.Dialect() == drivers.DialectDuckDB || r.olap.Dialect() == drivers.DialectClickHouse {
		return []byte(r.sql), len(r.refs) != 0, nil
	}
	return nil, false, nil
}

func (r *sqlResolver) Refs() []*runtimev1.ResourceName {
	return r.refs
}

func (r *sqlResolver) Validate(ctx context.Context) error {
	_, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:  r.sql,
		DryRun: true,
	})
	return err
}

func (r *sqlResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	// Wrap the SQL with an outer SELECT to limit the number of rows returned in interactive mode.
	// Adding +1 to the limit so we can return a nice error message if the limit is exceeded.
	sql := fmt.Sprintf("SELECT * FROM (%s) LIMIT %d", r.sql, r.interactiveRowLimit+1)

	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Priority: r.priority,
	})
	if err != nil {
		return nil, err
	}

	// This is a little hacky, but for now we only cache results from DuckDB queries that have refs.
	if r.interactiveRowLimit != 0 {
		res.SetCap(r.interactiveRowLimit)
	}
	return runtime.NewDriverResolverResult(res), nil
}

func (r *sqlResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	exportOpts := &runtime.ExportOptions{
		Format:       opts.Format,
		Priority:     r.priority,
		PreWriteHook: opts.PreWriteHook,
	}

	filename := "api_export_" + time.Now().Format("2006-01-02T15-04-05.000Z")

	switch r.olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			return queries.DuckDBCopyExport(ctx, w, exportOpts, r.sql, nil, filename, r.olap, opts.Format)
		}
		return r.generalExport(ctx, w, filename, exportOpts)
	case drivers.DialectDruid, drivers.DialectClickHouse:
		return r.generalExport(ctx, w, filename, exportOpts)
	default:
		return fmt.Errorf("export not available for dialect %q", r.olap.Dialect().String())
	}
}

func (r *sqlResolver) generalExport(ctx context.Context, w io.Writer, filename string, opts *runtime.ExportOptions) error {
	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    r.sql,
		Priority: opts.Priority,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	meta := make([]*runtimev1.MetricsViewColumn, len(res.Schema.Fields))
	for i, f := range res.Schema.Fields {
		meta[i] = &runtimev1.MetricsViewColumn{
			Name: f.Name,
			Type: f.Type.Code.String(),
		}
	}

	var data []*structpb.Struct
	for res.Rows.Next() {
		row := make(map[string]any)
		err = res.Rows.MapScan(row)
		if err != nil {
			return err
		}
		curr, err := structpb.NewStruct(row)
		if err != nil {
			return err
		}
		data = append(data, curr)
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return queries.WriteCSV(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return queries.WriteXLSX(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return queries.WriteParquet(meta, data, w)
	}

	return nil
}

// buildSQL resolves the SQL template and returns the resolved SQL and the resource names it references.
func buildSQL(sqlTemplate string, dialect drivers.Dialect, args map[string]any, inst *drivers.Instance, userAttributes map[string]any, forExport bool) (string, []*runtimev1.ResourceName, error) {
	// Resolve the SQL template
	sql, refs, err := resolveTemplate(sqlTemplate, args, inst, userAttributes, forExport)
	if err != nil {
		return "", nil, err
	}

	// For DuckDB, we can do ref inference using the SQL AST (similar to the rillv1 compiler).
	if dialect == drivers.DialectDuckDB {
		ast, err := duckdbsql.Parse(sql)
		if err != nil {
			return "", nil, err
		}
		for _, t := range ast.GetTableRefs() {
			if !t.LocalAlias && t.Name != "" && t.Function == "" && len(t.Paths) == 0 {
				// We don't know if it's a source or model (or neither), so we add both. Refs are just approximate.
				refs = append(refs,
					&runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: t.Name},
					&runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: t.Name},
				)
			}
		}
	}

	return sql, normalizeRefs(refs), nil
}

func resolveTemplate(sqlTemplate string, args map[string]any, inst *drivers.Instance, userAttributes map[string]any, forExport bool) (string, []*runtimev1.ResourceName, error) {
	var refs []*runtimev1.ResourceName
	sql, err := compilerv1.ResolveTemplate(sqlTemplate, compilerv1.TemplateData{
		Environment: inst.Environment,
		User:        userAttributes,
		Variables:   inst.ResolveVariables(false),
		ExtraProps: map[string]any{
			"args":   args,
			"export": forExport,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			// Add to the list of potential refs
			if ref.Kind == compilerv1.ResourceKindUnspecified {
				// We don't know if it's a source or model (or neither), so we add both. Refs are just approximate.
				refs = append(refs,
					&runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: ref.Name},
					&runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: ref.Name},
				)
			} else {
				refs = append(refs, runtime.ResourceNameFromCompiler(ref))
			}

			// Return the escaped identifier
			// TODO: As of now it is using `DialectDuckDB` in all cases since in certain cases like metrics_sql it is not possible to identify OLAP connector before template resolution.
			return drivers.DialectDuckDB.EscapeIdentifier(ref.Name), nil
		},
	})
	if err != nil {
		return "", nil, err
	}
	return sql, refs, nil
}
