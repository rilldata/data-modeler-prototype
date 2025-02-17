package query

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

// Command Long Description
var long = `Query a resolver within a project.

You can query a resolver by providing a SQL query, a resolver name, or a connector name.

Example Usage:

Query a resolver by providing a SQL query:
rill query my-project --sql "SELECT * FROM my-table"
rill query --sql "SELECT * FROM my-table" --limit 10
`

func QueryCmd(ch *cmdutil.Helper) *cobra.Command {
	var sql, connector, resolver, project, path string
	var properties, args map[string]string
	var limit int
	var local bool

	queryCmd := &cobra.Command{
		Use:   "query [<project>]",
		Short: "Query a resolver within a project",
		Long:  long,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			// Validate all inputs
			if err := validateQueryFlags(resolver, sql, properties, args); err != nil {
				return err
			}

			// Default resolver to "sql" if not provided
			if resolver == "" {
				resolver = "sql"
			}

			// Initialize maps if nil
			if properties == nil {
				properties = make(map[string]string)
			}
			if args == nil {
				args = make(map[string]string)
			}

			// Determine project name
			if len(cmdArgs) > 0 {
				project = cmdArgs[0]
			}
			if !local && !cmd.Flags().Changed("project") && len(cmdArgs) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Build resolver properties
			if sql != "" {
				properties["sql"] = sql
			}
			if connector != "" {
				properties["connector"] = connector
			}

			// Convert string maps to interface{} maps
			propsMap := make(map[string]any, len(properties))
			for k, v := range properties {
				propsMap[k] = v
			}
			argsMap := make(map[string]any, len(args))
			for k, v := range args {
				argsMap[k] = v
			}

			// Build the properties and args structs
			resolverProperties, resolverArgs, err := buildStructs(propsMap, argsMap)
			if err != nil {
				return fmt.Errorf("failed to build resolver properties and args: %w", err)
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}
			defer rt.Close()

			// Execute the query
			res, err := rt.RuntimeServiceClient.QueryResolver(cmd.Context(), &runtimev1.QueryResolverRequest{
				InstanceId:         instanceID,
				Resolver:           resolver,
				Limit:              int32(limit),
				ResolverProperties: resolverProperties,
				ResolverArgs:       resolverArgs,
			})
			if err != nil {
				return fmt.Errorf("failed to execute query: %w", err)
			}

			ch.PrintQueryResponse(res.Data)
			return nil
		},
	}

	// Project flags
	queryCmd.Flags().StringVar(&project, "project", "", "Project name")
	queryCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	queryCmd.Flags().BoolVar(&local, "local", false, "Target local runtime instead of Rill Cloud")

	// Query flags
	queryCmd.Flags().StringVar(&sql, "sql", "", "A SELECT query to execute")
	queryCmd.Flags().StringVar(&connector, "connector", "", "Connector to execute against. Defaults to the OLAP connector.")
	queryCmd.Flags().StringVar(&resolver, "resolver", "", "Explicit resolver (cannot be combined with --sql)")
	queryCmd.Flags().StringToStringVar(&properties, "properties", nil, "Explicit resolver properties (only with --resolver)")
	queryCmd.Flags().StringToStringVar(&args, "args", nil, "Explicit resolver args (only with --resolver)")
	queryCmd.Flags().IntVar(&limit, "limit", 100, "The maximum number of rows to print")

	return queryCmd
}

func validateQueryFlags(resolver, sql string, properties, args map[string]string) error {
	if resolver != "" && sql != "" {
		return fmt.Errorf("cannot combine --resolver and --sql")
	}

	if (len(args) > 0 || len(properties) > 0) && resolver == "" {
		return fmt.Errorf("must provide --resolver when using --args or --properties")
	}

	return nil
}

func buildStruct(m map[string]any) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}

	return structpb.NewStruct(m)
}

// buildStructs converts the properties and args maps into protobuf Struct types
func buildStructs(properties, args map[string]any) (*structpb.Struct, *structpb.Struct, error) {
	propertiesStruct, err := buildStruct(properties)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build properties struct: %w", err)
	}

	argsStruct, err := buildStruct(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build args struct: %w", err)
	}

	return propertiesStruct, argsStruct, nil
}
