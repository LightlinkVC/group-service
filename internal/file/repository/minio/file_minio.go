package minio

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v6"
)

type FileRepository struct {
	client     *minio.Client
	bucketName string
}

func NewFileRepository(
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	bucketName string,
	useSSL bool,
) (*FileRepository, error) {
	client, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	exists, err := client.BucketExists(bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(bucketName, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &FileRepository{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (r *FileRepository) UploadObject(
	objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := r.client.PutObjectWithContext(
		context.Background(),
		r.bucketName,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func (r *FileRepository) GetPresignedURL(
	objectName string,
	expiry time.Duration,
) (string, error) {
	presignedURL, err := r.client.PresignedGetObject(
		r.bucketName,
		objectName,
		expiry,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return strings.Replace(presignedURL.String(), "http://group-service-minio:9000", "http://localhost/minio", 1), nil
}
