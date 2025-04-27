package s3

import (
	"encoding/json"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"github.com/minio/minio-go/v7"
)

func mapMetadata(m backends.Metadata) map[string]string {
	mapped := make(map[string]string, 4)
	if m.DeleteKey != "" {
		mapped["deletekey"] = m.DeleteKey
	}
	if m.AccessKey != "" {
		mapped["accesskey"] = m.AccessKey
	}
	if m.Sha256sum != "" {
		mapped["sha256sum"] = m.Sha256sum
	}
	if !m.Expiry.IsZero() {
		mapped["expiry"] = m.Expiry.Format(time.RFC3339)
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
		case "deletekey", "delete_key":
			m.DeleteKey = v
		case "accesskey":
			m.AccessKey = v
		case "sha256sum":
			m.Sha256sum = v
		case "mimetype":
			m.Mimetype = v
		case "expiry":
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
