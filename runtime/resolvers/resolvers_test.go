package resolvers

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Project struct {
	Sources    map[string]yaml.Node
	Models     map[string]yaml.Node
	Dashboards map[string]yaml.Node
	APIs       map[string]yaml.Node
}
type Resolvers struct {
	Project    Project
	Connectors map[string]*testruntime.InstanceOptionsForResolvers
	Tests      map[string]*Test
}

type Test struct {
	Options struct {
		InstanceID         string
		Resolver           string
		ResolverProperties map[string]any "yaml:\"resolver_properties\""
		Args               map[string]any
		Claims             struct {
			UserAttributes map[string]any "yaml:\"user_attributes\""
		}
	}
	Result        []map[string]any
	ErrorContains string "yaml:\"error_contains\""
}

var update = flag.Bool("update", false, "Update test results")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestResolvers(t *testing.T) {
	files, err := filepath.Glob("./testdata/*_resolvers_test.yaml")
	require.NoError(t, err)
	for _, f := range files {
		t.Log("Running with", f)
		yamlFile, err := os.ReadFile(f)
		require.NoError(t, err)
		var r Resolvers
		err = yaml.Unmarshal(yamlFile, &r)
		require.NoError(t, err)

		files := make(map[string]string)
		for name, node := range r.Project.Sources {
			abs := filepath.Join("sources", name)
			bytes, err := yaml.Marshal(&node)
			require.NoError(t, err)
			files[abs] = string(bytes)
		}
		for name, node := range r.Project.Models {
			abs := filepath.Join("models", name)
			var bytes []byte
			bytes, err = yaml.Marshal(&node)
			require.NoError(t, err)
			files[abs] = string(bytes)
		}
		for name, node := range r.Project.Dashboards {
			abs := filepath.Join("dashboards", name)
			bytes, err := yaml.Marshal(&node)
			require.NoError(t, err)
			files[abs] = string(bytes)
		}
		for name, node := range r.Project.APIs {
			abs := filepath.Join("apis", name)
			bytes, err := yaml.Marshal(&node)
			require.NoError(t, err)
			files[abs] = string(bytes)
		}

		for connector, opts := range r.Connectors {
			t.Log("Running with", connector)
			if opts == nil {
				opts = &testruntime.InstanceOptionsForResolvers{}
			}
			if opts.Files == nil {
				opts.Files = map[string]string{"rill.yaml": ""}
			}

			switch connector {
			case "druid":
				opts.OLAPDriver = "druid"
			case "clickhouse":
				opts.OLAPDriver = "clickhouse"
			}

			maps.Copy(opts.Files, files)
			rt, instanceID := testruntime.NewInstanceForResolvers(t, *opts)
			for testName, test := range r.Tests {
				t.Run(testName, func(t *testing.T) {
					t.Log("======================")
					t.Log("Running ", testName, "with", f, "and", connector)
					testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

					ropts := test.Options
					ro := &runtime.ResolveOptions{}
					ro.InstanceID = instanceID
					ro.Resolver = ropts.Resolver
					ro.ResolverProperties = ropts.ResolverProperties
					ro.Args = ropts.Args
					ro.Claims = &runtime.SecurityClaims{
						UserAttributes: ropts.Claims.UserAttributes,
					}
					res, err := rt.Resolve(context.Background(), ro)
					if test.ErrorContains != "" {
						if *update {
							// todo
						} else {
							require.ErrorContains(t, err, test.ErrorContains)
						}
						return
					} else {
						require.NoError(t, err)
					}
					var rows []map[string]interface{}
					b, err := res.MarshalJSON()
					require.NoError(t, err)
					require.NoError(t, json.Unmarshal(b, &rows), string(b))
					if *update {
						test.Result = rows
						for _, m := range test.Result {
							for k, v := range m {
								node := yaml.Node{}
								node.Kind = yaml.ScalarNode
								switch val := v.(type) {
								case float32:
									node.Value = strconv.FormatFloat(float64(val), 'f', 2, 32)
									m[k] = &node
								case float64:
									node.Value = strconv.FormatFloat(val, 'f', 2, 64)
									m[k] = &node
								}
							}
						}
					} else {
						require.Equal(t, test.Result, rows)
					}
					t.Log("======================")
				})
			}
			if *update {
				buf := bytes.Buffer{}
				yamlEncoder := yaml.NewEncoder(&buf)
				yamlEncoder.SetIndent(2)
				err := yamlEncoder.Encode(r)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(f, buf.Bytes(), 0644))
			}
		}
	}
}
