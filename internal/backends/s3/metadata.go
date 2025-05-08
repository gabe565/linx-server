package s3

import (
	"encoding/json"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"github.com/minio/minio-go/v7"
)

const (
	OriginalName = "originalname"
	DeleteKey    = "deletekey"
	AccessKey    = "accesskey"
	Sha256sum    = "sha256sum"
	Mimetype     = "mimetype"
	Expiry       = "expiry"
)

func mapMetadata(m backends.Metadata) map[string]string {
	mapped := make(map[string]string, 4)
	if m.OriginalName != "" {
		mapped[OriginalName] = m.OriginalName
	}
	if m.DeleteKey != "" {
		mapped[DeleteKey] = m.DeleteKey
	}
	if m.AccessKey != "" {
		mapped[AccessKey] = m.AccessKey
	}
	if m.Sha256sum != "" {
		mapped[Sha256sum] = m.Sha256sum
	}
	if !m.Expiry.IsZero() {
		mapped[Expiry] = m.Expiry.Format(time.RFC3339)
	}
	return mapped
}

func unmapMetadata(info minio.ObjectInfo) (backends.Metadata, error) {
	m := backends.Metadata{
		Mimetype: info.ContentType,
		Size:     info.Size,
		ModTime:  info.LastModified,
	}
	for k, v := range info.UserMetadata {
		k = strings.ToLower(k)
		switch k {
		case OriginalName:
			m.OriginalName = v
		case DeleteKey, "delete_key":
			m.DeleteKey = v
		case AccessKey:
			m.AccessKey = v
		case Sha256sum:
			m.Sha256sum = v
		case Mimetype:
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
