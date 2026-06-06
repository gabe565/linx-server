package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/util"
)

type DisplayJSON struct {
	OriginalName string   `json:"original_name,omitzero"`
	Filename     string   `json:"filename"`
	DirectURL    string   `json:"direct_url"`
	DownloadURL  string   `json:"download_url"`
	TorrentURL   string   `json:"torrent_url,omitzero"`
	Expiry       string   `json:"expiry"`
	Size         string   `json:"size"`
	Mimetype     string   `json:"mimetype"`
	Language     string   `json:"language,omitzero"`
	ArchiveFiles []string `json:"archive_files,omitzero"`
}

func FileDisplay(w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		directURL := headers.GetSelifURL(r, fileName)
		downloadURL := *directURL
		q := downloadURL.Query()
		q.Set("download", "")
		downloadURL.RawQuery = q.Encode()

		if config.Default.S3.PresignedURLs {
			if pb, ok := config.StorageBackend.(backends.PresignedBackend); ok {
				if u, err := pb.GetPresignedURL(r.Context(), fileName, ""); err == nil {
					directURL = u
				} else if !errors.Is(err, backends.ErrPresignedUnsupported) {
					slog.Error("Failed to sign direct URL", "path", fileName, "error", err)
				}

				dlName := metadata.OriginalName
				if dlName == "" {
					dlName = fileName
				}
				disposition := util.EncodeContentDisposition("attachment", dlName)

				if u, err := pb.GetPresignedURL(r.Context(), fileName, disposition); err == nil {
					downloadURL = *u
				} else if !errors.Is(
					err,
					backends.ErrPresignedUnsupported,
				) {
					slog.Error("Failed to sign download URL", "path", fileName, "error", err)
				}
			}
		}

		res := DisplayJSON{
			OriginalName: metadata.OriginalName,
			Filename:     fileName,
			DirectURL:    directURL.String(),
			DownloadURL:  downloadURL.String(),
			Expiry:       strconv.FormatInt(max(metadata.Expiry.Unix(), 0), 10),
			Size:         strconv.FormatInt(metadata.Size, 10),
			Mimetype:     metadata.Mimetype,
			Language:     util.InferLang(fileName, metadata),
			ArchiveFiles: metadata.ArchiveFiles,
		}

		if !config.Default.NoTorrent {
			res.TorrentURL = headers.GetTorrentURL(r, fileName).String()
		}

		if metadata.AccessKey != "" || config.Default.Auth.File != "" || config.Default.Auth.RemoteFile != "" {
			w.Header().Set("Cache-Control", "private, no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, no-cache")
		}
		w.Header().Set("Vary", "Accept, Linx-Delete-Key")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", metadata.Etag())

		b, _ := json.Marshal(res)
		http.ServeContent(w, r, fileName, metadata.ModTime, bytes.NewReader(b))
		return
	}

	description := "Download this file on " + config.Default.SiteName + "."
	if !metadata.Expiry.IsZero() {
		description += " Expires " + metadata.Expiry.Format("Jan 2, 2006") + "."
	}

	prettyName := metadata.OriginalName
	if metadata.OriginalName == "" {
		prettyName = fileName
	}

	AssetHandler(
		template.WithTitle(prettyName),
		template.WithDescription(description),
	)(w, r)
}
