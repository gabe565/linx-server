package handlers

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/auth/keyhash"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/util"
	"github.com/go-chi/chi/v5"
)

type AccessKeySource int

const (
	AccessKeySourceNone AccessKeySource = iota
	AccessKeySourceCookie
	AccessKeySourceHeader
	AccessKeySourceForm
	AccessKeySourceQuery
)

const (
	AccessKeyHeader = "Linx-Access-Key"
	AccessKeyParam  = "access_key"
)

//nolint:gochecknoglobals
var (
	errInvalidAccessKey = errors.New("invalid access key")
	cliUserAgents       = []string{"curl", "wget"}
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func AccessKeyFromRequest(r *http.Request, src AccessKeySource) (string, bool) {
	var key string
	switch src {
	case AccessKeySourceCookie:
		cookieKey, err := r.Cookie(AccessKeyHeader)
		if err != nil {
			return "", false
		}
		key = util.TryPathUnescape(cookieKey.Value)
	case AccessKeySourceHeader:
		key = util.TryPathUnescape(r.Header.Get(AccessKeyHeader))
	case AccessKeySourceForm:
		key = r.PostFormValue(AccessKeyParam)
	case AccessKeySourceQuery:
		key = r.URL.Query().Get(AccessKeyParam)
	default:
		return "", false
	}
	return key, key != ""
}

func CheckAccessKey(r *http.Request, metadata *backends.Metadata) (AccessKeySource, error) {
	key := metadata.AccessKey
	if key == "" {
		return AccessKeySourceNone, nil
	}

	for _, src := range []AccessKeySource{
		AccessKeySourceCookie,
		AccessKeySourceHeader,
		AccessKeySourceForm,
		AccessKeySourceQuery,
	} {
		requestKey, ok := AccessKeyFromRequest(r, src)
		if !ok {
			continue
		}

		match, err := keyhash.CheckWithFallback(key, requestKey)
		if err != nil {
			return src, err
		}
		if match {
			return src, nil
		}
		return src, errInvalidAccessKey
	}

	return AccessKeySourceNone, errInvalidAccessKey
}

func SetAccessKeyCookies(w http.ResponseWriter, r *http.Request, fileName, value string, expires time.Time) {
	u := headers.GetSiteURL(r)
	cookie := http.Cookie{
		Name:     AccessKeyHeader,
		Value:    url.PathEscape(value),
		HttpOnly: true,
		Domain:   u.Hostname(),
		Expires:  expires,
		Secure:   u.Scheme == "https",
	}

	cookie.Path = path.Join(u.Path, fileName)
	http.SetCookie(w, &cookie)

	cookie.Path = path.Join(u.Path, config.Default.SelifPath, fileName)
	http.SetCookie(w, &cookie)
}

func IsDirectUA(r *http.Request) bool {
	ua := strings.ToLower(r.Header.Get("User-Agent"))
	return !config.Default.NoDirectAgents && !strings.EqualFold(r.Header.Get("Accept"), "application/json") &&
		slices.ContainsFunc(cliUserAgents, func(s string) bool {
			return strings.Contains(ua, s)
		})
}

func FileAccessHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "name")

	if _, err := fs.Stat(assets.Static(), fileName); err == nil || !os.IsNotExist(err) {
		AssetHandler()(w, r)
		return
	}

	if IsDirectUA(r) {
		FileServeHandler(w, r)
		return
	}

	metadata, err := CheckFile(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, backends.ErrNotFound) {
			Error(w, r, http.StatusNotFound, template.WithDescription("This file has expired or does not exist."))
			return
		}
		ErrorMsg(w, r, http.StatusInternalServerError, "Corrupt metadata")
		return
	}

	src, err := CheckAccessKey(r, &metadata)
	if err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, r, fileName, "", time.Time{})
		}

		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: errInvalidAccessKey.Error()})
			return
		}

		Error(w, r, http.StatusUnauthorized)
		return
	}

	if metadata.AccessKey != "" {
		if requestKey, ok := AccessKeyFromRequest(r, src); ok {
			var expiry time.Time
			if config.Default.Auth.CookieExpiry.Duration != 0 {
				expiry = time.Now().Add(config.Default.Auth.CookieExpiry.Duration)
			}
			SetAccessKeyCookies(w, r, fileName, requestKey, expiry)
		}
	}

	FileDisplay(w, r, fileName, metadata)
}
