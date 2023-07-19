package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListFiles implements RuntimeService.
func (s *Server) ListFiles(ctx context.Context, req *runtimev1.ListFilesRequest) (*runtimev1.ListFilesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.glob", req.Glob),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	glob := req.Glob
	if glob == "" {
		glob = "**"
	}

	paths, err := s.runtime.ListFiles(ctx, req.InstanceId, glob)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &runtimev1.ListFilesResponse{Paths: paths}, nil
}

// WatchFiles implements RuntimeService.
func (s *Server) WatchFiles(req *runtimev1.WatchFilesRequest, ss runtimev1.RuntimeService_WatchFilesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.replay", req.Replay),
	)

	if !auth.GetClaims(ss.Context()).CanInstance(req.InstanceId, auth.ReadRepo) {
		return ErrForbidden
	}

	repo, err := s.runtime.Repo(ss.Context(), req.InstanceId)
	if err != nil {
		return err
	}

	return repo.Watch(ss.Context(), req.Replay, func(event drivers.WatchEvent) error {
		if !event.Dir {
			err := ss.Send(&runtimev1.WatchFilesResponse{
				Event: event.Type,
				Path:  event.Path,
			})

			return err
		}
		return nil
	})
}

// GetFile implements RuntimeService.
func (s *Server) GetFile(ctx context.Context, req *runtimev1.GetFileRequest) (*runtimev1.GetFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadRepo) {
		return nil, ErrForbidden
	}

	blob, lastUpdated, err := s.runtime.GetFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetFileResponse{Blob: blob, UpdatedOn: timestamppb.New(lastUpdated)}, nil
}

// PutFile implements RuntimeService.
func (s *Server) PutFile(ctx context.Context, req *runtimev1.PutFileRequest) (*runtimev1.PutFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
		attribute.Bool("args.create", req.Create),
		attribute.Bool("args.create_only", req.CreateOnly),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileResponse{}, nil
}

// DeleteFile implements RuntimeService.
func (s *Server) DeleteFile(ctx context.Context, req *runtimev1.DeleteFileRequest) (*runtimev1.DeleteFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.path", req.Path),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileResponse{}, nil
}

// RenameFile implements RuntimeService.
func (s *Server) RenameFile(ctx context.Context, req *runtimev1.RenameFileRequest) (*runtimev1.RenameFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.from_path", req.FromPath),
		attribute.String("args.to_path", req.ToPath),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileResponse{}, nil
}

// UploadMultipartFile implements the same functionality as PutFile, but for multipart HTTP upload.
// It's mounted only on as a REST API and enables upload of large files (such as data files).
func (s *Server) UploadMultipartFile(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	if !auth.GetClaims(req.Context()).CanInstance(pathParams["instance_id"], auth.EditRepo) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	if err := req.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	f, _, err := req.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse file in request: %s", err), http.StatusBadRequest)
		return
	}

	if pathParams["path"] == "" {
		http.Error(w, "must have a path to file", http.StatusBadRequest)
		return
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", pathParams["instance_id"]),
		attribute.String("args.path", pathParams["path"]),
	)

	err = s.runtime.PutFile(ctx, pathParams["instance_id"], pathParams["path"], f, true, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %s", err), http.StatusBadRequest)
		return
	}

	res, err := protojson.Marshal(&runtimev1.PutFileResponse{})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to serialize response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
		return
	}
}
