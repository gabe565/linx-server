package backends

import (
	"context"
	"errors"
	"io"
	"iter"
	"net/http"
	"time"
)

type StorageBackend interface {
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Head(ctx context.Context, key string) (Metadata, error)
	Get(ctx context.Context, key string) (Metadata, io.ReadCloser, error)
	Put(ctx context.Context, r io.Reader, key string, opts PutOptions) (Metadata, error)
	PutMetadata(ctx context.Context, key string, m Metadata) error
	ServeFile(key string, w http.ResponseWriter, r *http.Request) error
	Size(ctx context.Context, key string) (int64, error)
}

type PutOptions struct {
	OriginalName string
	Expiry       time.Time
	DeleteKey    string
	AccessKey    string
}

type ListBackend interface {
	StorageBackend
	List(ctx context.Context) iter.Seq2[string, error]
}

var (
	ErrNotFound  = errors.New("file not found")
	ErrFileEmpty = errors.New("empty file")
)
