package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PresignedGetURL(ctx context.Context, key string, expirySeconds int) (string, error)
}
