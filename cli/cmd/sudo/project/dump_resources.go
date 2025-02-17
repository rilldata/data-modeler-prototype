package project

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func DumpResources(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var annotations map[string]string
	var typ string

	searchCmd := &cobra.Command{
		Use:   "dump-resources [<project-pattern>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Dump resources for projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var pattern string
			// If args is not empty, use the first element as the pattern
			if len(args) > 0 {
				pattern = args[0]
			} else {
				pattern = "%"
			}

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: pattern,
				Annotations: annotations,
				PageSize:    pageSize,
				PageToken:   pageToken,
			})
			if err != nil {
				return err
			}

			if len(res.Names) == 0 {
				ch.PrintfWarn("No projects found\n")
				return nil
			}

			var m sync.Mutex
			failedProjects := map[string]error{}
			resources := map[string]map[string][]*runtimev1.Resource{}
			grp, ctx := errgroup.WithContext(ctx)
			for _, name := range res.Names {
				org := strings.Split(name, "/")[0]
				project := strings.Split(name, "/")[1]

				grp.Go(func() error {
					row, err := resourcesForProject(ctx, client, org, project, typ)
					if err != nil {
						m.Lock()
						failedProjects[name] = err
						m.Unlock()
						return nil
					}
					m.Lock()
					projects, ok := resources[org]
					if !ok {
						projects = map[string][]*runtimev1.Resource{}
						resources[org] = projects
					}
					projects[project] = row
					m.Unlock()
					return nil
				})
			}

			err = grp.Wait()
			if err != nil {
				return err
			}

			printResources(ch.Printer, resources)

			for name, err := range failedProjects {
				ch.Println()
				ch.PrintfWarn("Failed to dump resources for project %v: %s\n", name, err)
			}
			if res.NextPageToken != "" {
				ch.Println()
				ch.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().StringVar(&typ, "type", "", "Filter for resources of a specific type")
	searchCmd.Flags().StringToStringVar(&annotations, "annotation", nil, "Annotations to filter projects by (supports wildcard values)")
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 1000, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}

func resourcesForProject(ctx context.Context, c *client.Client, org, project, filter string) ([]*runtimev1.Resource, error) {
	proj, err := c.GetProject(ctx, &adminv1.GetProjectRequest{
		OrganizationName:    org,
		Name:                project,
		IssueSuperuserToken: true,
	})
	if err != nil {
		return nil, err
	}

	depl := proj.ProdDeployment
	if depl == nil {
		return nil, nil
	}

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime: %w", err)
	}

	req := &runtimev1.ListResourcesRequest{
		InstanceId: depl.RuntimeInstanceId,
	}
	if filter != "" {
		req.Kind = parseResourceKind(filter)
	}
	res, err := rt.ListResources(ctx, req)
	if err != nil {
		msg := err.Error()
		if s, ok := status.FromError(err); ok {
			msg = s.Message()
		}
		return nil, fmt.Errorf("runtime error, failed to list resources: %v", msg)
	}

	return res.Resources, nil
}

func parseResourceKind(k string) string {
	switch strings.ToLower(strings.TrimSpace(k)) {
	case "source":
		return runtime.ResourceKindSource
	case "model":
		return runtime.ResourceKindModel
	case "metricsview", "metrics_view":
		return runtime.ResourceKindMetricsView
	case "explore":
		return runtime.ResourceKindExplore
	case "migration":
		return runtime.ResourceKindMigration
	case "report":
		return runtime.ResourceKindReport
	case "alert":
		return runtime.ResourceKindAlert
	case "theme":
		return runtime.ResourceKindTheme
	case "component":
		return runtime.ResourceKindComponent
	case "canvas":
		return runtime.ResourceKindCanvas
	case "api":
		return runtime.ResourceKindAPI
	case "connector":
		return runtime.ResourceKindConnector
	default:
		return k
	}
}

func printResources(p *printer.Printer, resources map[string]map[string][]*runtimev1.Resource) {
	if len(resources) == 0 {
		p.PrintfWarn("No resources found\n")
		return
	}

	rows := make([]map[string]any, 0)
	for org, projectRes := range resources {
		for proj, res := range projectRes {
			for _, r := range res {
				row := make(map[string]any, 0)
				rows = append(rows, row)

				// each resource has a meta field and a resource(source/model/metricsview etc) which has spec and state fields
				// we want to flatten the resource to have the meta fields and spec and state fields at the top level
				rowJSON, err := protojson.Marshal(r)
				if err != nil {
					p.PrintfWarn("Failed to marshal resource for org %v, project %v : %v\n", org, proj, err)
					continue
				}
				err = json.Unmarshal(rowJSON, &row)
				if err != nil {
					p.PrintfWarn("Failed to unmarshal resource for org %v, project %v : %v\n", org, proj, err)
					continue
				}
				for k := range row {
					if k == "meta" {
						continue
					}
					resource, ok := row[k].(map[string]any)
					if !ok {
						delete(row, k)
						continue
					}
					row["spec"] = resource["spec"]
					row["state"] = resource["state"]
					delete(row, k)
					break
				}

				row["org"] = org
				row["project"] = proj
				row["resource_type"] = runtime.PrettifyResourceKind(r.Meta.Name.Kind)
				row["resource_name"] = r.Meta.Name.Name
			}
		}
	}

	jsonData, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		p.PrintfWarn("Failed to marshal resources: %v\n", err)
	}
	fmt.Println(string(jsonData))
}
