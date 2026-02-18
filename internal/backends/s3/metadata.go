package s3

import (
	"encoding/json"
	"mime"
	"net/url"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/util"
	"github.com/minio/minio-go/v7"
)

const (
	DeleteKey = "deletekey"
	AccessKey = "accesskey"
	Salt      = "salt"
	Expiry    = "expiry"
)

func mapMetadata(m backends.Metadata) map[string]string {
	mapped := make(map[string]string)
	if m.DeleteKey != "" {
		mapped[DeleteKey] = url.QueryEscape(m.DeleteKey)
	}
	if m.AccessKey != "" {
		mapped[AccessKey] = url.QueryEscape(m.AccessKey)
	}
	if m.Salt != "" {
		mapped[Salt] = url.QueryEscape(m.Salt)
	}
	if !m.Expiry.IsZero() {
		mapped[Expiry] = m.Expiry.Format(time.RFC3339)
	}
	return mapped
}

func unmapMetadata(info minio.ObjectInfo) (backends.Metadata, error) {
	m := backends.Metadata{
		Checksum: info.ETag,
		Mimetype: info.ContentType,
		Size:     info.Size,
		ModTime:  info.LastModified,
	}
	if v := info.Metadata.Get("Content-Disposition"); v != "" {
		_, parsed, err := mime.ParseMediaType(v)
		if err == nil {
			if v, ok := parsed["filename"]; ok {
				m.OriginalName = v
			}
		}
	}
	for k, v := range info.UserMetadata {
		k = strings.ToLower(k)
		switch k {
		case "originalname":
			m.OriginalName = util.TryQueryUnescape(v)
		case DeleteKey, "delete_key":
			m.DeleteKey = util.TryQueryUnescape(v)
		case AccessKey:
			m.AccessKey = util.TryQueryUnescape(v)
		case Salt:
			m.Salt = util.TryQueryUnescape(v)
		case "sha256sum":
			m.Checksum = v
		case "mimetype":
			m.Mimetype = v
		case Expiry:
			b, err := json.Marshal(v)
			if err != nil {
				return m, err
			}

			var expiry backends.Expiry
			if err := expiry.UnmarshalJSON(b); err != nil {
				return m, err
			}

			m.Expiry = time.Time(expiry)
		}
	}
	return m, nil
}
