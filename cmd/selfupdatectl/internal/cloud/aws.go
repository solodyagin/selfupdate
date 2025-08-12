package cloud

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWSSession represent a live session to AWS services
type AWSSession struct {
	client *s3.Client
	bucket string
}

// NewAWSSession create a new session
func NewAWSSession(accessKey string, secret string, endpoint string, region string, bucket string) (*AWSSession, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secret, "")),
	)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // For MinIO or S3-compatible services
	})
	if err != nil {
		return nil, err
	}

	return &AWSSession{client: client, bucket: bucket}, nil
}

// UploadFile to a S3 bucket
func (s *AWSSession) UploadFile(localFile string, s3FilePath string) error {
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return err
	}

	pa := &progressAWS{File: file, file: s3FilePath, contentLength: st.Size()}

	uploader := manager.NewUploader(s.client)

	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3FilePath),
		Body:   pa,
	})

	return err
}

type progressAWS struct {
	*os.File
	file          string
	contentLength int64
	downloaded    int64
	ticker        int
}

var _ io.Reader = (*progressAWS)(nil)
var _ io.ReaderAt = (*progressAWS)(nil)
var _ io.Seeker = (*progressAWS)(nil)

// Read file content
func (pa *progressAWS) Read(p []byte) (int, error) {
	return pa.File.Read(p)
}

// ReadAt specific offset in a file
func (pa *progressAWS) ReadAt(p []byte, off int64) (int, error) {
	n, err := pa.File.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	atomic.AddInt64(&pa.downloaded, int64(n))

	fmt.Printf("\r%v: %v%% %c", pa.file, 100*pa.downloaded/(pa.contentLength*2), pa.tick())

	return n, err
}

// Seek in a file
func (pa *progressAWS) Seek(offset int64, whence int) (int64, error) {
	return pa.File.Seek(offset, whence)
}

// Size return the file content length
func (pa *progressAWS) Size() int64 {
	return pa.contentLength
}

var ticker = `|\-/`

func (pa *progressAWS) tick() rune {
	pa.ticker = (pa.ticker + 1) % len(ticker)
	return rune(ticker[pa.ticker])
}
