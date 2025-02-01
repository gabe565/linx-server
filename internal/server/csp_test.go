package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gabe565.com/linx-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCSPHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'none'; style-src 'self';",
	"Referrer-Policy":         "strict-origin-when-cross-origin",
	"X-Frame-Options":         "SAMEORIGIN",
}

func TestContentSecurityPolicy(t *testing.T) {
	config.Default.SiteURL = "http://linx.example.org/"
	config.Default.FilesDir = t.TempDir()
	config.Default.MetaDir = config.Default.FilesDir + "_meta"
	config.Default.MaxSize = 1024 * 1024 * 1024
	config.Default.NoLogs = true
	config.Default.SiteName = "linx"
	config.Default.SelifPath = "/selif"
	config.Default.ContentSecurityPolicy = testCSPHeaders["Content-Security-Policy"]
	config.Default.ReferrerPolicy = testCSPHeaders["Referrer-Policy"]
	config.Default.XFrameOptions = testCSPHeaders["X-Frame-Options"]
	r, err := Setup()
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)

	for k, v := range testCSPHeaders {
		assert.Equal(t, v, w.Header().Get(k))
	}
}
