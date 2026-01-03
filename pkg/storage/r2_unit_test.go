package storage

import (
	"context"
	"io"
	"testing"
	"time"
)

// mockS3ClientInterface defines the interface that R2Storage methods use.
// This allows us to create mock implementations for testing.
type mockS3ClientInterface interface {
	PutObject(ctx context.Context, params *putObjectInput, optFns ...func(*s3Options)) (*s3PutObjectOutput, error)
	GetObject(ctx context.Context, params *getObjectInput, optFns ...func(*s3Options)) (*s3GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *deleteObjectInput, optFns ...func(*s3Options)) (*s3DeleteObjectOutput, error)
	PresignGetObject(ctx context.Context, params *getObjectInput, optFns ...func(*presignOptions)) (*presignedGetObjectOutput, error)
}

// s3Options is a minimal interface matching s3.Options
type s3Options struct{}

// presignOptions is a minimal interface matching s3.PresignOptions
type presignOptions struct{}

type putObjectInput struct {
	Bucket *string
	Key    *string
	Body   io.Reader
}

type s3PutObjectOutput struct{}

type getObjectInput struct {
	Bucket *string
	Key    *string
}

type s3GetObjectOutput struct {
	Body io.ReadCloser
}

type deleteObjectInput struct {
	Bucket *string
	Key    *string
}

type s3DeleteObjectOutput struct{}

type presignedGetObjectOutput struct {
	URL string
}

// mockS3Client is a mock implementation for testing R2Storage methods.
type mockS3Client struct {
	putObjectFunc    func(ctx context.Context, params *putObjectInput) (*s3PutObjectOutput, error)
	getObjectFunc    func(ctx context.Context, params *getObjectInput) (*s3GetObjectOutput, error)
	deleteObjectFunc func(ctx context.Context, params *deleteObjectInput) (*s3DeleteObjectOutput, error)
	presignFunc      func(ctx context.Context, params *getObjectInput, expires time.Duration) (*presignedGetObjectOutput, error)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *putObjectInput, optFns ...func(*s3Options)) (*s3PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params)
	}
	return &s3PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *getObjectInput, optFns ...func(*s3Options)) (*s3GetObjectOutput, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, params)
	}
	return &s3GetObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *deleteObjectInput, optFns ...func(*s3Options)) (*s3DeleteObjectOutput, error) {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, params)
	}
	return &s3DeleteObjectOutput{}, nil
}

func (m *mockS3Client) PresignGetObject(ctx context.Context, params *getObjectInput, optFns ...func(*presignOptions)) (*presignedGetObjectOutput, error) {
	if m.presignFunc != nil {
		return m.presignFunc(ctx, params, 0)
	}
	return &presignedGetObjectOutput{
		URL: "https://mock-presigned-url.example.com",
	}, nil
}

func TestR2Storage_UploadFile_EmptyBucketName(t *testing.T) {
	t.Parallel()

	// Create a storage with empty bucket name
	// The methods should return an error before attempting to use the client
	storage := &R2Storage{
		client:     nil, // Won't be used since bucket is empty
		bucketName: "",
	}

	_, err := storage.UploadFile(context.Background(), "test.txt", nil)

	if err == nil {
		t.Error("UploadFile() expected error for empty bucket name, got nil")
	}
}

func TestR2Storage_DownloadFile_EmptyBucketName(t *testing.T) {
	t.Parallel()

	storage := &R2Storage{
		client:     nil,
		bucketName: "",
	}

	_, err := storage.DownloadFile(context.Background(), "test.txt")

	if err == nil {
		t.Error("DownloadFile() expected error for empty bucket name, got nil")
	}
}

func TestR2Storage_DeleteFile_EmptyBucketName(t *testing.T) {
	t.Parallel()

	storage := &R2Storage{
		client:     nil,
		bucketName: "",
	}

	err := storage.DeleteFile(context.Background(), "test.txt")

	if err == nil {
		t.Error("DeleteFile() expected error for empty bucket name, got nil")
	}
}

func TestR2Storage_GetSignedURL_EmptyBucketName(t *testing.T) {
	t.Parallel()

	storage := &R2Storage{
		client:     nil,
		bucketName: "",
	}

	_, err := storage.GetSignedURL(context.Background(), "test.txt", time.Hour)

	if err == nil {
		t.Error("GetSignedURL() expected error for empty bucket name, got nil")
	}
}
