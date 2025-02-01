package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/andreimarcu/linx-server/assets"
	"github.com/andreimarcu/linx-server/internal/backends"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/csrf"
	"github.com/andreimarcu/linx-server/internal/expiry"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/andreimarcu/linx-server/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func FileServeHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	metadata, err := CheckFile(fileName)
	if err == backends.NotFoundErr {
		NotFound(w, r)
		return
	} else if err != nil {
		Oops(w, r, RespAUTO, "Corrupt metadata.")
		return
	}

	if src, err := CheckAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, headers.GetSiteURL(r), fileName, "", time.Unix(0, 0))
		}
		Unauthorized(w, r)

		return
	}

	if !config.Default.AllowHotlink {
		referer := r.Header.Get("Referer")
		u, _ := url.Parse(referer)
		p, _ := url.Parse(headers.GetSiteURL(r))
		if referer != "" && !csrf.SameOrigin(u, p) {
			http.Redirect(w, r, config.Default.SitePath+fileName, 303)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", config.Default.FileContentSecurityPolicy)
	w.Header().Set("Referrer-Policy", config.Default.FileReferrerPolicy)

	w.Header().Set("Content-Type", metadata.Mimetype)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("Etag", fmt.Sprintf("\"%s\"", metadata.Sha256sum))
	w.Header().Set("Cache-Control", "public, no-cache")

	modtime := time.Unix(0, 0)
	if done := httputil.CheckPreconditions(w, r, modtime); done == true {
		return
	}

	if r.Method != "HEAD" {

		config.StorageBackend.ServeFile(fileName, w, r)
		if err != nil {
			Oops(w, r, RespAUTO, err.Error())
			return
		}
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if path == "/favicon.ico" {
		path = "/static/images/favicon.gif"
	}

	file, err := assets.Static.Open("static/" + path)
	if err != nil {
		NotFound(w, r)
		return
	}

	w.Header().Set("Etag", fmt.Sprintf("\"%s\"", config.TimeStartedStr))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, path, config.TimeStarted, file.(io.ReadSeeker))
	return
}

func CheckFile(filename string) (metadata backends.Metadata, err error) {
	metadata, err = config.StorageBackend.Head(filename)
	if err != nil {
		return
	}

	if expiry.IsTsExpired(metadata.Expiry) {
		config.StorageBackend.Delete(filename)
		err = backends.NotFoundErr
		return
	}

	return
}
