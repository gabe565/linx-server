package torrent

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/util"
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
		if err != nil {
			if err == io.EOF {
				break
			} else if !errors.Is(err, io.ErrUnexpectedEOF) {
				return []byte{}, err
			}
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

	metadata, f, err := config.StorageBackend.Get(r.Context(), fileName)
	if err != nil {
		switch {
		case errors.Is(err, backends.ErrNotFound):
			handlers.ErrorMsg(w, r, http.StatusNotFound, "File not found")
			return
		case errors.Is(err, backends.ErrBadMetadata):
			slog.Error("Corrupt metadata", "path", fileName, "error", err) //nolint:gosec
			handlers.ErrorMsg(w, r, http.StatusInternalServerError, "Corrupt metadata")
			return
		default:
			slog.Error("Failed to get file", "path", fileName, "error", err) //nolint:gosec
			handlers.Error(w, r, http.StatusInternalServerError)
			return
		}
	}
	defer func() {
		_ = f.Close()
	}()

	if metadata.Expired() {
		go func() {
			if err := config.StorageBackend.Delete(context.Background(), fileName); err != nil {
				slog.Error("Failed to delete expired file", "path", fileName)
			}
		}()
		handlers.ErrorMsg(w, r, http.StatusNotFound, "File not found")
		return
	}

	if metadata.OriginalName != "" {
		fileName = metadata.OriginalName
	}

	encoded, err := CreateTorrent(fileName, f, r)
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusInternalServerError, "Could not create torrent")
		return
	}

	w.Header().Set(`Content-Disposition`, util.EncodeContentDisposition("attachment", fileName+".torrent"))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}
