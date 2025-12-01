package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService interface {
	UploadFile(ctx context.Context, key string, body io.Reader) (string, error)
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
}

type R2Storage struct {
	client     *s3.Client
	bucketName string
}

func NewR2Storage(accountID, accessKeyID, secretAccessKey, bucketName string) (*R2Storage, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &R2Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *R2Storage) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to R2: %w", err)
	}
	return key, nil
}

func (s *R2Storage) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from R2: %w", err)
	}
	return out.Body, nil
}

func (s *R2Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}
	return nil
}

// Helper to generate a unique key
func GenerateKey(prefix, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
