package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) assetsHandler(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	path := req.PathValue("path")

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.path", path),
	)

	if !auth.GetClaims(req.Context()).CanInstance(instanceID, auth.ReadObjects) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to assets")
	}

	repo, release, err := s.runtime.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	paths := repo.GetCachedPaths()
	allowed := false
	for _, p := range paths {
		if strings.HasPrefix(strings.TrimLeft(path, ","), strings.TrimLeft(p, ",")) {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("path is not allowed")
	}

	str, err := repo.Get(ctx, path)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))
	return err
}
