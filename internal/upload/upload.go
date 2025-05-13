package upload

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/auth/apikeys"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/csrf"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/helpers"
	"gabe565.com/linx-server/internal/util"
	"github.com/dchest/uniuri"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
)

//nolint:gochecknoglobals
var fileDenylist = []string{
	"favicon.ico",
	"crossdomain.xml",
}

// Describes metadata directly from the user request.
type Request struct {
	src            io.Reader
	size           int64
	filename       string
	expiry         time.Duration // Seconds until expiry, 0 = never
	deleteKey      string        // Empty string if not defined
	randomBarename bool
	accessKey      string // Empty string if not defined
}

// Metadata associated with a file as it would actually be stored.
type Upload struct {
	OriginalName string
	Filename     string // Final filename on disk
	Metadata     backends.Metadata
}

type JSONResponse struct {
	URL          string `json:"url"`
	OriginalName string `json:"original_name,omitzero"`
	DirectURL    string `json:"direct_url"`
	Filename     string `json:"filename"`
	DeleteKey    string `json:"delete_key"`
	AccessKey    string `json:"access_key"`
	Expiry       string `json:"expiry"`
	Size         string `json:"size"`
	Mimetype     string `json:"mimetype"`
}

func (u Upload) JSONResponse(r *http.Request) JSONResponse {
	var expiry int64
	if v := u.Metadata.Expiry.Unix(); v > 0 {
		expiry = v
	}
	return JSONResponse{
		URL:          headers.GetFileURL(r, u.Filename).String(),
		OriginalName: u.OriginalName,
		DirectURL:    headers.GetSelifURL(r, u.Filename).String(),
		Filename:     u.Filename,
		DeleteKey:    u.Metadata.DeleteKey,
		AccessKey:    u.Metadata.AccessKey,
		Expiry:       strconv.FormatInt(expiry, 10),
		Size:         strconv.FormatInt(u.Metadata.Size, 10),
		Mimetype:     u.Metadata.Mimetype,
	}
}

func POSTHandler(w http.ResponseWriter, r *http.Request) {
	if config.Default.FrontendURL == "" {
		siteURL := headers.GetSiteURL(r).String()
		if !csrf.StrictReferrerCheck(r, siteURL,
			[]string{"Linx-Delete-Key", "Linx-Expiry", "Linx-Randomize", "X-Requested-With"},
		) {
			handlers.Error(w, r, http.StatusBadRequest)
			return
		}
	}

	upReq := Request{}
	HeaderProcess(r, &upReq)

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(int64(config.Default.UploadMaxMemory)); err != nil {
			HandleProcessError(w, r, err)
			return
		}

		file, headers, err := r.FormFile("file")
		if err != nil {
			HandleProcessError(w, r, err)
			return
		}
		defer func() {
			_ = file.Close()
		}()

		upReq.src = file
		upReq.size = headers.Size
		upReq.filename = headers.Filename
	} else {
		if r.PostFormValue("content") == "" {
			HandleProcessError(w, r, backends.ErrFileEmpty)
			return
		}
		extension := r.PostFormValue("extension")
		if extension == "" {
			extension = "txt"
		}

		content := r.PostFormValue("content")

		upReq.src = strings.NewReader(content)
		upReq.size = int64(len(content))
		upReq.filename = r.PostFormValue("filename") + "." + extension
	}

	upReq.expiry = ParseExpiry(r.PostFormValue("expires"))
	upReq.accessKey = r.PostFormValue(handlers.AccessKeyParam)
	upReq.randomBarename = util.ParseBool(r.PostFormValue("randomize"), false)

	upload, err := Process(r.Context(), upReq)
	if err != nil {
		HandleProcessError(w, r, err)
		return
	}

	w.Header().Set("Vary", "Accept")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(upload.JSONResponse(r))
	} else {
		http.Redirect(w, r, headers.GetFileURL(r, upload.Filename).String(), http.StatusSeeOther)
	}
}

func PUTHandler(w http.ResponseWriter, r *http.Request) {
	upReq := Request{}
	HeaderProcess(r, &upReq)

	upReq.filename = chi.URLParam(r, "name")
	upReq.src = r.Body

	upload, err := Process(r.Context(), upReq)
	if err != nil {
		HandleProcessError(w, r, err)
		return
	}

	w.Header().Set("Vary", "Accept")
	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(upload.JSONResponse(r))
	} else {
		_, _ = io.WriteString(w, headers.GetFileURL(r, upload.Filename).String()+"\n")
	}
}

func Remote(w http.ResponseWriter, r *http.Request) {
	if config.Default.Auth.RemoteFile != "" {
		key := r.FormValue("key")
		if key == "" && config.Default.Auth.Basic {
			_, password, ok := r.BasicAuth()
			if ok {
				key = password
			}
		}
		result, err := apikeys.CheckAuth(config.RemoteAuthKeys, key)
		if err != nil || !result {
			if config.Default.Auth.Basic {
				rs := ""
				if config.Default.SiteName != "" {
					rs = " realm=" + strconv.Quote(config.Default.SiteName)
				}
				w.Header().Set("WWW-Authenticate", `Basic`+rs)
			}
			handlers.Error(w, r, http.StatusUnauthorized)
			return
		}
	}

	if r.FormValue("url") == "" {
		http.Redirect(w, r, config.Default.SiteURL.String(), http.StatusSeeOther)
		return
	}

	grabURL, err := url.Parse(r.FormValue("url"))
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Invalid URL")
		return
	}
	directURL := util.ParseBool(r.FormValue("direct_url"), false)

	upReq := Request{
		filename: filepath.Base(grabURL.Path),
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, grabURL.String(), nil)
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusInternalServerError, "Failed to create request")
		return
	}
	req.Header.Set("User-Agent", r.UserAgent())

	client := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			upReq.filename = filepath.Base(req.URL.Path)
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusServiceUnavailable, "Could not retrieve URL")
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		handlers.ErrorMsg(w, r, resp.StatusCode, "Remote host returned error "+resp.Status)
		return
	}

	if v := chi.URLParam(r, "name"); v != "" {
		upReq.filename = v
	} else if upReq.filename == "" {
		if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
			_, params, err := mime.ParseMediaType(disposition)
			if err == nil && params["filename"] != "" {
				upReq.filename = params["filename"]
			}
		}
	}

	upReq.src = http.MaxBytesReader(w, resp.Body, int64(config.Default.MaxSize))
	upReq.deleteKey = r.FormValue("deletekey")
	upReq.accessKey = r.FormValue(handlers.AccessKeyParam)
	upReq.randomBarename = util.ParseBool(r.FormValue("randomize"), false)
	upReq.expiry = ParseExpiry(r.FormValue("expiry"))

	upload, err := Process(r.Context(), upReq)
	if err != nil {
		HandleProcessError(w, r, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(upload.JSONResponse(r))
	} else {
		var u *url.URL
		if directURL {
			u = headers.GetSelifURL(r, upload.Filename)
		} else {
			u = headers.GetFileURL(r, upload.Filename)
		}
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
	}
}

func HeaderProcess(r *http.Request, upReq *Request) {
	upReq.randomBarename = util.ParseBool(r.Header.Get("Linx-Randomize"), false)

	upReq.deleteKey = r.Header.Get("Linx-Delete-Key")
	upReq.accessKey = r.Header.Get(handlers.AccessKeyHeader)

	// Get seconds until expiry. Non-integer responses never expire.
	expStr := r.Header.Get("Linx-Expiry")
	upReq.expiry = ParseExpiry(expStr)
}

var ErrProhibitedFilename = errors.New("prohibited filename")

func Process(ctx context.Context, upReq Request) (Upload, error) {
	var upload Upload
	if upReq.size > int64(config.Default.MaxSize) {
		return upload, &http.MaxBytesError{Limit: int64(config.Default.MaxSize)}
	}

	// Determine the appropriate filename
	barename, extension := BarePlusExt(upReq.filename)
	randomize := false

	if config.Default.KeepOriginalFilename && !strings.HasPrefix(upReq.filename, ".") {
		upload.OriginalName = upReq.filename
	}

	// Randomize the "barename" (filename without extension) if needed
	if upReq.randomBarename || len(barename) == 0 {
		barename = GenerateBarename()
		randomize = true
	}

	if len(extension) == 0 {
		// Determine the type of file from the file header
		var kind *mimetype.MIME
		var err error
		kind, upReq.src, err = helpers.DetectMimetype(upReq.src)
		if err != nil {
			return upload, err
		}

		if len(kind.Extension()) < 2 {
			extension = "file"
		} else {
			extension = kind.Extension()[1:] // remove leading "."
		}
	}

	upload.Filename = barename + "." + extension

	fileexists, err := config.StorageBackend.Exists(ctx, upload.Filename)
	if err != nil {
		return upload, err
	}

	// Check if the delete key matches, in which case overwrite
	if fileexists {
		metad, merr := config.StorageBackend.Head(ctx, upload.Filename)
		if merr == nil {
			if upReq.deleteKey == metad.DeleteKey {
				fileexists = false
			} else if config.Default.ForceRandomFilename {
				// the file exists
				// the delete key doesn't match
				// force random filenames is enabled
				randomize = true
			}
		}
	} else if config.Default.ForceRandomFilename {
		// the file doesn't exist
		// force random filenames is enabled
		randomize = true

		// set fileexists to true to generate a new barename
		fileexists = true
	}

	origBarename := barename
	var counter int
	for fileexists {
		if randomize {
			barename = GenerateBarename()
		} else {
			counter++
			barename = origBarename + strconv.Itoa(counter)
		}
		upload.Filename = barename + "." + extension

		var err error
		fileexists, err = config.StorageBackend.Exists(ctx, upload.Filename)
		if err != nil {
			return upload, err
		}
	}

	if strings.HasPrefix(upload.Filename, "index.") {
		return upload, ErrProhibitedFilename
	}
	if slices.Contains(fileDenylist, upload.Filename) {
		return upload, ErrProhibitedFilename
	}
	if _, err := assets.Static().Open(strings.TrimPrefix(upload.Filename, "/")); err == nil || !os.IsNotExist(err) {
		return upload, ErrProhibitedFilename
	}

	// Get the rest of the metadata needed for storage
	var fileExpiry time.Time
	if upReq.expiry == 0 {
		fileExpiry = time.Time{}
	} else {
		fileExpiry = time.Now().Add(upReq.expiry)
	}

	if upReq.deleteKey == "" {
		upReq.deleteKey = uniuri.NewLen(30)
	}

	upload.Metadata, err = config.StorageBackend.Put(ctx,
		upload.OriginalName,
		upload.Filename,
		upReq.src,
		fileExpiry,
		upReq.deleteKey,
		upReq.accessKey,
	)
	if err != nil {
		return upload, err
	}

	return upload, err
}

func HandleProcessError(w http.ResponseWriter, r *http.Request, err error) {
	var maxBytes *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytes):
		handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
	case errors.Is(err, backends.ErrFileEmpty):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Empty file")
	case errors.Is(err, ErrProhibitedFilename):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Prohibited filename")
	case errors.Is(err, io.ErrUnexpectedEOF):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Upload canceled")
	default:
		slog.Error("Upload failed", "error", err)
		handlers.Error(w, r, http.StatusInternalServerError)
	}
}

func GenerateBarename() string {
	return uniuri.NewLenChars(8, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

//nolint:gochecknoglobals
var compressedExts = []string{
	".gz",
	".xz",
	".bz2",
	".zst",
	".lzma",
	".lzo",
	".z",
}

func BarePlusExt(filename string) (string, string) {
	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)

	extension := path.Ext(filename)
	barename := strings.TrimSuffix(filename, extension)
	if slices.Contains(compressedExts, extension) {
		if ext2 := path.Ext(barename); ext2 != "" {
			extension = ext2 + extension
			barename = strings.TrimSuffix(barename, ext2)
		}
	}

	extension = strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z', '0' <= r && r <= '9', r == '-', r == '.':
			return r
		default:
			return -1
		}
	}, extension)
	barename = slug.Make(barename)

	extension = strings.Trim(extension, ".")

	return barename, extension
}

func ParseExpiry(expStr string) time.Duration {
	if expStr == "" {
		return config.Default.MaxExpiry.Duration
	}

	var fileExpiry time.Duration
	if t, err := time.ParseDuration(expStr); err == nil {
		fileExpiry = t
	} else {
		seconds, err := strconv.ParseUint(expStr, 10, 64)
		if err != nil {
			return config.Default.MaxExpiry.Duration
		}

		fileExpiry = time.Duration(seconds) * time.Second //nolint:gosec
	}

	if config.Default.MaxExpiry.Duration > 0 && (fileExpiry > config.Default.MaxExpiry.Duration || fileExpiry == 0) {
		fileExpiry = config.Default.MaxExpiry.Duration
	}
	return fileExpiry
}
