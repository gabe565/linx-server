package server

import (
	"context"
	"log/slog"
	"net/http"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/csp"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/util"
)

const (
	cspHeader = "Content-Security-Policy"
	rpHeader  = "Referrer-Policy"
)

type CSPMiddleware struct {
	h    http.Handler
	opts Options
}

type Options struct {
	Policy         string
	ReferrerPolicy string
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

	c.h.ServeHTTP(w, r)
}

func NewCSPMiddleware(o Options) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return CSPMiddleware{h, o}
	}
	return fn
}

func GenerateCSP() csp.CSP {
	conf, err := template.ConfigBytes()
	if err != nil {
		panic(err)
	}

	defaultSrc := []string{csp.Self, util.SubresourceIntegrity(conf)}
	if u := config.Default.ViteURL; u != "" {
		defaultSrc = append(defaultSrc, u, "ws:")
	}

	policy := csp.CSP{
		"default-src":     defaultSrc,
		"img-src":         {csp.Self, csp.Data},
		"style-src":       {csp.Self, csp.UnsafeInline},
		"frame-ancestors": {csp.None},
	}

	if origin := s3PresignedOrigin(); origin != "" {
		policy["img-src"] = append(policy["img-src"], origin)
		policy["media-src"] = []string{csp.Self, origin}
		policy["object-src"] = []string{csp.Self, origin}
		policy["connect-src"] = []string{csp.Self, origin}
	}

	return policy
}

func s3PresignedOrigin() string {
	if !config.Default.S3.PresignedURLs {
		return ""
	}
	ctx := context.Background()
	backend, err := config.Default.NewS3Backend(ctx)
	if err != nil {
		slog.Warn("Could not initialize S3 backend to probe presigned origin", "error", err)
		return ""
	}
	origin, err := backend.PresignedOrigin(ctx)
	if err != nil {
		slog.Warn("Could not probe presigned URL origin for CSP", "error", err)
		return ""
	}
	return origin
}
