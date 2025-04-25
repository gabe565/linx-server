package s3

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
)

func mapMetadata(m backends.Metadata) map[string]string {
	return map[string]string{
		"expiry":    m.Expiry.Format(time.RFC3339),
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
			b, err := json.Marshal(v)
			if err != nil {
				return m, err
			}

			var expiry backends.Expiry
			if err := expiry.UnmarshalJSON(b); err != nil {
				return m, err
			}

			m.Expiry = time.Time(expiry)
		case "size":
			var err error
			m.Size, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return m, err
			}
		}
	}
	return m, nil
}
