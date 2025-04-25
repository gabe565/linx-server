package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/util"
)

type DisplayJSON struct {
	Filename     string   `json:"filename"`
	DirectURL    string   `json:"direct_url"`
	TorrentURL   string   `json:"torrent_url,omitzero"`
	Expiry       string   `json:"expiry"`
	Size         string   `json:"size"`
	Mimetype     string   `json:"mimetype"`
	Sha256sum    string   `json:"sha256sum"`
	Language     string   `json:"language,omitzero"`
	ArchiveFiles []string `json:"archive_files,omitzero"`
}

func FileDisplay(w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		res := DisplayJSON{
			Filename:     fileName,
			DirectURL:    headers.GetSelifURL(r, fileName).String(),
			Expiry:       strconv.FormatInt(max(metadata.Expiry.Unix(), 0), 10),
			Size:         strconv.FormatInt(metadata.Size, 10),
			Mimetype:     metadata.Mimetype,
			Sha256sum:    metadata.Sha256sum,
			ArchiveFiles: metadata.ArchiveFiles,
		}

		extension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		if strings.HasPrefix(metadata.Mimetype, "text/") || util.SupportedBinExtension(extension) {
			res.Language = util.ExtensionToHlLang(fileName, extension)
		}

		if !config.Default.NoTorrent {
			res.TorrentURL = headers.GetTorrentURL(r, fileName).String()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	AssetHandler(w, r)
}
