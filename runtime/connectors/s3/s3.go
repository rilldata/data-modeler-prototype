package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcerrors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName: "Amazon S3",
	Description: "Connect to AWS S3 Storage.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Note that glob patterns aren't yet supported",
		},
		{
			Key:         "region",
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    false,
			Hint:        "Rill will use the default region in your local AWS config, unless set here.",
		},
		{
			Key:         "aws.credentials",
			DisplayName: "AWS credentials",
			Description: "AWS credentials inferred from your local environment.",
			Type:        connectors.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>aws configure</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/using-rill/import-data#setting-amazon-s3-credentials",
		},
	},
}

type Config struct {
	Path                  string `mapstructure:"path"`
	AWSRegion             string `mapstructure:"region"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	S3Endpoint            string `mapstructure:"endpoint"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source) (connectors.FileIterator, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}

	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path %q, should start with s3://", conf.Path)
	}

	creds, err := getCredentials()
	if err != nil {
		return nil, err
	}

	bucketObj, err := fetchBucketObj(ctx, conf, url.Host, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", url.Host, err)
	}

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           url.Path,
		ExtractPolicy:         source.ExtractPolicy,
	}
	it, err := rillblob.NewIterator(ctx, bucketObj, opts)

	if gcerrors.Code(err) == gcerrors.PermissionDenied && creds != credentials.AnonymousCredentials {
		// s3 throws permission denied error in case we are trying to access public buckets and passing some credentials
		// we try again with anonymous credentials in case bucket is public
		creds = credentials.AnonymousCredentials
		bucketObj, err := fetchBucketObj(ctx, conf, url.Host, creds)
		if err != nil {
			return nil, fmt.Errorf("failed to open bucket %q, %w", url.Host, err)
		}
		return rillblob.NewIterator(ctx, bucketObj, opts)
	}

	return it, err
}

func fetchBucketObj(ctx context.Context, conf *Config, bucket string, creds *credentials.Credentials) (*blob.Bucket, error) {
	sess, err := getAwsSessionConfig(ctx, conf, bucket, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return s3blob.OpenBucket(ctx, sess, bucket, nil)
}

func getAwsSessionConfig(ctx context.Context, conf *Config, bucket string, creds *credentials.Credentials) (*session.Session, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if len(conf.S3Endpoint) > 0 {
		region := conf.AWSRegion
		if region == "" {
			// Set the default region for bwd compatibility reasons
			// cloudflare and minio ignore if us-east-1 is set, not tested for others
			region = "us-east-1"
		}
		return session.NewSession(&aws.Config{
			Region:           aws.String(region),
			Endpoint:         &conf.S3Endpoint,
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      creds,
		})
	}
	// The logic below is AWS-specific, so we ignore it when conf.S3Endpoint is set
	// The complexity below relates to AWS being pretty strict about regions (probably to avoid unexpected cross-region traffic).

	// If the user explicitly set a region, we use that
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region:      aws.String(conf.AWSRegion),
			Credentials: creds,
		})
	}

	// Create a session that tries to infer the region from the environment
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Tells to look for default region set with `aws configure`
		Config: aws.Config{
			Credentials: creds,
		},
	})
	if err != nil {
		return nil, err
	}

	// If no region was found, we default to us-east-1 (which will be used to resolve the lookup in the next step)
	if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}

	// Bucket names are globally unique, but requests will fail if their region doesn't match the one configured in the session.
	// So we do a lookup for the bucket's region and configure the session to use that.
	reg, err := s3manager.GetBucketRegion(ctx, sess, bucket, "")
	if err != nil {
		return nil, err
	}
	if reg != "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(reg)})
	}

	return sess, nil
}

func getCredentials() (*credentials.Credentials, error) {
	// Find credentials to use.
	// If no local credentials are found, we fallback to AnonymousCredentials.
	// AnonymousCredentials can't be chained, so we try to resolve local creds, and use anon if none were found.
	// The chain used here is a duplicate of defaults.CredProviders(), but without the remote credentials lookup (since they resolve too slowly).
	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
	})
	_, err := creds.Get()
	if err != nil {
		if !errors.Is(err, credentials.ErrNoValidProvidersFoundInChain) {
			return nil, err
		}
		creds = credentials.AnonymousCredentials
	}

	return creds, nil
}
