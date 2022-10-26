package s3

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
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
			Hint:        "Note that gzipped files & glob patterns aren't yet supported",
		},
		{
			Key:         "aws.region",
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
			Hint:        "Set your local credentials: <code>aws configure</code>. Click to learn more.",
			Href:        "https://docs.rilldata.com/import-data#setting-amazon-s3-credentials",
		},
	},
}

type Config struct {
	Path      string `mapstructure:"path" ignored:"true"`
	AWSRegion string `mapstructure:"aws.region"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, source *connectors.Source, callback func(filename string) error) error {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// The session the S3 Downloader will use
	sess, err := getAwsSessionConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	bucket, key, extension, err := getAwsUrlParts(conf.Path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s, %v", conf.Path, err)
	}

	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*.%s", source.Name, extension),
	)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	// Write the contents of S3 Object to the f
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download f, %v", err)
	}

	err = callback(f.Name())
	if err != nil {
		return fmt.Errorf("failed to ingest f, %v", err)
	}

	return nil
}

func getAwsSessionConfig(conf *Config) (*session.Session, error) {
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region: aws.String(conf.AWSRegion),
		})
	}
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
}

func getAwsUrlParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}

	p := strings.Split(u.Path, ".")

	return u.Host, u.Path, p[len(p)-1], nil
}
