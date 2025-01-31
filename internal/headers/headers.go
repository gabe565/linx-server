package headers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/andreimarcu/linx-server/internal/config"
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

func GetSiteURL(r *http.Request) string {
	if config.Default.SiteURL != "" {
		return config.Default.SiteURL
	} else {
		u := &url.URL{}
		u.Host = r.Host

		if config.Default.SitePath != "" {
			u.Path = config.Default.SitePath
		}

		if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
			u.Scheme = scheme
		} else if config.Default.CertFile != "" || (r.TLS != nil && r.TLS.HandshakeComplete == true) {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}

		return u.String()
	}
}
