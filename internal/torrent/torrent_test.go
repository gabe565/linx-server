package torrent

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gabe565.com/linx-server/internal/config"
	"github.com/zeebo/bencode"
)

func TestCreateTorrent(t *testing.T) {
	var decoded Torrent

	tmp := t.TempDir()
	tmpFile := filepath.Join(tmp, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = f.Close()
	})

	encoded, err := CreateTorrent(filepath.Base(tmpFile), f, nil)
	if err != nil {
		t.Fatal(err)
	}

	bencode.DecodeBytes(encoded, &decoded)

	if decoded.Encoding != "UTF-8" {
		t.Fatalf("Encoding was %s, expected UTF-8", decoded.Encoding)
	}

	if decoded.Info.Name != filepath.Base(tmpFile) {
		t.Fatalf("Name was %s, expected %s", decoded.Info.Name, filepath.Base(tmpFile))
	}

	if decoded.Info.PieceLength <= 0 {
		t.Fatal("Expected a piece length, got none")
	}

	if len(decoded.Info.Pieces) <= 0 {
		t.Fatal("Expected at least one piece, got none")
	}

	if decoded.Info.Length <= 0 {
		t.Fatal("Length was less than or equal to 0, expected more")
	}

	tracker := fmt.Sprintf("%s%s/%s", config.Default.SiteURL, config.Default.SelifPath, filepath.Base(tmpFile))
	if decoded.UrlList[0] != tracker {
		t.Fatalf("First entry in URL list was %s, expected %s", decoded.UrlList[0], tracker)
	}
}

func TestCreateTorrentWithImage(t *testing.T) {
	var decoded Torrent

	f, err := os.Open("../../assets/static/images/404.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	encoded, err := CreateTorrent("test.jpg", f, nil)
	if err != nil {
		t.Fatal(err)
	}

	bencode.DecodeBytes(encoded, &decoded)

	if decoded.Info.Pieces != "\xd6\xff\xbf'^)\x85?\xb4.\xb0\xc1|\xa3\x83\xeeX\xf9\xfd\xd7" {
		t.Fatal("Torrent pieces did not match expected pieces for image")
	}
}
