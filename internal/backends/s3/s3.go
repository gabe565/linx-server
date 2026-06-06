package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
	"gabe565.com/linx-server/internal/util"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	_ backends.ListBackend      = Backend{}
	_ backends.PresignedBackend = Backend{}
)

type Backend struct {
	bucket          string
	client          *minio.Client
	presignedExpiry time.Duration
}

func (b Backend) Delete(ctx context.Context, key string) error {
	return b.client.RemoveObject(ctx, b.bucket, key, minio.RemoveObjectOptions{})
}

func (b Backend) Exists(ctx context.Context, key string) (bool, error) {
	_, err := b.client.StatObject(ctx, b.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b Backend) Head(ctx context.Context, key string) (backends.Metadata, error) {
	info, err := b.client.StatObject(ctx, b.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return backends.Metadata{}, backends.ErrNotFound
		}
		return backends.Metadata{}, err
	}

	return unmapMetadata(info)
}

func (b Backend) Get(ctx context.Context, key string) (backends.Metadata, io.ReadCloser, error) {
	obj, err := b.client.GetObject(ctx, b.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return backends.Metadata{}, nil, backends.ErrNotFound
		}
		return backends.Metadata{}, nil, err
	}

	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return backends.Metadata{}, nil, err
	}

	m, err := unmapMetadata(info)
	if err != nil {
		_ = obj.Close()
		return backends.Metadata{}, nil, err
	}

	return m, obj, nil
}

func (b Backend) GetPresignedURL(ctx context.Context, key, contentDisposition string) (*url.URL, error) {
	meta, err := b.Head(ctx, key)
	if err != nil {
		return nil, err
	}

	if meta.Salt == "" {
		switch {
		case meta.AccessKey != "":
			return nil, fmt.Errorf("%w: legacy access key", backends.ErrPresignedUnsupported)
		case meta.DeleteKey != "":
			return nil, fmt.Errorf("%w: legacy delete key", backends.ErrPresignedUnsupported)
		}
	}

	var reqParams url.Values
	if contentDisposition != "" {
		reqParams = url.Values{"response-content-disposition": []string{contentDisposition}}
	}

	expires := b.presignedExpiry
	if !meta.Expiry.IsZero() {
		expires = min(expires, time.Until(meta.Expiry))
	}

	u, err := b.client.PresignedGetObject(ctx, b.bucket, key, expires, reqParams)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (b Backend) ServeFile(key string, w http.ResponseWriter, r *http.Request) error {
	obj, err := b.client.GetObject(r.Context(), b.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return backends.ErrNotFound
		}
		return err
	}
	defer func() {
		_ = obj.Close()
	}()

	var mod time.Time
	if stat, err := obj.Stat(); err == nil {
		mod = stat.LastModified
	}

	http.ServeContent(w, r, key, mod, obj)
	return nil
}

func (b Backend) Put(
	ctx context.Context,
	r io.Reader,
	key string,
	size int64,
	opts backends.PutOptions,
) (backends.Metadata, error) {
	var m backends.Metadata

	mime, r, err := helpers.DetectMimetype(r)
	if err != nil {
		return m, err
	}

	if size == 0 {
		size = -1
	}

	m = backends.Metadata{
		OriginalName: opts.OriginalName,
		DeleteKey:    opts.DeleteKey,
		AccessKey:    opts.AccessKey,
		Salt:         opts.Salt,
		Mimetype:     mime.String(),
		Expiry:       opts.Expiry,
	}

	info, err := b.client.PutObject(ctx, b.bucket, key, r, size, minio.PutObjectOptions{
		ContentType:        m.Mimetype,
		ContentDisposition: util.EncodeContentDisposition("attachment", m.OriginalName),
		UserMetadata:       mapMetadata(m),
	})
	if err != nil {
		if _, ok := errors.AsType[*url.Error](err); ok && strings.Contains(err.Error(), "ContentLength=") &&
			strings.Contains(err.Error(), "with Body length") {
			err = fmt.Errorf("%w: %w", backends.ErrSizeMismatch, err)
		}
	}

	m.Size = info.Size
	m.Checksum = info.ETag

	return m, err
}

func (b Backend) PutMetadata(ctx context.Context, key string, m backends.Metadata) error {
	src := minio.CopySrcOptions{Bucket: b.bucket, Object: key}
	dst := minio.CopyDestOptions{
		Bucket:          b.bucket,
		Object:          key,
		ReplaceMetadata: true,
		UserMetadata:    mapMetadata(m),
	}
	_, err := b.client.CopyObject(ctx, dst, src)
	return err
}

func (b Backend) Size(ctx context.Context, key string) (int64, error) {
	info, err := b.client.StatObject(ctx, b.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (b Backend) List(ctx context.Context) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for item := range b.client.ListObjectsIter(ctx, b.bucket, minio.ListObjectsOptions{Recursive: true}) {
			if item.Err != nil {
				yield("", item.Err)
				return
			}

			if !yield(item.Key, nil) {
				return
			}
		}
	}
}

func (b Backend) PresignedOrigin(ctx context.Context) (string, error) {
	u, err := b.client.PresignedGetObject(ctx, b.bucket, "_probe", time.Minute, nil)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}

func New(
	_ context.Context,
	bucket, region, endpoint string,
	forcePathStyle bool,
	presignedExpiry time.Duration,
) (Backend, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return Backend{}, err
	}

	opts := &minio.Options{
		Creds: credentials.NewChainCredentials([]credentials.Provider{
			&credentials.EnvAWS{},
			&credentials.IAM{},
		}),
		Secure: u.Scheme == "https",
		Region: region,
	}
	if forcePathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(u.Host, opts)
	if err != nil {
		return Backend{}, err
	}
	return Backend{bucket: bucket, client: client, presignedExpiry: presignedExpiry}, nil
}
