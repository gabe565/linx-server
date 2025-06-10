package util

import (
	"encoding/base64"

	"github.com/minio/sha256-simd"
)

func SubresourceIntegrity(src []byte) string {
	hash := sha256.Sum256(src)
	return "'sha256-" + base64.StdEncoding.EncodeToString(hash[:]) + "'"
}
