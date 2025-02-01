package headers

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"gabe565.com/linx-server/internal/config"
)

type addheaders struct {
	h       http.Handler
	headers []string
}

func (a addheaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, header := range a.headers {
		headerSplit := strings.SplitN(header, ": ", 2)
		w.Header().Add(headerSplit[0], headerSplit[1])
	}

	a.h.ServeHTTP(w, r)
}

func AddHeaders(headers []string) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return addheaders{h, headers}
	}
	return fn
}

func GetSiteURL(r *http.Request) (*url.URL, error) {
	if config.Default.SiteURL != "" || r == nil {
		return url.Parse(config.Default.SiteURL)
	}

	u := &url.URL{Host: r.Host}

	if config.Default.SitePath != "" {
		u.Path = config.Default.SitePath
	}

	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		u.Scheme = scheme
	} else if config.Default.TLSCert != "" || (r.TLS != nil && r.TLS.HandshakeComplete) {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	return u, nil
}

func GetFileURL(r *http.Request, filename string) (*url.URL, error) {
	u, err := GetSiteURL(r)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, filename)
	return u, nil
}

func GetSelifURL(r *http.Request, filename string) (*url.URL, error) {
	u, err := GetSiteURL(r)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, config.Default.SelifPath)
	if filename != "" {
		u.Path = path.Join(u.Path, filename)
	}
	return u, nil
}
