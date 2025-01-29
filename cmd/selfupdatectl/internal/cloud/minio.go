package cloud

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	c      *minio.Client
	bucket string
}

// NewMinIOClientFromEnvironment create a new client from environment variable.
// This will be looking for the environment variable AWS_S3_ENDPOINT, AWS_S3_REGION and AWS_S3_BUCKET
func NewMinIOClientFromEnvironment() (*MinIOClient, error) {
	return NewMinIOClient("", "", os.Getenv("AWS_S3_ENDPOINT"), os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_S3_BUCKET"))
}

// NewMinIOClient create a new client
func NewMinIOClient(akid string, secret string, endpoint string, region string, bucket string) (*MinIOClient, error) {
	var cred *credentials.Credentials

	if akid != "" && secret != "" {
		cred = credentials.NewStaticV4(akid, secret, "")
	}

	c, err := minio.New(endpoint, &minio.Options{
		Creds:  cred,
		Secure: true,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOClient{c: c, bucket: bucket}, nil
}

// UploadFile to a MinIO bucket
func (a *MinIOClient) UploadFile(localFile string, s3FilePath string) error {
	ctx := context.Background()

	_, err := a.c.FPutObject(ctx, a.bucket, s3FilePath, localFile, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return err
}

// GetBucket associated with a client
func (a *MinIOClient) GetBucket() string {
	return a.bucket
}
