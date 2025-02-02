package torrent

import (
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

	f, err := os.Open("../../assets/static/images/404.jpg")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	encoded, err := CreateTorrent("test.jpg", f, nil)
	require.NoError(t, err)

	require.NoError(t, bencode.DecodeBytes(encoded, &decoded))

	if decoded.Info.Pieces != "\xd6\xff\xbf'^)\x85?\xb4.\xb0\xc1|\xa3\x83\xeeX\xf9\xfd\xd7" {
		t.Fatal("Torrent pieces did not match expected pieces for image")
	}
}
