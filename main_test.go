package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"gabe565.com/linx-server/internal/auth/keyhash"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/server"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/upload"
	"gabe565.com/utils/bytefmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RespOkJSON struct {
	Filename  string `json:"filename"`
	URL       string `json:"url"`
	DeleteKey string `json:"delete_key"`
	AccessKey string `json:"access_key"`
	Expiry    string `json:"expiry"`
	Size      string `json:"size"`
}

type RespErrJSON struct {
	Error string `json:"error"`
}

const testURL = "http://linx.example.org/"

func setup(t *testing.T, overrides func()) (*chi.Mux, *httptest.ResponseRecorder) {
	t.Cleanup(func() { config.Default = config.New() })

	u, err := url.Parse(testURL)
	require.NoError(t, err)
	config.Default.SiteURL.URL = *u

	config.Default.FilesPath = t.TempDir()
	config.Default.MetaPath = config.Default.FilesPath + "_meta"
	config.StorageBackend, err = config.Default.NewStorageBackend(t.Context())
	require.NoError(t, err)
	config.Default.MaxSize = bytefmt.GiB
	config.Default.NoLogs = true
	config.Default.SiteName = "linx"
	config.Default.ForceRandomFilename = false

	if overrides != nil {
		overrides()
	}

	r, err := server.Setup()
	require.NoError(t, err)
	return r, httptest.NewRecorder()
}

func assertResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantContentType string) {
	assert.Equal(t, wantStatus, w.Code)
	assert.Equal(t, wantContentType, w.Header().Get("Content-Type"))
}

var indexConfigRe = regexp.MustCompile(`window\.config=(.+?)\s+;`)

func extractConfig(t *testing.T, w *httptest.ResponseRecorder) template.Config {
	require.Regexp(t, indexConfigRe, w.Body.String())
	raw := indexConfigRe.FindSubmatch(w.Body.Bytes())
	require.Len(t, raw, 2)

	var conf template.Config
	require.NoError(t, json.Unmarshal(raw[1], &conf))

	return conf
}

func TestIndex(t *testing.T) {
	r, w := setup(t, nil)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)

	extractConfig(t, w)
}

func TestConfigStandardMaxExpiry(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.MaxExpiry.Duration = 60 * time.Second
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")

	conf := extractConfig(t, w)
	for _, v := range conf.ExpirationTimes {
		assert.NotContains(t, "1 hour", v.Name)
	}
}

func TestConfigWeirdMaxExpiry(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.MaxExpiry.Duration = 25 * time.Minute
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")

	conf := extractConfig(t, w)
	for _, v := range conf.ExpirationTimes {
		assert.NotContains(t, "never", v.Name)
	}
}

func TestAddHeader(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.Header.AddHeaders = map[string]string{"Linx-Test": "It works!"}
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Equal(t, "It works!", w.Header().Get("Linx-Test"))
}

func TestNotFound(t *testing.T) {
	r, w := setup(t, nil)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/url/should/not/exist", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)
}

func TestFileNotFound(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename()
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodGet, path.Join("/", config.Default.SelifPath, filename), nil,
	)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
}

func TestDisplayNotFound(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", filename), nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)
}

func newPostForm(
	t *testing.T,
	filename, content string,
	expiry time.Duration,
	accessKey string,
	randomize bool,
) (*multipart.Writer, bytes.Buffer) {
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	if expiry != 0 {
		fw, err := mw.CreateFormField("expires")
		require.NoError(t, err)

		_, err = io.WriteString(fw, expiry.String())
		require.NoError(t, err)
	}

	if accessKey != "" {
		fw, err := mw.CreateFormField("access_key")
		require.NoError(t, err)

		_, err = io.WriteString(fw, accessKey)
		require.NoError(t, err)
	}

	fw, err := mw.CreateFormField("randomize")
	require.NoError(t, err)

	_, err = io.WriteString(fw, strconv.FormatBool(randomize))
	require.NoError(t, err)

	fw, err = mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = io.WriteString(fw, content)
	require.NoError(t, err)

	require.NoError(t, mw.Close())
	return mw, b
}

const ExtTxt = "txt"

func TestPostUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "File content", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusSeeOther, "")
	assert.Equal(t, testURL+filename, w.Header().Get("Location"))
}

func TestPostUploadWhitelistedHeader(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "File content", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Linx-Expiry", "0")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusSeeOther, "")
}

func TestPostUploadBadOrigin(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "File content", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPostJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".txt"
	mw, b := newPostForm(t, filename, "File content", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	assert.Equal(t, filename, myjson.Filename)
	assert.Equal(t, "0", myjson.Expiry)
	assert.Equal(t, "12", myjson.Size)
}

func TestPostJSONUploadAccessKeyStoredHashed(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".txt"
	mw, b := newPostForm(t, filename, "File content", 0, "supersecret", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.Equal(t, "supersecret", myjson.AccessKey)

	metadata, err := config.StorageBackend.Head(t.Context(), myjson.Filename)
	require.NoError(t, err)
	assert.NotEqual(t, "supersecret", metadata.AccessKey)
	assert.True(t, strings.HasPrefix(metadata.AccessKey, keyhash.KeyPrefix))

	ok, err := keyhash.CheckWithFallback(metadata.AccessKey, "supersecret")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAccessProtectedFileWithAccessKeyHeader(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".txt"
	mw, b := newPostForm(t, filename, "File content", 0, "supersecret", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &myjson))

	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusUnauthorized, "application/json")

	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Access-Key", "supersecret")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")
}

func TestPostJSONUploadMaxExpiry(t *testing.T) {
	r, _ := setup(t, func() {
		config.Default.MaxExpiry.Duration = 5 * time.Minute
	})

	// include 0 to test edge case
	// https://github.com/andreimarcu/linx-server/issues/111
	testExpiries := []string{"86http.StatusBadRequest", "-150", "0"}
	for _, expiry := range testExpiries {
		w := httptest.NewRecorder()

		filename := upload.GenerateBarename() + "." + ExtTxt
		mw, b := newPostForm(t, filename, "File content", 0, "", false)

		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
		require.NoError(t, err)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Linx-Expiry", expiry)
		require.NoError(t, err)

		r.ServeHTTP(w, req)
		assertResponse(t, w, http.StatusOK, "application/json")

		var myjson RespOkJSON
		err = json.Unmarshal(w.Body.Bytes(), &myjson)
		require.NoError(t, err)

		myExp, err := strconv.ParseInt(myjson.Expiry, 10, 64)
		require.NoError(t, err)

		expected := time.Now().Add(config.Default.MaxExpiry.Duration).Unix()
		assert.Equal(t, expected, myExp)
	}
}

func TestPostExpiresJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".txt"
	mw, b := newPostForm(t, filename, "File content", time.Minute, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	assert.Equal(t, filename, myjson.Filename)

	myExp, err := strconv.ParseInt(myjson.Expiry, 10, 64)
	require.NoError(t, err)
	curTime := time.Now().Unix()
	assert.Less(t, curTime, myExp, "file expiry smaller than current time")
	assert.Equal(t, "12", myjson.Size)
}

func TestPostRandomizeJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "File content", 0, "", true)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.NotEqual(t, filename, myjson.Filename, "filename is not random")
	assert.Equal(t, "12", myjson.Size)
}

func TestPostEmptyUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPostTooLargeUpload(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.MaxSize = 2
	})

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "File content", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusRequestEntityTooLarge, "text/html; charset=utf-8")
}

func TestPostEmptyJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + "." + ExtTxt
	mw, b := newPostForm(t, filename, "", 0, "", false)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "application/json")

	var myjson RespErrJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	assert.Equal(t, "Empty file", myjson.Error)
}

func TestPutUpload(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"with filename", "/upload/" + upload.GenerateBarename() + ".txt"},
		{"bare", "/upload"},
		{"trailing slash", "/upload/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w := setup(t, nil)

			req, err := http.NewRequestWithContext(t.Context(),
				http.MethodPut, tt.path, strings.NewReader("File content"),
			)
			require.NoError(t, err)

			r.ServeHTTP(w, req)
			assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

			filename := strings.TrimPrefix(strings.TrimPrefix(tt.path, "/upload"), "/")
			if filename == "" {
				assert.NotEmpty(t, w.Body.String())
			} else {
				expect, err := config.Default.SiteURL.Parse(filename)
				require.NoError(t, err)
				assert.Equal(t, expect.String()+"\n", w.Body.String())
			}
		})
	}
}

func TestPutRandomizedUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	req.Header.Set("Linx-Randomize", "true")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutForceRandomUpload(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.ForceRandomFilename = true
	})

	filename := "randomizeme.file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	// while this should also work without this header, let's try to force
	// the randomized filename off to be sure
	req.Header.Set("Linx-Randomize", "false")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutNoExtensionUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename()
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	req.Header.Set("Linx-Randomize", "true")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutEmptyUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader(""),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Randomize", "true")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPutTooLargeUpload(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.MaxSize = 2
	})

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File too big"),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Randomize", "true")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusRequestEntityTooLarge, "text/html; charset=utf-8")
	assert.NotContains(t, "request body too large", w.Body.String())
}

func TestPutJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.Equal(t, filename, myjson.Filename, "filename is not random")
}

func TestPutRandomizedJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Randomize", "true")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.NotEqual(t, filename, myjson.Filename, "filename was not random")
}

func TestPutExpireJSONUpload(t *testing.T) {
	r, w := setup(t, nil)

	filename := upload.GenerateBarename() + ".file"
	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload/", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Expiry", "600")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	expiry, err := strconv.Atoi(myjson.Expiry)
	require.NoError(t, err)
	assert.NotZero(t, expiry)
}

func TestPutAndDelete(t *testing.T) {
	r, w := setup(t, nil)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodDelete, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)
	req.Header.Set("Linx-Delete-Key", myjson.DeleteKey)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")

	// Make sure torrent is also gone
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/torrent", myjson.Filename), nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
}

func TestPutAndOverwrite(t *testing.T) {
	r, w := setup(t, nil)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	// Overwrite it
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", myjson.Filename), strings.NewReader("New file content"),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Delete-Key", myjson.DeleteKey)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")
	assert.Equal(t, http.StatusOK, w.Code)

	// Make sure it's the new file
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(),
		http.MethodGet, path.Join("/", config.Default.SelifPath, myjson.Filename), nil,
	)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")
	assert.Equal(t, "New file content", w.Body.String())
}

func TestPutAndOverwriteForceRandom(t *testing.T) {
	r, w := setup(t, func() {
		config.Default.ForceRandomFilename = true
	})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	// Overwrite it
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", myjson.Filename), strings.NewReader("New file content"),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Delete-Key", myjson.DeleteKey)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	// Make sure it's the new file
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(),
		http.MethodGet, path.Join("/", config.Default.SelifPath, myjson.Filename), nil,
	)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	assert.Equal(t, "New file content", w.Body.String())
}

func TestPutAndSpecificDelete(t *testing.T) {
	r, w := setup(t, nil)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Delete-Key", "supersecret")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.Equal(t, "supersecret", myjson.DeleteKey)

	metadata, err := config.StorageBackend.Head(t.Context(), myjson.Filename)
	require.NoError(t, err)
	assert.NotEqual(t, "supersecret", metadata.DeleteKey)
	assert.NotEmpty(t, metadata.DeleteKey)

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodDelete, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)
	req.Header.Set("Linx-Delete-Key", "supersecret")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", myjson.Filename), nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")

	// Make sure torrent is gone too
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/torrent", myjson.Filename), nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
}

func TestPutAndGetCLI(t *testing.T) {
	r, _ := setup(t, nil)

	// upload file
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	// request file without wget user agent
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, myjson.URL, nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")

	contentType := w.Header().Get("Content-Type")
	assert.NotRegexp(t, "^text/plain", contentType, "didn't receive file display page")

	// request file with wget user agent
	w = httptest.NewRecorder()
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, myjson.URL, nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "wget")
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")
}
