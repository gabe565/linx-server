package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
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
	HeaderName = "Linx-Access-Key"
	ParamName  = "access_key"
)

//nolint:gochecknoglobals
var (
	errInvalidAccessKey = errors.New("invalid access key")
	cliUserAgents       = []string{"curl", "wget"}
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func CheckAccessKey(r *http.Request, metadata *backends.Metadata) (AccessKeySource, error) {
	key := metadata.AccessKey
	if key == "" {
		return AccessKeySourceNone, nil
	}

	cookieKey, err := r.Cookie(HeaderName)
	if err == nil {
		if cookieKey.Value == key {
			return AccessKeySourceCookie, nil
		}
		return AccessKeySourceCookie, errInvalidAccessKey
	}

	headerKey := r.Header.Get(HeaderName)
	if headerKey == key {
		return AccessKeySourceHeader, nil
	} else if headerKey != "" {
		return AccessKeySourceHeader, errInvalidAccessKey
	}

	formKey := r.PostFormValue(ParamName)
	if formKey == key {
		return AccessKeySourceForm, nil
	} else if formKey != "" {
		return AccessKeySourceForm, errInvalidAccessKey
	}

	queryKey := r.URL.Query().Get(ParamName)
	if queryKey == key {
		return AccessKeySourceQuery, nil
	} else if formKey != "" {
		return AccessKeySourceQuery, errInvalidAccessKey
	}

	return AccessKeySourceNone, errInvalidAccessKey
}

func SetAccessKeyCookies(w http.ResponseWriter, r *http.Request, fileName, value string, expires time.Time) {
	u := headers.GetSiteURL(r)
	cookie := http.Cookie{
		Name:     HeaderName,
		Value:    value,
		HttpOnly: true,
		Domain:   u.Hostname(),
		Expires:  expires,
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

	if _, err := assets.Static().Open(fileName); err == nil {
		AssetHandler(w, r)
		return
	}

	if IsDirectUA(r) {
		FileServeHandler(w, r)
		return
	}

	metadata, err := CheckFile(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, backends.ErrNotFound) {
			ServeAsset(w, r, http.StatusNotFound)
			return
		}
		ErrorMsg(w, r, http.StatusInternalServerError, "Corrupt metadata")
		return
	}

	if src, err := CheckAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == AccessKeySourceCookie {
			SetAccessKeyCookies(w, r, fileName, "", time.Unix(0, 0))
		}

		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: errInvalidAccessKey.Error()})
			return
		}

		ServeAsset(w, r, http.StatusUnauthorized)
		return
	}

	if metadata.AccessKey != "" {
		var expiry time.Time
		if config.Default.Auth.CookieExpiry.Duration != 0 {
			expiry = time.Now().Add(config.Default.Auth.CookieExpiry.Duration)
		}
		SetAccessKeyCookies(w, r, fileName, metadata.AccessKey, expiry)
	}

	FileDisplay(w, r, fileName, metadata)
}
