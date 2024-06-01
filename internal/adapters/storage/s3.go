package storage

import (
	"context"
	domains "github.com/WildEgor/e-shop-cdn/internal/domain"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

// DownloadCallbackHandler using for callback downloads (from service, not from s3)
type DownloadCallbackHandler = func(key string) string

// S3StorageConfig for s3
type S3StorageConfig struct {
	Endpoint string
	Region   string
	Bucket   string
	AccessId string
	Secret   string
	Secure   bool

	DownloadFn DownloadCallbackHandler
}

// S3Storage represent s3 (minio) storage
type S3Storage struct {
	config *S3StorageConfig
	client *s3.Client `wire:"-"`
}

// NewS3Storage create default s3 storage
func NewS3Storage(config *S3StorageConfig) (*S3Storage, error) {
	client, err := s3.New(config.Endpoint, &s3.Options{
		Creds:  credentials.NewStaticV4(config.AccessId, config.Secret, ""),
		Secure: config.Secure,
	})
	if err != nil {
		return nil, err
	}

	s := &S3Storage{
		client: client,
		config: config,
	}

	if err = s.Ping(); err != nil {
		return nil, err
	}

	return s, nil
}

// DownloadURL call download callback
func (s *S3Storage) DownloadURL(objectName string) string {
	return s.config.DownloadFn(objectName)
}

// Metadata check if file exists and return meta
func (s *S3Storage) Metadata(objectName string) (*domains.FileMetadata, error) {
	data, err := s.client.StatObject(context.Background(), s.config.Bucket, objectName, s3.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &domains.FileMetadata{
		Name: data.Key,
		Size: data.Size,
	}, nil
}

// Upload file to s3
func (s *S3Storage) Upload(ctx context.Context, objectName string, reader io.Reader) error {
	_, err := s.client.PutObject(ctx, s.config.Bucket, objectName, reader, -1, s3.PutObjectOptions{})
	return err
}

// Download from s3
func (s *S3Storage) Download(ctx context.Context, objectName string) ([]byte, error) {
	reader, err := s.client.GetObject(ctx, s.config.Bucket, objectName, s3.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Delete from s3
func (s *S3Storage) Delete(objectName string) error {
	return s.client.RemoveObject(
		context.Background(),
		s.config.Bucket, objectName,
		s3.RemoveObjectOptions{},
	)
}

// Exists check if exists
func (s *S3Storage) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.config.Bucket, objectName, s3.StatObjectOptions{})
	if err != nil {
		// If the error is due to the object not found, return false with no error
		if errResp := s3.ToErrorResponse(err); errResp.Code == "NoSuchKey" {
			return false, nil
		}
		// For other errors, return them
		return false, err
	}
	// If there is no error, the object exists
	return true, nil
}

// Ping s3 bucket
func (s *S3Storage) Ping() error {
	_, err := s.client.BucketExists(context.Background(), s.config.Bucket)

	// TODO: create bucket

	return err
}
