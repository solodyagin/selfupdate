package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinIOSource struct {
	client     *minio.Client
	bucketName string
	baseURL    string
}

var _ Source = (*MinIOSource)(nil)

func NewMinIOSource(client *minio.Client, bucketName string, base string) Source {
	base = replaceURLTemplate(base)

	return &MinIOSource{client: client, bucketName: bucketName, baseURL: base}
}

// Get will return if it succeed an io.ReaderCloser to the new executable being downloaded and its length
func (s *MinIOSource) Get(v *Version) (io.ReadCloser, int64, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucketName, s.baseURL, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}

	info, err := obj.Stat()
	if err != nil {
		return nil, 0, err
	}

	return obj, info.Size, nil
}

// GetSignature will return the content of ${URL}.ed25519
func (s *MinIOSource) GetSignature() ([64]byte, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucketName, s.baseURL+".ed25519", minio.GetObjectOptions{})
	if err != nil {
		return [64]byte{}, err
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		return [64]byte{}, err
	}

	if info.Size != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", info.Size)
	}

	writer := bytes.NewBuffer(make([]byte, 0, 64))
	n, err := io.Copy(writer, obj)
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

// LatestVersion will return the URL Last-Modified time
func (s *MinIOSource) LatestVersion() (*Version, error) {
	info, err := s.client.StatObject(context.Background(), s.bucketName, s.baseURL, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &Version{Date: info.LastModified}, nil
}
