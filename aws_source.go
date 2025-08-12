package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSSource struct {
	client *s3.Client
	bucket string
	key    string
}

var _ Source = (*AWSSource)(nil)

func NewAWSSource(client *s3.Client, bucket string, base string) Source {
	key := replaceURLTemplate(base)
	return &AWSSource{client: client, bucket: bucket, key: key}
}

// Get will return if it succeed an io.ReaderCloser to the new executable being downloaded and its length
func (s *AWSSource) Get(v *Version) (io.ReadCloser, int64, error) {
	obj, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	})
	if err != nil {
		return nil, 0, err
	}

	return obj.Body, aws.ToInt64(obj.ContentLength), nil
}

// GetSignature will return the content of ${URL}.ed25519
func (s *AWSSource) GetSignature() ([64]byte, error) {
	obj, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key + ".ed25519"),
	})
	if err != nil {
		return [64]byte{}, err
	}
	defer obj.Body.Close()

	size := aws.ToInt64(obj.ContentLength)
	if size != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", size)
	}

	writer := bytes.NewBuffer(make([]byte, 0, 64))
	n, err := io.Copy(writer, obj.Body)
	if err != nil {
		return [64]byte{}, err
	}

	if n != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", n)
	}

	r := [64]byte{}
	copy(r[:], writer.Bytes())

	return r, nil
}

// LatestVersion will return the LastModified time
func (s *AWSSource) LatestVersion() (*Version, error) {
	info, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	})
	if err != nil {
		return nil, err
	}

	return &Version{Date: aws.ToTime(info.LastModified)}, nil
}
