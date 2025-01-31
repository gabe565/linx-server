package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andreimarcu/linx-server/internal/backends"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/expiry"
	"github.com/andreimarcu/linx-server/internal/handlers"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/zeebo/bencode"
	"github.com/zenazn/goji/web"
)

const (
	TORRENT_PIECE_LENGTH = 262144
)

type TorrentInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

type Torrent struct {
	Encoding string      `bencode:"encoding"`
	Info     TorrentInfo `bencode:"info"`
	UrlList  []string    `bencode:"url-list"`
}

func HashPiece(piece []byte) []byte {
	h := sha1.New()
	h.Write(piece)
	return h.Sum(nil)
}

func CreateTorrent(fileName string, f io.Reader, r *http.Request) ([]byte, error) {
	url := headers.GetSiteURL(r) + config.Default.SelifPath + fileName
	chunk := make([]byte, TORRENT_PIECE_LENGTH)

	t := Torrent{
		Encoding: "UTF-8",
		Info: TorrentInfo{
			PieceLength: TORRENT_PIECE_LENGTH,
			Name:        fileName,
		},
		UrlList: []string{url},
	}

	for {
		n, err := io.ReadFull(f, chunk)
		if err == io.EOF {
			break
		} else if err != nil && err != io.ErrUnexpectedEOF {
			return []byte{}, err
		}

		t.Info.Length += n
		t.Info.Pieces += string(HashPiece(chunk[:n]))
	}

	data, err := bencode.EncodeBytes(&t)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func FileTorrentHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	metadata, f, err := config.StorageBackend.Get(fileName)
	if err == backends.NotFoundErr {
		handlers.NotFound(c, w, r)
		return
	} else if err == backends.BadMetadata {
		handlers.Oops(c, w, r, handlers.RespAUTO, "Corrupt metadata.")
		return
	} else if err != nil {
		handlers.Oops(c, w, r, handlers.RespAUTO, err.Error())
		return
	}
	defer f.Close()

	if expiry.IsTsExpired(metadata.Expiry) {
		config.StorageBackend.Delete(fileName)
		handlers.NotFound(c, w, r)
		return
	}

	encoded, err := CreateTorrent(fileName, f, r)
	if err != nil {
		handlers.Oops(c, w, r, handlers.RespHTML, "Could not create torrent.")
		return
	}

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}
