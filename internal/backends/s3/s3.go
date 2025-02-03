package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Backend struct {
	bucket string
	client *s3.Client
}

func (b Backend) Delete(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (b Backend) Exists(ctx context.Context, key string) (bool, error) {
	_, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	exists := true
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			exists = false
			err = nil
		}
	}
	return exists, err
}

func (b Backend) Head(ctx context.Context, key string) (backends.Metadata, error) {
	var metadata backends.Metadata
	result, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			err = backends.ErrNotFound
		}
		return metadata, err
	}

	return unmapMetadata(result.Metadata)
}

func (b Backend) Get(ctx context.Context, key string) (backends.Metadata, io.ReadCloser, error) {
	var metadata backends.Metadata
	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			err = backends.ErrNotFound
		}
		return metadata, nil, err
	}

	if metadata, err = unmapMetadata(result.Metadata); err != nil {
		return metadata, nil, err
	}
	return metadata, result.Body, nil
}

func (b Backend) ServeFile(key string, w http.ResponseWriter, r *http.Request) error {
	var result *s3.GetObjectOutput
	var err error

	if r.Header.Get("Range") != "" {
		result, err = b.client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(key),
			Range:  aws.String(r.Header.Get("Range")),
		})
		if err != nil {
			var nf *types.NotFound
			if errors.As(err, &nf) {
				err = backends.ErrNotFound
			}
			return err
		}
		defer func() {
			_ = result.Body.Close()
		}()

		w.WriteHeader(http.StatusPartialContent)
		w.Header().Set("Content-Range", *result.ContentRange)
		w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		result, err = b.client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			var nf *types.NotFound
			if errors.As(err, &nf) {
				err = backends.ErrNotFound
			}
			return err
		}
		defer func() {
			_ = result.Body.Close()
		}()
	}

	_, err = io.Copy(w, result.Body)
	return err
}

func mapMetadata(m backends.Metadata) map[string]string {
	return map[string]string{
		"expiry":    strconv.FormatInt(m.Expiry.Unix(), 10),
		"deletekey": m.DeleteKey,
		"size":      strconv.FormatInt(m.Size, 10),
		"mimetype":  m.Mimetype,
		"sha256sum": m.Sha256sum,
		"accesskey": m.AccessKey,
	}
}

func unmapMetadata(input map[string]string) (backends.Metadata, error) {
	var m backends.Metadata
	for k, v := range input {
		k = strings.ToLower(k)
		switch k {
		case "deletekey", "delete_key":
			m.DeleteKey = v
		case "accesskey":
			m.AccessKey = v
		case "sha256sum":
			m.Sha256sum = v
		case "mimetype":
			m.Mimetype = v
		case "expiry":
			expiry, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return m, err
			}
			m.Expiry = time.Unix(expiry, 0)
		case "size":
			var err error
			m.Size, err = strconv.ParseInt(input["size"], 10, 64)
			if err != nil {
				return m, err
			}
		}
	}
	return m, nil
}

func (b Backend) Put(ctx context.Context, key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (backends.Metadata, error) {
	var m backends.Metadata
	tmpDst, err := os.CreateTemp("", "linx-server-upload")
	if err != nil {
		return m, err
	}
	defer func() {
		_ = tmpDst.Close()
		_ = os.Remove(tmpDst.Name())
	}()

	bytes, err := io.Copy(tmpDst, r)
	if err != nil {
		return m, err
	}
	if bytes == 0 {
		return m, backends.ErrFileEmpty
	}

	_, err = tmpDst.Seek(0, 0)
	if err != nil {
		return m, err
	}

	m, err = helpers.GenerateMetadata(tmpDst)
	if err != nil {
		return m, err
	}
	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.AccessKey = accessKey
	// XXX: we may not be able to write this to AWS easily
	// m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, tmpDst)

	_, err = tmpDst.Seek(0, 0)
	if err != nil {
		return m, err
	}

	_, err = manager.NewUploader(b.client).Upload(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(b.bucket),
		Key:      aws.String(key),
		Body:     tmpDst,
		Metadata: mapMetadata(m),
	})
	return m, err
}

func (b Backend) PutMetadata(ctx context.Context, key string, m backends.Metadata) error {
	_, err := b.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            aws.String(b.bucket),
		Key:               aws.String(key),
		CopySource:        aws.String("/" + b.bucket + "/" + key),
		Metadata:          mapMetadata(m),
		MetadataDirective: types.MetadataDirectiveReplace,
	})
	return err
}

func (b Backend) Size(ctx context.Context, key string) (int64, error) {
	result, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}
	return *result.ContentLength, nil
}

func (b Backend) List(ctx context.Context) ([]string, error) {
	results, err := b.client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(b.bucket),
	})
	if err != nil {
		return nil, err
	}

	output := make([]string, 0, len(results.Contents))
	for _, object := range results.Contents {
		output = append(output, *object.Key)
	}

	return output, nil
}

func NewS3Backend(ctx context.Context, bucket string, region string, endpoint string, forcePathStyle bool) (Backend, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return Backend{}, err
	}
	if region != "" {
		cfg.Region = region
	}
	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
	})
	return Backend{bucket: bucket, client: client}, nil
}
