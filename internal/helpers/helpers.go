package helpers

import (
	"encoding/hex"
	"fmt"
	"io"

	"gabe565.com/linx-server/internal/backends"
	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sha256-simd"
)

func GenerateMetadata(r io.Reader) (backends.Metadata, error) {
	hasher := sha256.New()
	sz := &sizeWriter{w: hasher}

	kind, err := mimetype.DetectReader(io.TeeReader(r, sz))
	if err != nil {
		return backends.Metadata{}, fmt.Errorf("detecting mimetype: %w", err)
	}

	if _, err := io.Copy(sz, r); err != nil {
		return backends.Metadata{}, fmt.Errorf("hashing data: %w", err)
	}

	return backends.Metadata{
		Size:      sz.size,
		Sha256sum: hex.EncodeToString(hasher.Sum(nil)),
		Mimetype:  kind.String(),
	}, nil
}

type sizeWriter struct {
	w    io.Writer
	size int64
}

func (s *sizeWriter) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	s.size += int64(n)
	return n, err
}
