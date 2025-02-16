package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/azure"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

var errGCSUsesNativeCreds = errors.New("GCS uses native credentials")

type objectStoreInputProps struct {
	Path string `mapstructure:"path"`
	// URI is renamed to path now. It is kept for backward compatibility.
	URI    string             `mapstructure:"uri"`
	Format drivers.FileFormat `mapstructure:"format"`
	DuckDB map[string]any     `mapstructure:"duckdb"`
}

func (p *objectStoreInputProps) Validate() error {
	if p.Path == "" && p.URI == "" {
		return fmt.Errorf("missing property `path`")
	}
	if p.Path != "" && p.URI != "" {
		return fmt.Errorf("cannot specify both `path` and `uri`")
	}
	p.Path = p.URI
	return nil
}

type objectStoreToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &objectStoreToSelfExecutor{}

func (e *objectStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *objectStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputConnector, opts.InputHandle, opts.InputProperties)
	if err != nil {
		if errors.Is(err, errGCSUsesNativeCreds) {
			e := &objectStoreToSelfExecutorNonNative{c: e.c}
			return e.Execute(ctx, opts)
		}
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *objectStoreToSelfExecutor) modelInputProperties(model, inputConnector string, inputHandle drivers.Handle, inputProps map[string]any) (map[string]any, error) {
	parsed := &objectStoreInputProps{}
	if err := mapstructure.WeakDecode(inputProps, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := parsed.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	m := &ModelInputProperties{}
	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	// Generate secret SQL to access the service and set as pre_exec_query
	var err error
	m.PreExec, err = objectStoreSecretSQL(model, inputConnector, inputHandle, inputProps)
	if err != nil {
		return nil, err
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{parsed.Path}, format, parsed.DuckDB)
	if err != nil {
		return nil, err
	}
	m.SQL = "SELECT * FROM " + from

	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

func objectStoreSecretSQL(model, inputConnector string, inputHandle drivers.Handle, inputProps map[string]any) (string, error) {
	config := inputHandle.Config()
	// config properties can also be set as input properties
	maps.Copy(config, inputProps)
	safeSecretName := safeName(fmt.Sprintf("%s__%s__secret", model, inputConnector))
	switch inputHandle.Driver() {
	case "s3":
		s3Config := &s3.ConfigProperties{}
		err := mapstructure.WeakDecode(config, s3Config)
		if err != nil {
			return "", fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE S3")
		if s3Config.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if s3Config.AccessKeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(s3Config.AccessKeyID), safeSQLString(s3Config.SecretAccessKey))
		}
		if s3Config.SessionToken != "" {
			fmt.Fprintf(&sb, ", SESSION_TOKEN %s", safeSQLString(s3Config.SessionToken))
		}
		if s3Config.Endpoint != "" {
			sb.WriteString(", ENDPOINT ")
			sb.WriteString(safeSQLString(s3Config.Endpoint))
		}
		if s3Config.Region != "" {
			sb.WriteString(", REGION ")
			sb.WriteString(safeSQLString(s3Config.Region))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	case "gcs":
		// GCS works via S3 compatibility mode
		gcsConfig, err := gcs.NewConfigProperties(config)
		if err != nil {
			return "", err
		}
		// If no credentials are provided we assume that the user wants to use the native credentials
		if gcsConfig.SecretJSON != "" || (gcsConfig.KeyID == "" && gcsConfig.Secret == "" && gcsConfig.SecretJSON == "") {
			return "", errGCSUsesNativeCreds
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE GCS")
		if gcsConfig.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if gcsConfig.KeyID != "" {
			fmt.Fprintf(&sb, ", KEY_ID %s, SECRET %s", safeSQLString(gcsConfig.KeyID), safeSQLString(gcsConfig.Secret))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	case "azure":
		azureConfig := &azure.ConfigProperties{}
		err := mapstructure.WeakDecode(config, azureConfig)
		if err != nil {
			return "", fmt.Errorf("failed to parse s3 config properties: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("CREATE OR REPLACE TEMPORARY SECRET ")
		sb.WriteString(safeSecretName)
		sb.WriteString(" (TYPE AZURE")
		if azureConfig.AllowHostAccess {
			sb.WriteString(", PROVIDER CREDENTIAL_CHAIN")
		}
		if azureConfig.ConnectionString != "" {
			fmt.Fprintf(&sb, ", CONNECTION_STRING %s", safeSQLString(azureConfig.ConnectionString))
		}
		if azureConfig.Account != "" {
			fmt.Fprintf(&sb, ", ACCOUNT_NAME %s", safeSQLString(azureConfig.Account))
		}
		sb.WriteRune(')')
		return sb.String(), nil
	default:
		return "", fmt.Errorf("internal error: unsupported object store: %s", inputHandle.Driver())
	}
}

// objectStoreToSelfExecutorNonNative is a non-native implementation of objectStoreToSelfExecutor.
// It uses Rill's own connectors instead of duckdb's native connectors.
type objectStoreToSelfExecutorNonNative struct {
	c *connection
}

func (e *objectStoreToSelfExecutorNonNative) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	parsed := &objectStoreInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := parsed.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	store, ok := opts.InputHandle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("input handle is not an object store")
	}

	iter, err := store.DownloadFiles(ctx, opts.InputProperties)
	if err != nil {
		return nil, err
	}

	var (
		res           *drivers.ModelResult
		resErr        error
		appendToTable = false
	)
	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}
	for {
		files, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		m := &ModelInputProperties{}
		from, err := sourceReader(files, format, parsed.DuckDB)
		if err != nil {
			return nil, err
		}
		m.SQL = "SELECT * FROM " + from
		propsMap := make(map[string]any)
		if err := mapstructure.Decode(m, &propsMap); err != nil {
			return nil, err
		}
		opts.InputProperties = propsMap

		if appendToTable {
			opts.Incremental = true
			opts.IncrementalRun = true
			opts.PreviousResult = res
		}
		appendToTable = true
		executor := &selfToSelfExecutor{c: e.c}
		res, resErr = executor.Execute(ctx, opts)
		if resErr != nil {
			return nil, resErr
		}
	}
	if res == nil && resErr == nil {
		// the iterator returns an error if no files are found
		// this should never happen
		return nil, fmt.Errorf("no result")
	}
	return res, resErr
}
