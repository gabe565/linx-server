package s3

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Backend struct {
	bucket string
	svc    *s3.S3
}

func (b Backend) Delete(key string) error {
	_, err := b.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (b Backend) Exists(key string) (bool, error) {
	_, err := b.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	return err == nil, err
}

const CodeNotFound = "NotFound"

func (b Backend) Head(key string) (backends.Metadata, error) {
	var metadata backends.Metadata
	result, err := b.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok { //nolint:errorlint
			if aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == CodeNotFound {
				err = backends.ErrNotFound
			}
		}
		return metadata, err
	}

	return unmapMetadata(result.Metadata)
}

func (b Backend) Get(key string) (backends.Metadata, io.ReadCloser, error) {
	var metadata backends.Metadata
	result, err := b.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok { //nolint:errorlint
			if aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == CodeNotFound {
				err = backends.ErrNotFound
			}
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
		result, err = b.svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(key),
			Range:  aws.String(r.Header.Get("Range")),
		})

		w.WriteHeader(http.StatusPartialContent)
		w.Header().Set("Content-Range", *result.ContentRange)
		w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		result, err = b.svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(key),
		})
	}

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok { //nolint:errorlint
			if aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == CodeNotFound {
				err = backends.ErrNotFound
			}
		}
		return err
	}

	_, err = io.Copy(w, result.Body)

	return err
}

func mapMetadata(m backends.Metadata) map[string]*string {
	return map[string]*string{
		"Expiry":    aws.String(strconv.FormatInt(m.Expiry.Unix(), 10)),
		"Deletekey": aws.String(m.DeleteKey),
		"Size":      aws.String(strconv.FormatInt(m.Size, 10)),
		"Mimetype":  aws.String(m.Mimetype),
		"Sha256sum": aws.String(m.Sha256sum),
		"AccessKey": aws.String(m.AccessKey),
	}
}

func unmapMetadata(input map[string]*string) (backends.Metadata, error) {
	var m backends.Metadata
	expiry, err := strconv.ParseInt(aws.StringValue(input["Expiry"]), 10, 64)
	if err != nil {
		return m, err
	}
	m.Expiry = time.Unix(expiry, 0)

	m.Size, err = strconv.ParseInt(aws.StringValue(input["Size"]), 10, 64)
	if err != nil {
		return m, err
	}

	m.DeleteKey = aws.StringValue(input["Deletekey"])
	if m.DeleteKey == "" {
		m.DeleteKey = aws.StringValue(input["Delete_key"])
	}

	m.Mimetype = aws.StringValue(input["Mimetype"])
	m.Sha256sum = aws.StringValue(input["Sha256sum"])

	if key, ok := input["AccessKey"]; ok {
		m.AccessKey = aws.StringValue(key)
	}

	return m, nil
}

func (b Backend) Put(key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (backends.Metadata, error) {
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
	if bytes == 0 {
		return m, backends.ErrFileEmpty
	} else if err != nil {
		return m, err
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

	uploader := s3manager.NewUploaderWithClient(b.svc)
	input := &s3manager.UploadInput{
		Bucket:   aws.String(b.bucket),
		Key:      aws.String(key),
		Body:     tmpDst,
		Metadata: mapMetadata(m),
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return m, err
	}

	return m, err
}

func (b Backend) PutMetadata(key string, m backends.Metadata) error {
	_, err := b.svc.CopyObject(&s3.CopyObjectInput{
		Bucket:            aws.String(b.bucket),
		Key:               aws.String(key),
		CopySource:        aws.String("/" + b.bucket + "/" + key),
		Metadata:          mapMetadata(m),
		MetadataDirective: aws.String("REPLACE"),
	})
	return err
}

func (b Backend) Size(key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}
	result, err := b.svc.HeadObject(input)
	if err != nil {
		return 0, err
	}

	return *result.ContentLength, nil
}

func (b Backend) List() ([]string, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(b.bucket),
	}

	results, err := b.svc.ListObjects(input)
	if err != nil {
		return nil, err
	}

	output := make([]string, 0, len(results.Contents))
	for _, object := range results.Contents {
		output = append(output, *object.Key)
	}

	return output, nil
}

func NewS3Backend(bucket string, region string, endpoint string, forcePathStyle bool) Backend {
	awsConfig := &aws.Config{}
	if region != "" {
		awsConfig.Region = aws.String(region)
	}
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
	}
	if forcePathStyle {
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sess := session.Must(session.NewSession(awsConfig))
	svc := s3.New(sess)
	return Backend{bucket: bucket, svc: svc}
}
