package torrent

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"testing"

	"gabe565.com/linx-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/bencode"
)

func TestCreateTorrent(t *testing.T) {
	var decoded Torrent

	tmp := t.TempDir()
	tmpFile := filepath.Join(tmp, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0o600)
	require.NoError(t, err)

	f, err := os.Open(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	encoded, err := CreateTorrent(filepath.Base(tmpFile), f, nil)
	require.NoError(t, err)

	require.NoError(t, bencode.DecodeBytes(encoded, &decoded))

	assert.Equal(t, "UTF-8", decoded.Encoding)
	assert.Equal(t, filepath.Base(tmpFile), decoded.Info.Name)
	assert.NotZero(t, decoded.Info.PieceLength, "expected a piece length")
	assert.NotEmpty(t, decoded.Info.Pieces, "expected at least one piece")
	assert.NotZero(t, decoded.Info.Length, "invalid length")

	tracker := config.Default.SiteURL.URL
	tracker.Path = path.Join(tracker.Path, config.Default.SelifPath, filepath.Base(tmpFile))
	assert.Equal(t, tracker.String(), decoded.URLList[0])
}

func TestCreateTorrentWithImage(t *testing.T) {
	var decoded Torrent

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	encoded, err := CreateTorrent("test.png", &buf, nil)
	require.NoError(t, err)

	require.NoError(t, bencode.DecodeBytes(encoded, &decoded))
	assert.Equal(t, "\x1f?\xe6a#\xe3wIi\xf5}\xf2\x87X\x89\r\xf8t\xdc\xc0", decoded.Info.Pieces)
}
