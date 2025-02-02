package torrent

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/utils/bytefmt"
	"github.com/go-chi/chi/v5"
	"github.com/zeebo/bencode"
)

const PieceLength = 256 * bytefmt.KiB

type Info struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

type Torrent struct {
	Encoding string   `bencode:"encoding"`
	Info     Info     `bencode:"info"`
	URLList  []string `bencode:"url-list"`
}

func HashPiece(piece []byte) []byte {
	h := sha1.New() //nolint:gosec
	h.Write(piece)
	return h.Sum(nil)
}

func CreateTorrent(fileName string, f io.Reader, r *http.Request) ([]byte, error) {
	url := headers.GetSelifURL(r, fileName)
	chunk := make([]byte, PieceLength)

	t := Torrent{
		Encoding: "UTF-8",
		Info: Info{
			PieceLength: PieceLength,
			Name:        fileName,
		},
		URLList: []string{url.String()},
	}

	for {
		n, err := io.ReadFull(f, chunk)
		if err == io.EOF {
			break
		} else if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
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

func FileTorrentHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	metadata, f, err := config.StorageBackend.Get(fileName)
	if err != nil {
		switch {
		case errors.Is(err, backends.ErrNotFound):
			handlers.NotFound(w, r)
			return
		case errors.Is(err, backends.ErrBadMetadata):
			handlers.Oops(w, r, handlers.RespAUTO, "Corrupt metadata.")
			return
		default:
			handlers.Oops(w, r, handlers.RespAUTO, err.Error())
			return
		}
	}
	defer func() {
		_ = f.Close()
	}()

	if expiry.IsTSExpired(metadata.Expiry) {
		if err := config.StorageBackend.Delete(fileName); err != nil {
			slog.Error("Failed to delete expired file", "path", fileName)
		}
		handlers.NotFound(w, r)
		return
	}

	encoded, err := CreateTorrent(fileName, f, r)
	if err != nil {
		handlers.Oops(w, r, handlers.RespHTML, "Could not create torrent.")
		return
	}

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}
