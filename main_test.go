package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/server"
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
	Expiry    string `json:"expiry"`
	Size      string `json:"size"`
}

type RespErrJSON struct {
	Error string `json:"error"`
}

const testURL = "http://linx.example.org/"

func setup(t *testing.T) (*chi.Mux, *httptest.ResponseRecorder) {
	u, err := url.Parse(testURL)
	require.NoError(t, err)
	config.Default.SiteURL.URL = *u

	config.Default.FilesPath = t.TempDir()
	config.Default.MetaPath = config.Default.FilesPath + "_meta"
	config.Default.MaxSize = bytefmt.GiB
	config.Default.NoLogs = true
	config.Default.SiteName = "linx"
	t.Cleanup(func() { config.Default = config.New() })

	r, err := server.Setup(t.Context())
	require.NoError(t, err)
	return r, httptest.NewRecorder()
}

func assertResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantContentType string) {
	assert.Equal(t, wantStatus, w.Code)
	assert.Equal(t, wantContentType, w.Header().Get("Content-Type"))
}

func TestIndex(t *testing.T) {
	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)
}

func TestConfigStandardMaxExpiry(t *testing.T) {
	r, w := setup(t)

	config.Default.MaxExpiry.Duration = 60 * time.Second

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/api/config", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")
	assert.NotContains(t, w.Body.String(), "1 hour")
}

func TestConfigWeirdMaxExpiry(t *testing.T) {
	r, w := setup(t)

	config.Default.MaxExpiry.Duration = 25 * time.Minute

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/api/config", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")
	assert.NotContains(t, w.Body.String(), "never")
}

func TestAddHeader(t *testing.T) {
	config.Default.Header.AddHeaders = map[string]string{"Linx-Test": "It works!"}
	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Equal(t, "It works!", w.Header().Get("Linx-Test"))
}

func TestNotFound(t *testing.T) {
	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/url/should/not/exist", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)
}

func TestFileNotFound(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodGet, path.Join("/", config.Default.SelifPath, filename), nil,
	)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
}

func TestDisplayNotFound(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path.Join("/", filename), nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusNotFound, "text/html; charset=utf-8")
	assert.Contains(t, w.Body.String(), `<div id="app">`)
}

const ExtTxt = "txt"

func TestPostCodeUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()
	extension := ExtTxt

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", nil)
	require.NoError(t, err)
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", config.Default.SiteURL.String())

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusSeeOther, "")
	assert.Equal(t, testURL+filename+"."+extension, w.Header().Get("Location"))
}

func TestPostCodeUploadWhitelistedHeader(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()
	extension := ExtTxt

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", nil)
	require.NoError(t, err)
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Linx-Expiry", "0")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusSeeOther, "")
}

func TestPostCodeUploadNoReferrer(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()
	extension := ExtTxt

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", nil)
	require.NoError(t, err)
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPostCodeUploadBadOrigin(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()
	extension := ExtTxt

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", nil)
	require.NoError(t, err)
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPostCodeExpiryJSONUpload(t *testing.T) {
	r, w := setup(t)

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", "")
	form.Add("expires", "60")

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", nil)
	require.NoError(t, err)
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", config.Default.SiteURL.String())
	req.Header.Set("Origin", strings.TrimSuffix(config.Default.SiteURL.String(), "/"))

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assertResponse(t, w, http.StatusOK, "application/json")

	var myjson RespOkJSON
	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	myExp, err := strconv.ParseInt(myjson.Expiry, 10, 64)
	require.NoError(t, err)
	curTime := time.Now().Unix()
	assert.Less(t, curTime, myExp, "file expiry smaller than current time")
	assert.Equal(t, "12", myjson.Size)
}

func TestPostUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte("File content"))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusSeeOther, "")
	assert.Equal(t, testURL+filename, w.Header().Get("Location"))
}

func TestPostJSONUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte("File content"))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

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

func TestPostJSONUploadMaxExpiry(t *testing.T) {
	r, _ := setup(t)
	config.Default.MaxExpiry.Duration = 5 * time.Minute

	// include 0 to test edge case
	// https://github.com/andreimarcu/linx-server/issues/111
	testExpiries := []string{"86http.StatusBadRequest", "-150", "0"}
	for _, expiry := range testExpiries {
		w := httptest.NewRecorder()

		filename := upload.GenerateBarename() + ".txt"

		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		fw, err := mw.CreateFormFile("file", filename)
		require.NoError(t, err)

		_, err = fw.Write([]byte("File content"))
		require.NoError(t, err)
		require.NoError(t, mw.Close())

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

	config.Default.MaxExpiry.Duration = 0
}

func TestPostExpiresJSONUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write([]byte("File content"))
	require.NoError(t, err)

	exp, err := mw.CreateFormField("expires")
	require.NoError(t, err)
	_, err = exp.Write([]byte("60"))
	require.NoError(t, err)

	require.NoError(t, mw.Close())

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
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write([]byte("File content"))
	require.NoError(t, err)

	rnd, err := mw.CreateFormField("randomize")
	require.NoError(t, err)
	_, err = rnd.Write([]byte("true"))
	require.NoError(t, err)

	require.NoError(t, mw.Close())

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
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte(""))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPostTooLargeUpload(t *testing.T) {
	r, w := setup(t)
	config.Default.MaxSize = 2

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte("test content"))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/upload", &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", config.Default.SiteURL.String())
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusRequestEntityTooLarge, "text/html; charset=utf-8")
}

func TestPostEmptyJSONUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte(""))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

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
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.Equal(t, expect.String()+"\n", w.Body.String())
}

func TestPutRandomizedUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	req.Header.Set("Linx-Randomize", "yes")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutForceRandomUpload(t *testing.T) {
	r, w := setup(t)

	config.Default.ForceRandomFilename = true
	filename := "randomizeme.file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	// while this should also work without this header, let's try to force
	// the randomized filename off to be sure
	req.Header.Set("Linx-Randomize", "no")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutNoExtensionUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename()

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)

	req.Header.Set("Linx-Randomize", "yes")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "text/plain; charset=utf-8")

	expect, err := config.Default.SiteURL.Parse(filename)
	require.NoError(t, err)
	assert.NotEqual(t, expect.String(), w.Body.String())
}

func TestPutEmptyUpload(t *testing.T) {
	r, w := setup(t)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader(""),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Randomize", "yes")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusBadRequest, "text/html; charset=utf-8")
}

func TestPutTooLargeUpload(t *testing.T) {
	_, w := setup(t)
	config.Default.MaxSize = 2
	r, err := server.Setup(t.Context())
	require.NoError(t, err)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File too big"),
	)
	require.NoError(t, err)
	req.Header.Set("Linx-Randomize", "yes")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusRequestEntityTooLarge, "text/html; charset=utf-8")
	assert.NotContains(t, "request body too large", w.Body.String())
}

func TestPutJSONUpload(t *testing.T) {
	r, w := setup(t)

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
	var myjson RespOkJSON

	r, w := setup(t)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Randomize", "yes")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)
	assert.NotEqual(t, filename, myjson.Filename, "filename was not random")
}

func TestPutExpireJSONUpload(t *testing.T) {
	var myjson RespOkJSON

	r, w := setup(t)

	filename := upload.GenerateBarename() + ".file"

	req, err := http.NewRequestWithContext(t.Context(),
		http.MethodPut, path.Join("/upload/", filename), strings.NewReader("File content"),
	)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Expiry", "600")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

	expiry, err := strconv.Atoi(myjson.Expiry)
	require.NoError(t, err)
	assert.NotZero(t, expiry)
}

func TestPutAndDelete(t *testing.T) {
	var myjson RespOkJSON

	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

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
	var myjson RespOkJSON

	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

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
	var myjson RespOkJSON

	r, w := setup(t)

	config.Default.ForceRandomFilename = true

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

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
	var myjson RespOkJSON

	r, w := setup(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Delete-Key", "supersecret")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

	err = json.Unmarshal(w.Body.Bytes(), &myjson)
	require.NoError(t, err)

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
	var myjson RespOkJSON
	r, _ := setup(t)

	// upload file
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, "/upload", strings.NewReader("File content"))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	r.ServeHTTP(w, req)
	assertResponse(t, w, http.StatusOK, "application/json")

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
