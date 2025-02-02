package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/templates"
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

var (
	errInvalidAccessKey = errors.New("invalid access key")

	cliUserAgentRe = regexp.MustCompile("(?i)(lib)?curl|wget")
)

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

func FileAccessHeader(w http.ResponseWriter, r *http.Request) {
	if !config.Default.NoDirectAgents && cliUserAgentRe.MatchString(r.Header.Get("User-Agent")) && !strings.EqualFold("application/json", r.Header.Get("Accept")) {
		FileServeHandler(w, r)
		return
	}

	fileName := chi.URLParam(r, "name")

	metadata, err := CheckFile(r.Context(), fileName)
	if errors.Is(err, backends.ErrNotFound) {
		NotFound(w, r)
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

		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			dec := json.NewEncoder(w)
			_ = dec.Encode(map[string]string{
				"error": errInvalidAccessKey.Error(),
			})

			return
		}

		_ = templates.Render("access.html", map[string]any{
			"FileName":   fileName,
			"AccessPath": fileName,
		}, r, w)

		return
	}

	if metadata.AccessKey != "" {
		var expiry time.Time
		if config.Default.AccessKeyCookieExpiry != 0 {
			expiry = time.Now().Add(time.Duration(config.Default.AccessKeyCookieExpiry) * time.Second) //nolint:gosec
		}
		SetAccessKeyCookies(w, r, fileName, metadata.AccessKey, expiry)
	}

	FileDisplay(w, r, fileName, metadata)
}
