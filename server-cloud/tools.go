//go:build tools

package server_cloud

// Tools installed with go install that `go mod tidy` should keep.
import (
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
)
