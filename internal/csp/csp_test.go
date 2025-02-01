package csp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/server"
	"gabe565.com/linx-server/internal/upload"
)

var testCSPHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'none'; style-src 'self';",
	"Referrer-Policy":         "strict-origin-when-cross-origin",
	"X-Frame-Options":         "SAMEORIGIN",
}

func TestContentSecurityPolicy(t *testing.T) {
	config.Default.SiteURL = "http://linx.example.org/"
	config.Default.FilesDir = path.Join(os.TempDir(), upload.GenerateBarename())
	config.Default.MetaDir = config.Default.FilesDir + "_meta"
	config.Default.MaxSize = 1024 * 1024 * 1024
	config.Default.NoLogs = true
	config.Default.SiteName = "linx"
	config.Default.SelifPath = "/selif"
	config.Default.ContentSecurityPolicy = testCSPHeaders["Content-Security-Policy"]
	config.Default.ReferrerPolicy = testCSPHeaders["Referrer-Policy"]
	config.Default.XFrameOptions = testCSPHeaders["X-Frame-Options"]
	mux, err := server.Setup()
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.Use(ContentSecurityPolicy(CSPOptions{
		Policy:         testCSPHeaders["Content-Security-Policy"],
		ReferrerPolicy: testCSPHeaders["Referrer-Policy"],
		Frame:          testCSPHeaders["X-Frame-Options"],
	}))

	mux.ServeHTTP(w, req)

	for k, v := range testCSPHeaders {
		if w.HeaderMap[k][0] != v {
			t.Fatalf("%s header did not match expected value set by middleware", k)
		}
	}
}
