package server

import (
	"net/http"
)

const (
	cspHeader          = "Content-Security-Policy"
	rpHeader           = "Referrer-Policy"
	frameOptionsHeader = "X-Frame-Options"
)

type CSP struct {
	h    http.Handler
	opts CSPOptions
}

type CSPOptions struct {
	Policy         string
	ReferrerPolicy string
	Frame          string
}

func (c CSP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func ContentSecurityPolicy(o CSPOptions) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return CSP{h, o}
	}
	return fn
}
