package server

import (
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/template"
)

const (
	DefaultCSP    = "default-src 'self' " + defaultSrcKey + "; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none';"
	defaultSrcKey = "$DEFAULT_SRC"

	cspHeader          = "Content-Security-Policy"
	rpHeader           = "Referrer-Policy"
	frameOptionsHeader = "X-Frame-Options"
)

type CSPMiddleware struct {
	h    http.Handler
	opts Options
}

type Options struct {
	Policy         string
	ReferrerPolicy string
	Frame          string
}

func (c CSPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only add a CSP if one is not already set
	if existing := w.Header().Get(cspHeader); existing == "" {
		w.Header().Add(cspHeader, c.opts.Policy)
	}

	// only add a Referrer Policy if one is not already set
	if existing := w.Header().Get(rpHeader); existing == "" {
		w.Header().Add(rpHeader, c.opts.ReferrerPolicy)
	}

	w.Header().Set(frameOptionsHeader, c.opts.Frame)

	c.h.ServeHTTP(w, r)
}

func NewCSPMiddleware(o Options) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return CSPMiddleware{h, o}
	}
	return fn
}

func GenerateCSP() string {
	defaultSrc := template.ConfigHash()
	if u := config.Default.ViteURL; u != "" {
		defaultSrc += " " + u + " ws:"
	}
	return strings.Replace(DefaultCSP, defaultSrcKey, defaultSrc, 1)
}
