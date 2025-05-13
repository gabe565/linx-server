package helpers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/utils/bytefmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sha256-simd"
)

func DetectMimetype(r io.Reader) (*mimetype.MIME, io.Reader, error) {
	if seeker, ok := r.(io.Seeker); ok {
		if n, err := seeker.Seek(0, io.SeekEnd); err != nil {
			return nil, r, err
		} else if n == 0 {
			return nil, r, backends.ErrFileEmpty
		}

		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, r, err
		}

		kind, err := mimetype.DetectReader(r)
		if err != nil {
			return nil, r, err
		}

		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, r, err
		}

		return kind, r, nil
	}

	var buf bytes.Buffer
	buf.Grow(3 * bytefmt.KiB)

	kind, err := mimetype.DetectReader(io.TeeReader(r, &buf))
	if err == nil && buf.Len() == 0 {
		err = backends.ErrFileEmpty
	} else {
		r = io.MultiReader(bytes.NewReader(buf.Bytes()), r)
	}
	return kind, r, err
}

func GenerateMetadata(r io.Reader) (backends.Metadata, error) {
	mime, r, err := DetectMimetype(r)
	if err != nil {
		return backends.Metadata{}, fmt.Errorf("detecting mimetype: %w", err)
	}

	hasher := sha256.New()
	n, err := io.Copy(hasher, r)
	if err != nil {
		return backends.Metadata{}, fmt.Errorf("hashing data: %w", err)
	}

	return backends.Metadata{
		Size:      n,
		Sha256sum: hex.EncodeToString(hasher.Sum(nil)),
		Mimetype:  mime.String(),
	}, nil
}
