package resolvers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestFileYAML is the structure of a test file.
// One runtime instance will created for each test file, with the provided file contents and connectors initialized.
// Each test in the file will be run in sequence against that runtime instance.
// The available connectors are defined in runtime/testruntime/connectors.go.
type TestFileYAML struct {
	Connectors   []string             `yaml:"connectors,omitempty"`
	Variables    map[string]string    `yaml:"variables,omitempty"`
	DataFiles    map[string]string    `yaml:"data_files,omitempty"`
	ProjectFiles map[string]yaml.Node `yaml:"project_files"`
	Tests        []*TestYAML          `yaml:"tests"`
}

// TestYAML is the structure of a single resolver test executed against the runtime instance defined in TestFileYAML.
type TestYAML struct {
	Name               string           `yaml:"name"`
	Resolver           string           `yaml:"resolver"`
	Properties         yaml.Node        `yaml:"properties,omitempty"`      // Expects map[string]any, but using yaml.Node to preserve order for -update.
	Args               yaml.Node        `yaml:"args,omitempty"`            // Expects map[string]any, but using yaml.Node to preserve order for -update.
	UserAttributes     yaml.Node        `yaml:"user_attributes,omitempty"` // Expects map[string]any, but using yaml.Node to preserve order for -update.
	SkipSecurityChecks bool             `yaml:"skip_security_checks,omitempty"`
	Result             []map[string]any `yaml:"result,omitempty"`
	ResultCSV          string           `yaml:"result_csv,omitempty"`
	ErrorContains      string           `yaml:"error_contains,omitempty"`
}

// update is a flag that denotes whether to overwrite the results in the test files instead of checking them.
var update = flag.Bool("update", false, "Update test results")

// TestResolvers loads the test YAML files in ./testdata and runs them. Each test file should match the schema of TestFileYAML.
//
// The test YAML files provide a compact format for initializing runtime instances for a set of project files and connectors,
// and running a series of resolvers against them and testing the results.
//
// Example: run all resolver tests:
//
//	go test -run ^TestResolvers$ ./runtime/resolvers
//
// Example: update all resolver tests:
//
//	go test -run ^TestResolvers$ ./runtime/resolvers -update
//
// Example: run a single resolver test file:
//
// go test -run ^TestResolvers/metrics_clickhouse$ ./runtime/resolvers
//
// Example: run a single resolver file test case:
//
// go test -run ^TestResolvers/metrics_clickhouse/simple$ ./runtime/resolvers
func TestResolvers(t *testing.T) {
	// Evaluate the -update flag.
	flag.Parse()
	update := update != nil && *update

	// Discover the test files.
	files, err := filepath.Glob("./testdata/*.yaml")
	require.NoError(t, err)

	// Run each test file as a subtest.
	for _, f := range files {
		t.Run(fileutil.Stem(f), func(t *testing.T) {
			// Load the test file.
			data, err := os.ReadFile(f)
			require.NoError(t, err)
			var tf TestFileYAML
			err = yaml.Unmarshal(data, &tf)
			require.NoError(t, err)

			// Create a map of project files for the runtime instance.
			projectFiles := make(map[string]string)
			projectFiles["rill.yaml"] = ""
			for name, node := range tf.ProjectFiles {
				bytes, err := yaml.Marshal(&node)
				require.NoError(t, err)
				projectFiles[name] = string(bytes)
			}

			// Add local data files to project files
			for name, data := range tf.DataFiles {
				bytes, err := os.ReadFile(data)
				require.NoError(t, err)
				projectFiles[name] = string(bytes)
			}

			// Acquire the connectors for the runtime instance.
			vars := make(map[string]string)
			for k, v := range tf.Variables {
				vars[k] = v
			}
			for _, connector := range tf.Connectors {
				acquire, ok := testruntime.Connectors[connector]
				require.True(t, ok, "unknown connector %q", connector)
				cfg := acquire(t)
				for k, v := range cfg {
					k = fmt.Sprintf("connector.%s.%s", connector, k)
					vars[k] = v
				}
			}

			// Create the test runtime instance.
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files:     projectFiles,
				Variables: vars,
			})
			testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

			// Run each test case against the test runtime instance as a subtest.
			for _, tc := range tf.Tests {
				t.Run(tc.Name, func(t *testing.T) {
					// Read mapping properties that were parsed as yaml.Node to avoid reshuffling the order when using -update.
					properties := make(map[string]any)
					err := tc.Properties.Decode(&properties)
					require.NoError(t, err, "failed to decode properties into map[string]any")

					args := make(map[string]any)
					err = tc.Args.Decode(&args)
					require.NoError(t, err, "failed to decode args into map[string]any")

					userAttributes := make(map[string]any)
					err = tc.UserAttributes.Decode(&userAttributes)
					require.NoError(t, err, "failed to decode user_attributes into map[string]any")

					// Run the resolver.
					ctx := context.Background()
					res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
						InstanceID:         instanceID,
						Resolver:           tc.Resolver,
						ResolverProperties: properties,
						Args:               args,
						Claims: &runtime.SecurityClaims{
							UserAttributes: userAttributes,
							SkipChecks:     tc.SkipSecurityChecks,
						},
					})

					// If it succeeded, get the result rows.
					// Does a JSON roundtrip to coerce to simple types (easier to compare).
					var rows []map[string]any
					if err == nil {
						data, err2 := res.MarshalJSON()
						if err2 != nil {
							err = err2
						} else {
							err = json.Unmarshal(data, &rows)
						}
					}

					// If the -update flag is set, update the test case results instead of checking them.
					// The updated test case will be written back to the test file later.
					if update {
						if tc.ResultCSV != "" {
							tc.Result = nil
							tc.ResultCSV = resultToCSV(t, rows, res.Schema())
						} else {
							tc.Result = rows
						}

						tc.ErrorContains = ""
						if err != nil {
							tc.ErrorContains = err.Error()
						}
						return
					}

					// Check if an error was expected.
					if tc.ErrorContains != "" {
						require.Error(t, err)
						require.Contains(t, err.Error(), tc.ErrorContains)
						return
					}
					require.NoError(t, err)

					// We support expressing the expected result as a CSV string, which is more compact.
					// Serialize the result to CSV and compare.
					if tc.ResultCSV != "" {
						actual := resultToCSV(t, rows, res.Schema())
						require.Equal(t, strings.TrimSpace(tc.ResultCSV), strings.TrimSpace(actual))
						return
					}

					// Compare the result rows to the expected result.
					// Like for rows, we do a JSON roundtrip on the expected result (parsed from YAML) to coerce to simple types.
					var expected []map[string]any
					data, err := json.Marshal(tc.Result)
					require.NoError(t, err)
					err = json.Unmarshal(data, &expected)
					require.NoError(t, err)
					if len(expected) != 0 || len(rows) != 0 {
						require.EqualValues(t, expected, rows)
					}
				})
			}

			// If the -update flag is set, the TestYAML values have been updated with the output results.
			// Write out the updated test file.
			if update {
				buf := &bytes.Buffer{}
				yamlEncoder := yaml.NewEncoder(buf)
				yamlEncoder.SetIndent(2)
				err := yamlEncoder.Encode(tf)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(f, buf.Bytes(), 0644))
			}
		})
	}

}

// resultToCSV serializes the rows to a CSV formatted string.
// It is derived from runtime/drivers/file/model_executor_olap_self.go#writeCSV.
func resultToCSV(t *testing.T, rows []map[string]any, schema *runtimev1.StructType) string {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	strs := make([]string, len(schema.Fields))
	for i, f := range schema.Fields {
		strs[i] = f.Name
	}
	err := w.Write(strs)
	require.NoError(t, err)

	for _, row := range rows {
		for i, f := range schema.Fields {
			v, ok := row[f.Name]
			require.True(t, ok, "missing field %q", f.Name)

			var s string
			if v != nil {
				if v2, ok := v.(string); ok {
					s = v2
				} else {
					tmp, err := json.Marshal(v)
					require.NoError(t, err)
					s = string(tmp)
				}
			}

			strs[i] = s
		}

		err = w.Write(strs)
		require.NoError(t, err)
	}

	w.Flush()
	return buf.String()
}
