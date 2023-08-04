package sources

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var protocolExtraction = regexp.MustCompile(`^(\w*?)://(.*)$`)

func ParseEmbeddedSource(path string) (*runtimev1.Source, bool) {
	path, connector, ok := parseEmbeddedSourceConnector(path)
	if !ok {
		return nil, false
	}

	hash := md5.Sum([]byte(fmt.Sprintf("%s_%s", connector, path)))
	// prepend a char to make sure the name is a valid table name
	name := "a" + hex.EncodeToString(hash[:])

	props, err := structpb.NewStruct(map[string]any{
		"path": path,
	})
	if err != nil {
		// shouldn't happen
		return nil, false
	}
	return &runtimev1.Source{
		Name:       name,
		Connector:  connector,
		Properties: props,
	}, true
}

func parseEmbeddedSourceConnector(path string) (string, string, bool) {
	path = strings.TrimSpace(strings.Trim(path, `"'`))
	matches := protocolExtraction.FindStringSubmatch(path)
	var connector string
	if len(matches) < 3 {
		if strings.Contains(path, "/") {
			connector = "local_file"
		} else {
			return "", "", false
		}
	} else {
		switch matches[1] {
		case "http", "https":
			connector = "https"
		case "s3":
			connector = "s3"
		case "gs":
			connector = "gcs"
		default:
			return "", "", false
		}
	}
	return path, connector, true
}
