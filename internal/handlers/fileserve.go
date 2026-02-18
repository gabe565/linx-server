package handlers

import (
	"bytes"
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
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/util"
	"github.com/go-chi/chi/v5"
)

const FileCSP = "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline';"

func FileServeHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	metadata, err := CheckFile(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, backends.ErrNotFound) {
			ErrorMsg(w, r, http.StatusNotFound, "File not found")
		} else {
			slog.Error("Corrupt metadata", "path", fileName, "error", err) //nolint:gosec
			ErrorMsg(w, r, http.StatusInternalServerError, "Corrupt metadata")
		}
		return
	}

	if src, err := CheckAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, r, fileName, "", time.Time{})
		}
		Error(w, r, http.StatusUnauthorized)
		return
	}

	if !config.Default.AllowHotlink {
		referer := r.Header.Get("Referer")
		ok := referer == ""

		if !ok {
			got, _ := url.Parse(referer)
			want := headers.GetSiteURL(r)

			if ok = csrf.SameOrigin(got, want); !ok {
				for _, allowed := range config.Default.AllowReferrers {
					want, err := url.Parse(allowed)
					if err != nil {
						slog.Error("Failed to parse allowed referrer", "referrer", allowed, "error", err)
						continue
					}

					if csrf.SameOrigin(got, want) {
						ok = true
						break
					}
				}
			}
		}

		if !ok {
			http.Redirect(w, r, headers.GetFileURL(r, fileName).String(), http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", FileCSP)
	w.Header().Set("Referrer-Policy", config.Default.Header.FileReferrerPolicy)

	w.Header().Set("Content-Type", metadata.Mimetype)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("ETag", metadata.Etag())
	if metadata.AccessKey != "" || config.Default.Auth.File != "" || config.Default.Auth.RemoteFile != "" {
		w.Header().Set("Cache-Control", "private, no-cache")
	} else {
		w.Header().Set("Cache-Control", "public, no-cache")
	}

	if r.URL.Query().Has("download") || IsDirectUA(r) {
		dlName := fileName
		if metadata.OriginalName != "" {
			dlName = metadata.OriginalName
		}
		w.Header().Set("Content-Disposition", util.EncodeContentDisposition("attachment", dlName))
	}

	if err := config.StorageBackend.ServeFile(fileName, w, r); err != nil {
		slog.Error("Failed to serve file", "path", fileName, "error", err) //nolint:gosec
		Error(w, r, http.StatusInternalServerError)
		return
	}
}

func AssetHandler(opts ...template.OptionFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Accept"), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Not found"})
			return
		}

		ServeAsset(w, r, http.StatusOK, opts...)
	}
}

type StatusResponseWriter struct {
	http.ResponseWriter
	code int
}

func (m StatusResponseWriter) WriteHeader(int) {
	m.ResponseWriter.WriteHeader(m.code)
}

func ServeAsset(w http.ResponseWriter, r *http.Request, status int, opts ...template.OptionFunc) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if strings.HasPrefix(path, ".vite") {
		path = "/"
		status = http.StatusNotFound
	}

	var file io.ReadSeeker
	if asset, err := assets.Static().Open(path); err == nil {
		defer func() {
			_ = asset.Close()
		}()

		var ok bool
		file, ok = asset.(io.ReadSeeker)
		if !ok {
			slog.Error("Static asset is not a ReadSeeker", "path", path) //nolint:gosec
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
	} else {
		path = "index.html"
		var buf bytes.Buffer
		if err := template.Index(r, opts...).Render(&buf); err != nil {
			slog.Error("Failed to render index.html", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		file = bytes.NewReader(buf.Bytes())

		w.Header().Set("Cache-Control", "public, no-cache")
	}

	if status != http.StatusOK {
		w = StatusResponseWriter{w, status}
	}

	w.Header().Set("Vary", "Accept")
	http.ServeContent(w, r, path, config.TimeStarted, file)
}

func CheckFile(ctx context.Context, filename string) (backends.Metadata, error) {
	metadata, err := config.StorageBackend.Head(ctx, filename)
	if err != nil {
		return metadata, err
	}

	if metadata.Expired() {
		go func() {
			if err := config.StorageBackend.Delete(context.Background(), filename); err != nil {
				slog.Error("Failed to delete expired file", "path", filename, "error", err)
			}
		}()
		return metadata, backends.ErrNotFound
	}

	return metadata, nil
}
