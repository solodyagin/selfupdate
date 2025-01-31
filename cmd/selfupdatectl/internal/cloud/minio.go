package cloud

import (
	"context"
	"net/url"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOSession struct {
	sess   *minio.Client
	bucket string
}

// NewMinIOSessionFromEnvironment create a new session from environment variable.
// This will be looking for the environment variable MINIO_ENDPOINT, MINIO_REGION and MINIO_BUCKET
func NewMinIOSessionFromEnvironment() (*MinIOSession, error) {
	return NewMinIOSession("", "", os.Getenv("MINIO_ENDPOINT"), os.Getenv("MINIO_REGION"), os.Getenv("MINIO_BUCKET"))
}

// NewMinIOSession create a new session
func NewMinIOSession(accessKey string, secret string, endpoint string, region string, bucket string) (*MinIOSession, error) {
	var creds *credentials.Credentials

	if accessKey != "" && secret != "" {
		creds = credentials.NewStaticV4(accessKey, secret, "")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	secure := true
	if u.Scheme == "http" {
		secure = false
	}
	endpoint = u.Host + u.Path

	sess, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: secure,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOSession{sess: sess, bucket: bucket}, nil
}

// UploadFile to a MinIO S3 bucket
func (s *MinIOSession) UploadFile(localFile string, s3FilePath string) error {
	ctx := context.Background()

	_, err := s.sess.FPutObject(ctx, s.bucket, s3FilePath, localFile, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return err
}

// GetBucket associated with a session
func (s *MinIOSession) GetBucket() string {
	return s.bucket
}
