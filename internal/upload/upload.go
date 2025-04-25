package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
	"gabe565.com/utils/bytefmt"
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
	Filename string // Final filename on disk
	Metadata backends.Metadata
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
		file, headers, err := r.FormFile("file")
		if err != nil {
			var maxBytes *http.MaxBytesError
			if errors.As(err, &maxBytes) {
				handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
			} else {
				slog.Error("Upload failed", "error", err)
				handlers.Error(w, r, http.StatusInternalServerError)
			}
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
			handlers.ErrorMsg(w, r, http.StatusBadRequest, "Empty file")
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
	upReq.accessKey = r.PostFormValue(handlers.ParamName)

	if r.PostFormValue("randomize") == "true" {
		upReq.randomBarename = true
	}

	upload, err := Process(r.Context(), upReq)
	if err != nil {
		var maxBytes *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytes):
			handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
		case errors.Is(err, backends.ErrFileEmpty):
			handlers.ErrorMsg(w, r, http.StatusBadRequest, "Empty file")
		default:
			slog.Error("Upload failed", "error", err)
			handlers.Error(w, r, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Vary", "Accept")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
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
		var maxBytes *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytes):
			handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
		case errors.Is(err, backends.ErrFileEmpty):
			handlers.ErrorMsg(w, r, http.StatusBadRequest, "Empty file")
		default:
			handlers.Error(w, r, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Vary", "Accept")
	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
	} else {
		_, _ = io.WriteString(w, headers.GetFileURL(r, upload.Filename).String()+"\n")
	}
}

const InputYes = "yes"

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

	upReq := Request{}
	grabURL, err := url.Parse(r.FormValue("url"))
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Invalid URL")
		return
	}
	directURL := r.FormValue("direct_url") == InputYes

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, grabURL.String(), nil)
	if err != nil {
		handlers.ErrorMsg(w, r, http.StatusInternalServerError, "Failed to create request")
		return
	}

	resp, err := http.DefaultClient.Do(req)
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

	upReq.filename = filepath.Base(grabURL.Path)
	upReq.src = http.MaxBytesReader(w, resp.Body, int64(config.Default.MaxSize))
	upReq.deleteKey = r.FormValue("deletekey")
	upReq.accessKey = r.FormValue(handlers.ParamName)
	upReq.randomBarename = r.FormValue("randomize") == InputYes
	upReq.expiry = ParseExpiry(r.FormValue("expiry"))

	upload, err := Process(r.Context(), upReq)
	if err != nil {
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
		} else {
			handlers.Error(w, r, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
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
	if r.Header.Get("Linx-Randomize") == InputYes {
		upReq.randomBarename = true
	}

	upReq.deleteKey = r.Header.Get("Linx-Delete-Key")
	upReq.accessKey = r.Header.Get(handlers.HeaderName)

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

	// Randomize the "barename" (filename without extension) if needed
	if upReq.randomBarename || len(barename) == 0 {
		barename = GenerateBarename()
		randomize = true
	}

	if len(extension) == 0 {
		var header bytes.Buffer
		header.Grow(3 * bytefmt.KiB)

		// Determine the type of file from the file header
		kind, err := mimetype.DetectReader(io.TeeReader(upReq.src, &header))
		if err != nil {
			return upload, err
		}

		upReq.src = io.MultiReader(bytes.NewReader(header.Bytes()), upReq.src)

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

	for fileexists {
		if randomize {
			barename = GenerateBarename()
		} else {
			counter, err := strconv.Atoi(string(barename[len(barename)-1]))
			if err != nil {
				barename += "1"
			} else {
				barename = barename[:len(barename)-1] + strconv.Itoa(counter+1)
			}
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

func GenerateBarename() string {
	return uniuri.NewLenChars(8, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

func GenerateJSONresponse(upload Upload, r *http.Request) []byte {
	var expiry int64
	if v := upload.Metadata.Expiry.Unix(); v > 0 {
		expiry = v
	}
	js, _ := json.Marshal(map[string]string{
		"url":        headers.GetFileURL(r, upload.Filename).String(),
		"direct_url": headers.GetSelifURL(r, upload.Filename).String(),
		"filename":   upload.Filename,
		"delete_key": upload.Metadata.DeleteKey,
		"access_key": upload.Metadata.AccessKey,
		"expiry":     strconv.FormatInt(expiry, 10),
		"size":       strconv.FormatInt(upload.Metadata.Size, 10),
		"mimetype":   upload.Metadata.Mimetype,
		"sha256sum":  upload.Metadata.Sha256sum,
	})

	return js
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
