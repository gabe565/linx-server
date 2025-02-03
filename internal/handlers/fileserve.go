package handlers

import (
	"context"
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
	"gabe565.com/linx-server/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func FileServeHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	metadata, err := CheckFile(r.Context(), fileName)
	if errors.Is(err, backends.ErrNotFound) {
		AssetHandler(w, r)
		return
	} else if err != nil {
		Oops(w, r, RespAUTO, "Corrupt metadata.")
		return
	}

	if src, err := CheckAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, r, fileName, "", time.Unix(0, 0))
		}
		Unauthorized(w, r)

		return
	}

	if !config.Default.AllowHotlink {
		referer := r.Header.Get("Referer")
		u, _ := url.Parse(referer)
		p := headers.GetSiteURL(r)
		if referer != "" && !csrf.SameOrigin(u, p) {
			http.Redirect(w, r, headers.GetFileURL(r, fileName).String(), http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", config.Default.FileContentSecurityPolicy)
	w.Header().Set("Referrer-Policy", config.Default.FileReferrerPolicy)

	w.Header().Set("Content-Type", metadata.Mimetype)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("Etag", strconv.Quote(metadata.Sha256sum))
	w.Header().Set("Cache-Control", "public, no-cache")

	modtime := time.Unix(0, 0)
	if done := httputil.CheckPreconditions(w, r, modtime); done {
		return
	}

	if r.Method != http.MethodHead {
		if err := config.StorageBackend.ServeFile(fileName, w, r); err != nil {
			Oops(w, r, RespAUTO, err.Error())
			return
		}
	}
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/favicon.ico" {
		path = "/images/favicon.gif"
	}

	file, err := assets.Static().Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		NotFound(w, r)
		return
	}

	w.Header().Set("Etag", strconv.Quote(config.TimeStartedStr))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, path, config.TimeStarted, file.(io.ReadSeeker))
}

func CheckFile(ctx context.Context, filename string) (backends.Metadata, error) {
	metadata, err := config.StorageBackend.Head(ctx, filename)
	if err != nil {
		return metadata, err
	}

	if expiry.IsTSExpired(metadata.Expiry) {
		if err := config.StorageBackend.Delete(ctx, filename); err != nil {
			slog.Error("Failed to delete expired file", "path", filename)
		}
		return metadata, backends.ErrNotFound
	}

	return metadata, nil
}
