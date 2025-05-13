package helpers

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMetadata(t *testing.T) {
	r := strings.NewReader("This is my test content")
	m, err := GenerateMetadata(r)
	require.NoError(t, err)

	assert.Equal(t, "966152d20a77e739716a625373ee15af16e8f4aec631a329a27da41c204b0171", m.Checksum)
	assert.Equal(t, "text/plain; charset=utf-8", m.Mimetype)
	assert.Equal(t, int64(23), m.Size)
}

func TestTextCharsets(t *testing.T) {
	// verify that different text encodings are detected and passed through
	orig := "This is a text string"
	utf16 := utf16.Encode([]rune(orig))
	utf16LE := make([]byte, len(utf16)*2+2)
	utf16BE := make([]byte, len(utf16)*2+2)
	utf8 := []byte(orig)
	utf16LE[0] = 0xff
	utf16LE[1] = 0xfe
	utf16BE[0] = 0xfe
	utf16BE[1] = 0xff
	for i := range utf16 {
		lsb := utf16[i] & 0xff
		msb := utf16[i] >> 8
		utf16LE[i*2+2] = byte(lsb)
		utf16LE[i*2+3] = byte(msb)
		utf16BE[i*2+2] = byte(msb)
		utf16BE[i*2+3] = byte(lsb)
	}

	testcases := []struct {
		data      []byte
		extension string
		mimetype  string
	}{
		{mimetype: "text/plain; charset=utf-8", data: utf8},
		{mimetype: "text/plain; charset=utf-16le", data: utf16LE},
		{mimetype: "text/plain; charset=utf-16be", data: utf16BE},
	}

	for _, testcase := range testcases {
		r := bytes.NewReader(testcase.data)
		m, err := GenerateMetadata(r)
		require.NoError(t, err)
		assert.Equal(t, testcase.mimetype, m.Mimetype)
	}
}
