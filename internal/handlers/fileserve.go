package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/csrf"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/headers"
	"github.com/go-chi/chi/v5"
)

func FileServeHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	metadata, err := CheckFile(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, backends.ErrNotFound) {
			ErrorMsg(w, r, http.StatusNotFound, "File not found")
		} else {
			slog.Error("Corrupt metadata", "path", fileName, "error", err)
			ErrorMsg(w, r, http.StatusInternalServerError, "Corrupt metadata")
		}
		return
	}

	if src, err := CheckAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, r, fileName, "", time.Unix(0, 0))
		}
		Error(w, r, http.StatusUnauthorized)
		return
	}

	if !config.Default.AllowHotlink {
		referer := r.Header.Get("Referer")
		u, _ := url.Parse(referer)
		frontend, _ := url.Parse(config.Default.FrontendURL)
		p := headers.GetSiteURL(r)
		if referer != "" && !csrf.SameOrigin(u, p) &&
			(config.Default.FrontendURL == "" || !csrf.SameOrigin(u, frontend)) {
			http.Redirect(w, r, headers.GetFileURL(r, fileName).String(), http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", config.Default.Header.FileContentSecurityPolicy)
	w.Header().Set("Referrer-Policy", config.Default.Header.FileReferrerPolicy)

	w.Header().Set("Content-Type", metadata.Mimetype)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("ETag", strconv.Quote(metadata.Sha256sum))
	if metadata.AccessKey != "" || config.Default.Auth.File != "" || config.Default.Auth.RemoteFile != "" {
		w.Header().Set("Cache-Control", "private, no-cache")
	} else {
		w.Header().Set("Cache-Control", "public, no-cache")
	}
	if r.URL.Query().Has("download") {
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	}

	if r.Method != http.MethodHead {
		if err := config.StorageBackend.ServeFile(fileName, w, r); err != nil {
			slog.Error("Failed to serve file", "path", fileName, "error", err)
			Error(w, r, http.StatusInternalServerError)
			return
		}
	}
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Not found"})
	}

	ServeAsset(w, r, http.StatusOK)
}

type StatusResponseWriter struct {
	http.ResponseWriter
	code int
}

func (m StatusResponseWriter) WriteHeader(int) {
	m.ResponseWriter.WriteHeader(m.code)
}

func ServeAsset(w http.ResponseWriter, r *http.Request, status int) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	file, err := assets.Static().Open(path)
	if err != nil {
		path = "index.html"
		file, err = assets.Static().Open(path)
		if err != nil {
			slog.Error("Failed to open index.html", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if status == http.StatusOK {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	} else {
		w = StatusResponseWriter{w, status}
	}

	w.Header().Set("Vary", "Accept")
	http.ServeContent(w, r, path, config.TimeStarted, file.(io.ReadSeeker)) //nolint:errcheck
}

func CheckFile(ctx context.Context, filename string) (backends.Metadata, error) {
	metadata, err := config.StorageBackend.Head(ctx, filename)
	if err != nil {
		return metadata, err
	}

	if expiry.IsTSExpired(metadata.Expiry) {
		go func() {
			if err := config.StorageBackend.Delete(context.Background(), filename); err != nil {
				slog.Error("Failed to delete expired file", "path", filename, "error", err)
			}
		}()
		return metadata, backends.ErrNotFound
	}

	return metadata, nil
}
