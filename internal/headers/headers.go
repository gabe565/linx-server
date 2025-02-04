package headers

import (
	"net/http"
	"net/url"
	"path"

	"gabe565.com/linx-server/internal/config"
)

type addheaders struct {
	h       http.Handler
	headers map[string]string
}

func (a addheaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range a.headers {
		w.Header().Set(k, v)
	}

	a.h.ServeHTTP(w, r)
}

func AddHeaders(headers map[string]string) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return addheaders{h, headers}
	}
	return fn
}

func GetSiteURL(r *http.Request) *url.URL {
	switch {
	case config.Default.SiteURL.Host != "", r == nil:
		u := config.Default.SiteURL.URL
		return &u
	default:
		u := config.Default.SiteURL.URL
		u.Host = r.Host

		if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
			u.Scheme = scheme
		} else if config.Default.TLS.Cert != "" || (r.TLS != nil && r.TLS.HandshakeComplete) {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}

		return &u
	}
}

func GetFileURL(r *http.Request, filename string) *url.URL {
	u := GetSiteURL(r)
	u.Path = path.Join(u.Path, filename)
	return u
}

func GetSelifURL(r *http.Request, filename string) *url.URL {
	u := GetSiteURL(r)
	u.Path = path.Join(u.Path, config.Default.SelifPath)
	if filename != "" {
		u.Path = path.Join(u.Path, filename)
	}
	return u
}
