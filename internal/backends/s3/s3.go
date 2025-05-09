package s3

import (
	"context"
	"io"
	"iter"
	"net/http"
	"net/url"
	"os"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ backends.ListBackend = Backend{}

type Backend struct {
	bucket string
	client *minio.Client
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
	originalName, key string,
	r io.Reader,
	expiry time.Time,
	deleteKey, accessKey string,
) (backends.Metadata, error) {
	var m backends.Metadata
	var err error
	seeker, ok := r.(io.ReadSeeker)
	if ok {
		if m, err = helpers.GenerateMetadata(seeker); err != nil {
			return m, err
		}
	} else {
		tmp, err := os.CreateTemp("", "linx-*-"+key)
		if err != nil {
			return backends.Metadata{}, err
		}
		defer func() {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}()

		if m, err = helpers.GenerateMetadata(io.TeeReader(r, tmp)); err != nil {
			return m, err
		}

		seeker = tmp
	}

	if m.Size == 0 {
		return m, backends.ErrFileEmpty
	}

	if _, err = seeker.Seek(0, io.SeekStart); err != nil {
		return m, err
	}

	m.OriginalName = originalName
	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.AccessKey = accessKey

	_, err = b.client.PutObject(ctx, b.bucket, key, seeker, m.Size, minio.PutObjectOptions{
		ContentType:  m.Mimetype,
		UserMetadata: mapMetadata(m),
	})
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

func New(
	_ context.Context,
	bucket, region, endpoint string,
	forcePathStyle bool,
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
	return Backend{bucket: bucket, client: client}, nil
}
