package repository

import (
	"io"
	"time"
)

type FileRepositoryI interface {
	UploadObject(objectName string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(objectName string, expiry time.Duration) (string, error)
}
