package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const magicAuthTokenFilterMaxSize = 1024

func (s *Server) IssueMagicAuthToken(ctx context.Context, req *adminv1.IssueMagicAuthTokenRequest) (*adminv1.IssueMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.display_name", req.DisplayName),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find organization for project: %v", err.Error())
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.CreateMagicAuthTokens {
		return nil, status.Error(codes.PermissionDenied, "not allowed to create a magic auth token")
	}

	if req.ResourceName != "" && req.ResourceType != "" { // nolint:staticcheck // for backwards compatibility
		addResource := true
		for _, r := range req.Resources {
			if r.Type == req.ResourceType && r.Name == req.ResourceName { // nolint:staticcheck // for backwards compatibility
				addResource = false
				break
			}
		}
		if addResource {
			req.Resources = append(req.Resources, &adminv1.ResourceName{
				Type: req.ResourceType, // nolint:staticcheck // for backwards compatibility
				Name: req.ResourceName, // nolint:staticcheck // for backwards compatibility
			})
		}
	}

	resources := make([]database.ResourceName, len(req.Resources))
	for i, r := range req.Resources {
		resources[i] = database.ResourceName{
			Type: r.Type,
			Name: r.Name,
		}
	}
	opts := &admin.IssueMagicAuthTokenOptions{
		ProjectID:   proj.ID,
		Fields:      req.Fields,
		State:       req.State,
		DisplayName: req.DisplayName,
		Resources:   resources,
	}

	if req.TtlMinutes != 0 {
		ttl := time.Duration(req.TtlMinutes) * time.Minute
		opts.TTL = &ttl
	}

	if claims.OwnerType() == auth.OwnerTypeUser {
		id := claims.OwnerID()
		opts.CreatedByUserID = &id

		// Generate JWT attributes based on the creating user's, but with limited project-level permissions.
		// We store these attributes with the magic token, so it can simulate the creating user (even if the creating user is later deleted or their permissions change).
		//
		// NOTE: A problem with this approach is that if we change the built-in format of JWT attributes, these will remain as they were when captured.
		// NOTE: Another problem is that if the creator is an admin, attrs["admin"] will be true. It shouldn't be a problem today, but could end up leaking some privileges in the future if we're not careful.
		attrs, err := s.jwtAttributesForUser(ctx, claims.OwnerID(), proj.OrganizationID, projPerms)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		opts.Attributes = attrs
	}

	if req.Filter != nil {
		val, err := protojson.Marshal(req.Filter)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if len(val) > magicAuthTokenFilterMaxSize {
			return nil, status.Errorf(codes.InvalidArgument, "filter size exceeds limit (got %d bytes, but the limit is %d bytes)", len(val), magicAuthTokenFilterMaxSize)
		}

		opts.FilterJSON = string(val)
	}

	token, err := s.admin.IssueMagicAuthToken(ctx, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenStr := token.Token().String()
	return &adminv1.IssueMagicAuthTokenResponse{
		Token: tokenStr,
		Url:   s.admin.URLs.WithCustomDomain(org.CustomDomain).MagicAuthTokenOpen(req.Organization, req.Project, tokenStr),
	}, nil
}

func (s *Server) GetCurrentMagicAuthToken(ctx context.Context, req *adminv1.GetCurrentMagicAuthTokenRequest) (*adminv1.GetCurrentMagicAuthTokenResponse, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeMagicAuthToken {
		return nil, status.Error(codes.PermissionDenied, "request was not made with a magic auth token")
	}

	tkn, err := s.admin.DB.FindMagicAuthTokenWithUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("project with id %s not found", tkn.ProjectID))
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find organization for project: %v", err.Error())
	}

	pb, err := s.magicAuthTokenToPB(tkn, org, proj)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.GetCurrentMagicAuthTokenResponse{
		Token: pb,
	}, nil
}

func (s *Server) ListMagicAuthTokens(ctx context.Context, req *adminv1.ListMagicAuthTokensRequest) (*adminv1.ListMagicAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.CreateMagicAuthTokens && !projPerms.ManageMagicAuthTokens {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find organization for project: %v", err.Error())
	}

	var createdByUserID *string
	if !projPerms.ManageMagicAuthTokens {
		if claims.OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
		}

		id := claims.OwnerID()
		createdByUserID = &id
	}

	tokens, err := s.admin.DB.FindMagicAuthTokensWithUser(ctx, proj.ID, createdByUserID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if len(tokens) >= pageSize {
		nextPageToken = marshalPageToken(tokens[len(tokens)-1].ID)
	}

	pbs, err := s.magicAuthTokensToPB(tokens, org, proj)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ListMagicAuthTokensResponse{
		Tokens:        pbs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) RevokeMagicAuthToken(ctx context.Context, req *adminv1.RevokeMagicAuthTokenRequest) (*adminv1.RevokeMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	tkn, err := s.admin.DB.FindMagicAuthToken(ctx, req.TokenId, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find project for token: %v", err.Error())
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.ManageMagicAuthTokens {
		// If they don't have manage permissions, they can only revoke tokens they created themselves.
		isCreator := tkn.CreatedByUserID != nil && *tkn.CreatedByUserID == claims.OwnerID()
		if !projPerms.CreateMagicAuthTokens || !isCreator {
			return nil, status.Error(codes.PermissionDenied, "not allowed to revoke this magic auth token")
		}
	}

	err = s.admin.DB.DeleteMagicAuthToken(ctx, tkn.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RevokeMagicAuthTokenResponse{}, nil
}

func (s *Server) magicAuthTokensToPB(tkns []*database.MagicAuthTokenWithUser, org *database.Organization, proj *database.Project) ([]*adminv1.MagicAuthToken, error) {
	var pbs []*adminv1.MagicAuthToken
	for _, tkn := range tkns {
		pb, err := s.magicAuthTokenToPB(tkn, org, proj)
		if err != nil {
			return nil, err
		}
		pbs = append(pbs, pb)
	}
	return pbs, nil
}

func (s *Server) magicAuthTokenToPB(tkn *database.MagicAuthTokenWithUser, org *database.Organization, proj *database.Project) (*adminv1.MagicAuthToken, error) {
	attrs, err := structpb.NewStruct(tkn.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert attributes to structpb: %w", err)
	}

	var filter *runtimev1.Expression
	if tkn.FilterJSON != "" {
		filter = &runtimev1.Expression{}
		err := protojson.Unmarshal([]byte(tkn.FilterJSON), filter)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter: %w", err)
		}
	}

	// backwards compatibility
	tokenStr := ""
	url := ""
	if len(tkn.Secret) != 0 {
		id, err := uuid.Parse(tkn.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token ID: %w", err)
		}
		token, err := authtoken.FromParts(authtoken.TypeMagic, id, tkn.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to create magic auth token from parts: %w", err)
		}
		tokenStr = token.String()
		url = s.admin.URLs.WithCustomDomain(org.CustomDomain).MagicAuthTokenOpen(org.Name, proj.Name, tokenStr)
	}

	rs := make([]*adminv1.ResourceName, len(tkn.Resources))
	for i, r := range tkn.Resources {
		rs[i] = &adminv1.ResourceName{
			Type: r.Type,
			Name: r.Name,
		}
	}

	res := &adminv1.MagicAuthToken{
		Id:                 tkn.ID,
		ProjectId:          tkn.ProjectID,
		Url:                url,
		Token:              tokenStr,
		CreatedOn:          timestamppb.New(tkn.CreatedOn),
		ExpiresOn:          nil,
		UsedOn:             timestamppb.New(tkn.UsedOn),
		CreatedByUserId:    safeStr(tkn.CreatedByUserID),
		CreatedByUserEmail: tkn.CreatedByUserEmail,
		Attributes:         attrs,
		Resources:          rs,
		Filter:             filter,
		Fields:             tkn.Fields,
		State:              tkn.State,
		DisplayName:        tkn.DisplayName,
	}
	if tkn.ExpiresOn != nil {
		res.ExpiresOn = timestamppb.New(*tkn.ExpiresOn)
	}
	return res, nil
}
