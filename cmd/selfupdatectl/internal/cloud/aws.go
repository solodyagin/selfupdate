package cloud

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AWSSession represent a live session to AWS services
type AWSSession struct {
	sess   *session.Session
	bucket string
}

// NewAWSSessionFromEnvironment create a new session from environment variable.
// This will be looking for the environment variable AWS_S3_ENDPOINT, AWS_S3_REGION and AWS_S3_BUCKET
func NewAWSSessionFromEnvironment() (*AWSSession, error) {
	return NewAWSSession("", "", os.Getenv("AWS_S3_ENDPOINT"), os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_S3_BUCKET"))
}

// NewAWSSession create a new session
func NewAWSSession(accessKey string, secret string, endpoint string, region string, bucket string) (*AWSSession, error) {
	var creds *credentials.Credentials

	if accessKey != "" && secret != "" {
		creds = credentials.NewStaticCredentials(accessKey, secret, "")
	}

	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}

	return &AWSSession{sess: sess, bucket: bucket}, nil
}

// GetCredentials from the established session
func (s *AWSSession) GetCredentials() (credentials.Value, error) {
	return s.sess.Config.Credentials.Get()
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

	uploader := s3manager.NewUploader(s.sess)

	_, err = uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3FilePath),
		Body:   pa,
	})

	return err
}

// GetBucket associated with a session
func (s *AWSSession) GetBucket() string {
	return s.bucket
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
