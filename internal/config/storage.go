package config

import (
	"context"
	"fmt"
	"os"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/backends/localfs"
	"gabe565.com/linx-server/internal/backends/s3"
)

func (c *Config) NewStorageBackend(ctx context.Context) (backends.StorageBackend, error) { //nolint:ireturn
	if c.S3.Bucket != "" {
		return c.NewS3Backend(ctx)
	}
	return c.NewLocalBackend()
}

func (c *Config) NewS3Backend(ctx context.Context) (s3.Backend, error) {
	return s3.New(ctx, c.S3.Bucket, c.S3.Region, c.S3.Endpoint, c.S3.ForcePathStyle)
}

func (c *Config) NewLocalBackend() (localfs.Backend, error) {
	err := os.MkdirAll(c.FilesPath, 0o755)
	if err != nil {
		return localfs.Backend{}, fmt.Errorf("could not create files directory: %w", err)
	}

	err = os.MkdirAll(c.MetaPath, 0o700)
	if err != nil {
		return localfs.Backend{}, fmt.Errorf("could not create metadata directory: %w", err)
	}

	return localfs.New(c.MetaPath, c.FilesPath), nil
}
