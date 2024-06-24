package server

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 100 MB
const maxAssetSize = 104857600

var signingHeaderMap = map[string]string{
	"Content-Type":                "application/octet-stream",
	"x-goog-content-length-range": fmt.Sprintf("1,%d", maxAssetSize),
}

// a copy of signingHeaderMap but kept in array form to pass to SignedURL API
var signingHeaders = []string{
	"Content-Type:application/octet-stream",
	fmt.Sprintf("x-goog-content-length-range:1,%d", maxAssetSize), // validates that the request body is between 1 byte to 100MB
}

func (s *Server) CreateAsset(ctx context.Context, req *adminv1.CreateAssetRequest) (*adminv1.CreateAssetResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.OrganizationName),
		attribute.String("args.type", req.Type),
	)

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser && claims.OwnerType() != auth.OwnerTypeService {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Find parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions
	// create asset and create project should be the same permission
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create assets")
	}

	// generate a signed url
	object := path.Join(req.Type, fmt.Sprintf("%s__%s.%s", req.Name, uuid.New().String(), req.Extension))
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "PUT",
		Headers: signingHeaders,
		Expires: time.Now().Add(15 * time.Minute),
	}
	signedURL, err := s.assetsBucket.SignedURL(object, opts)
	if err != nil {
		return nil, err
	}

	// create an asset
	assetPath, err := s.assetPath(object)
	if err != nil {
		return nil, err
	}

	asset, err := s.admin.DB.InsertAsset(ctx, org.ID, assetPath, claims.OwnerID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert asset: %s", err.Error())
	}

	return &adminv1.CreateAssetResponse{
		AssetId:        asset.ID,
		SignedUrl:      signedURL,
		SigningHeaders: signingHeaderMap,
	}, nil
}

func (s *Server) assetPath(object string) (string, error) {
	uploadPath, err := url.Parse(s.opts.AssetsBucket)
	if err != nil {
		return "", err
	}
	uploadPath.Host = s.opts.AssetsBucket
	uploadPath.Scheme = "gs"
	uploadPath.Path = object
	return uploadPath.String(), nil
}
