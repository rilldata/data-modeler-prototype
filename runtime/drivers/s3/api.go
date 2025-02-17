package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Connection) ListBuckets(ctx context.Context) ([]string, error) {
	creds, err := c.newCredentials()
	if err != nil {
		return nil, err
	}
	if creds == credentials.AnonymousCredentials {
		return nil, fmt.Errorf("no credentials exist")
	}

	sharedConfigState := session.SharedConfigDisable
	if c.config.AllowHostAccess {
		sharedConfigState = session.SharedConfigEnable // Tells to look for default region set with `aws configure`
	}
	// Create a session that tries to infer the region from the environment
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: sharedConfigState,
		Config: aws.Config{
			Credentials: creds,
		},
	})
	if err != nil {
		return nil, err
	}

	// no region found, default to us-east-1
	if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}
	svc := s3.New(sess)
	output, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(output.Buckets))
	for _, bucket := range output.Buckets {
		if bucket.Name != nil {
			buckets = append(buckets, *bucket.Name)
		}
	}
	return buckets, nil
}

func (c *Connection) ListObjectsRaw(ctx context.Context, req *runtimev1.S3ListObjectsRequest) ([]*runtimev1.S3Object, string, error) {
	creds, err := c.newCredentials()
	if err != nil {
		return nil, "", err
	}

	bucket, err := c.openBucket(ctx, &sourceProperties{AWSRegion: req.Region}, req.Bucket, creds)
	if err != nil {
		return nil, "", err
	}
	defer bucket.Close()

	var pageToken []byte
	if req.PageToken == "" {
		pageToken = blob.FirstPageToken
	} else {
		pageToken = []byte(req.PageToken)
	}

	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	objects, nextToken, err := fetchObjects(ctx, bucket, pageToken, pageSize, req)
	if err != nil {
		var failureErr awserr.RequestFailure
		if !errors.As(err, &failureErr) {
			return nil, "", err
		}

		if (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) && creds != credentials.AnonymousCredentials {
			// try again with anonymous credentials
			creds = credentials.AnonymousCredentials
			bucketObj, bucketErr := c.openBucket(ctx, &sourceProperties{AWSRegion: req.Region}, req.Bucket, creds)
			if bucketErr != nil {
				return nil, "", fmt.Errorf("failed to open bucket %q, %w", req.Bucket, bucketErr)
			}
			objects, nextToken, err = fetchObjects(ctx, bucketObj, pageToken, pageSize, req)
		}
	}
	if err != nil {
		return nil, "", err
	}

	s3Objects := make([]*runtimev1.S3Object, len(objects))
	for i, object := range objects {
		s3Objects[i] = &runtimev1.S3Object{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return s3Objects, string(nextToken), nil
}

func (c *Connection) GetBucketRegion(ctx context.Context, bkt string) (string, error) {
	creds, err := c.newCredentials()
	if err != nil {
		return "", err
	}

	sess, err := c.newSessionForBucket(ctx, bkt, "", "", creds)
	if err != nil {
		return "", err
	}

	if sess.Config.Region != nil {
		return *sess.Config.Region, nil
	}
	return "", fmt.Errorf("unable to get region")
}

func (c *Connection) GetCredentialsInfo(ctx context.Context) (provider string, exist bool, err error) {
	creds, err := c.newCredentials()
	if creds == credentials.AnonymousCredentials {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	val, err := creds.Get()
	if err != nil {
		return "", false, err
	}

	return val.ProviderName, true, nil
}

func fetchObjects(ctx context.Context, bucket *blob.Bucket, pageToken []byte, pageSize int, req *runtimev1.S3ListObjectsRequest) ([]*blob.ListObject, []byte, error) {
	objects, nextToken, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix:    req.Prefix,
		Delimiter: req.Delimiter,
		BeforeList: func(as func(interface{}) bool) error {
			if req.StartAfter == "" {
				return nil
			}
			var q *s3.ListObjectsV2Input
			if as(&q) {
				q.StartAfter = &req.StartAfter
			}
			return nil
		},
	})
	return objects, nextToken, err
}
