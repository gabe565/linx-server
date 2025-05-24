package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/utils/bytefmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentSecurityPolicy(t *testing.T) {
	testCSPHeaders := map[string]string{
		"Content-Security-Policy": DefaultCSP,
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"X-Frame-Options":         "SAMEORIGIN",
	}

	// config.Default.SiteURL = "http://linx.example.org/"
	config.Default.SiteURL.URL = url.URL{Scheme: "http", Host: "linx.example.org"}
	config.Default.FilesPath = t.TempDir()
	config.Default.MetaPath = config.Default.FilesPath + "_meta"
	config.Default.MaxSize = bytefmt.GiB
	config.Default.NoLogs = true
	config.Default.SiteName = "linx"
	config.Default.SelifPath = "/selif"
	config.Default.Header.ReferrerPolicy = testCSPHeaders["Referrer-Policy"]
	config.Default.Header.XFrameOptions = testCSPHeaders["X-Frame-Options"]
	r, err := Setup()
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)

	for k, v := range testCSPHeaders {
		assert.Equal(t, v, w.Header().Get(k))
	}
}
