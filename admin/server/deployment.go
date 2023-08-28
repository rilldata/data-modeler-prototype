package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) TriggerReconcile(ctx context.Context, req *adminv1.TriggerReconcileRequest) (*adminv1.TriggerReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerReconcile(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerReconcileResponse{}, nil
}

func (s *Server) TriggerRefreshSources(ctx context.Context, req *adminv1.TriggerRefreshSourcesRequest) (*adminv1.TriggerRefreshSourcesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
		attribute.StringSlice("args.sources", req.Sources),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerRefreshSources(ctx, depl, req.Sources)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRefreshSourcesResponse{}, nil
}

func (s *Server) triggerRefreshSourcesInternal(w http.ResponseWriter, r *http.Request) {
	orgName := r.URL.Query().Get("organization")
	projectName := r.URL.Query().Get("project")
	if orgName == "" || projectName == "" {
		http.Error(w, "organization or project not specified", http.StatusBadRequest)
		return
	}

	proj, err := s.admin.DB.FindProjectByName(r.Context(), orgName, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if proj.ProdDeploymentID == nil {
		http.Error(w, "project does not have a deployment", http.StatusBadRequest)
		return
	}

	depl, err := s.admin.DB.FindDeployment(r.Context(), *proj.ProdDeploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.admin.TriggerRefreshSources(r.Context(), depl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) TriggerRedeploy(ctx context.Context, req *adminv1.TriggerRedeployRequest) (*adminv1.TriggerRedeployResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	// For backwards compatibility, this RPC supports passing either DeploymentId or Organization+Project names
	var proj *database.Project
	var depl *database.Deployment
	if req.DeploymentId != "" {
		var err error
		depl, err = s.admin.DB.FindDeployment(ctx, req.DeploymentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		proj, err = s.admin.DB.FindProject(ctx, depl.ProjectID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		var err error
		proj, err = s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if proj.ProdDeploymentID != nil {
			depl, err = s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err := s.admin.TriggerRedeploy(ctx, proj, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRedeployResponse{}, nil
}

// GetDeploymentCredentials returns runtime info and JWT on behalf of a specific user, or alternatively for a raw set of JWT attributes
func (s *Server) GetDeploymentCredentials(ctx context.Context, req *adminv1.GetDeploymentCredentialsRequest) (*adminv1.GetDeploymentCredentialsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.branch", req.Branch),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var deplToUse *database.Deployment
	if req.Branch == "" {
		prodDepl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		deplToUse = prodDepl
	} else {
		depls, err := s.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		for _, d := range depls {
			if d.Branch == req.Branch {
				deplToUse = d
				break
			}
		}

		if deplToUse == nil {
			return nil, status.Error(codes.InvalidArgument, "no deployment found for branch")
		}
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	// If the user is not a superuser, they must have ManageProd permissions
	if !permissions.ManageProd && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	var attr map[string]any
	switch id := req.For.(type) {
	case *adminv1.GetDeploymentCredentialsRequest_UserId:
		// Find User
		user, err := s.admin.DB.FindUser(ctx, id.UserId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		// Find User groups
		groups, err := s.admin.DB.FindUsergroupsForUser(ctx, user.ID, proj.OrganizationID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		groupNames := make([]string, len(groups))
		for i, group := range groups {
			groupNames[i] = group.Name
		}
		attr = map[string]any{
			"name":   user.DisplayName,
			"email":  user.Email,
			"domain": user.Email[strings.LastIndex(user.Email, "@")+1:],
			"groups": groupNames,
			"admin":  permissions.ManageProject,
		}
	case *adminv1.GetDeploymentCredentialsRequest_Attrs:
		attr = id.Attrs.AsMap()
	default:
		return nil, status.Error(codes.InvalidArgument, "id must be either user or attrs")
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: deplToUse.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         time.Hour,
		InstancePermissions: map[string][]runtimeauth.Permission{
			deplToUse.RuntimeInstanceID: {
				// TODO: Remove ReadProfiling and ReadRepo (may require frontend changes)
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadProfiling,
				runtimeauth.ReadRepo,
			},
		},
		Attributes: attr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(deplToUse.ID)

	return &adminv1.GetDeploymentCredentialsResponse{
		RuntimeHost:       deplToUse.RuntimeHost,
		RuntimeInstanceId: deplToUse.RuntimeInstanceID,
		Jwt:               jwt,
	}, nil
}
