package backends

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

type StorageBackend interface {
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Head(ctx context.Context, key string) (Metadata, error)
	Get(ctx context.Context, key string) (Metadata, io.ReadCloser, error)
	Put(ctx context.Context, key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (Metadata, error)
	PutMetadata(ctx context.Context, key string, m Metadata) error
	ServeFile(key string, w http.ResponseWriter, r *http.Request) error
	Size(ctx context.Context, key string) (int64, error)
}

type MetaStorageBackend interface {
	StorageBackend
	List() ([]string, error)
}

var (
	ErrNotFound  = errors.New("file not found")
	ErrFileEmpty = errors.New("empty file")
)
