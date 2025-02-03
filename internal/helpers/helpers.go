package helpers

import (
	"bytes"
	"encoding/hex"
	"io"

	"gabe565.com/linx-server/internal/backends"
	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sha256-simd"
)

func GenerateMetadata(r io.Reader) (backends.Metadata, error) {
	// Since we don't have the ability to seek within a file, we can use a
	// Buffer in combination with a TeeReader to keep a copy of the bytes
	// we read when detecting the file type. These bytes are still needed
	// to hash the file and determine its size and cannot be discarded.
	var buf bytes.Buffer
	buf.Grow(512)
	var m backends.Metadata
	teeReader := io.TeeReader(r, &buf)

	// Get first 512 bytes for mimetype detection
	header := make([]byte, 512)
	headerlen, err := teeReader.Read(header)
	if err != nil {
		return m, err
	}

	// Create a Hash and a MultiReader that includes the Buffer we created
	// above along with the original Reader, which will have the rest of
	// the file.
	hasher := sha256.New()
	multiReader := io.MultiReader(&buf, r)

	// Copy everything into the Hash, then use the number of bytes written
	// as the file size.
	var readLen int64
	readLen, err = io.Copy(hasher, multiReader)
	if err != nil {
		return m, err
	}
	m.Size += readLen

	// Get the hex-encoded string version of the Hash checksum
	m.Sha256sum = hex.EncodeToString(hasher.Sum(nil))

	// Use the bytes we extracted earlier and attempt to determine the file
	// type
	kind := mimetype.Detect(header[:headerlen])
	m.Mimetype = kind.String()

	return m, err
}
