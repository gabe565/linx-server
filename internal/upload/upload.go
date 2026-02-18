package upload

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
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
	"gabe565.com/linx-server/internal/auth/keyhash"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/csrf"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/helpers"
	"gabe565.com/linx-server/internal/util"
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
	allowZeroSize  bool
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
	AccessKey    string `json:"access_key"` //nolint:gosec
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
	siteURL := headers.GetSiteURL(r).String()
	if !csrf.StrictReferrerCheck(r, siteURL,
		[]string{"Linx-Delete-Key", "Linx-Expiry", "Linx-Randomize", "X-Requested-With"},
	) {
		handlers.Error(w, r, http.StatusBadRequest)
		return
	}

	upReq := Request{
		expiry: config.Default.MaxExpiry.Duration,
	}
	HeaderProcess(r, &upReq)

	multipart, err := r.MultipartReader()
	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) || r.Body == nil {
			handlers.ErrorMsg(w, r, http.StatusBadRequest, err.Error())
		} else {
			HandleProcessError(w, r, err)
		}
		return
	}

	for {
		part, err := multipart.NextPart()
		if err != nil {
			HandleProcessError(w, r, err)
			return
		}

		if part.FormName() == "file" {
			upReq.src = part
			upReq.filename = part.FileName()
			break
		}

		b, err := io.ReadAll(io.LimitReader(part, 32*bytefmt.KiB))
		if err != nil {
			HandleProcessError(w, r, err)
			return
		}
		_ = part.Close()

		switch part.FormName() {
		case "size":
			upReq.size, err = strconv.ParseInt(string(b), 10, 64)
			if err != nil {
				handlers.ErrorMsg(w, r, http.StatusBadRequest, "size must be an integer")
				return
			}
		case "expires":
			upReq.expiry = ParseExpiry(string(b))
		case handlers.AccessKeyParam:
			upReq.accessKey = string(b)
		case "randomize":
			upReq.randomBarename = util.ParseBool(string(b), false)
		}
	}

	if upReq.src == nil {
		HandleProcessError(w, r, backends.ErrFileEmpty)
		return
	}

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
	upReq.size = r.ContentLength

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
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, headers.GetFileURL(r, upload.Filename).String()+"\n") //nolint:gosec
	}
}

func Remote(w http.ResponseWriter, r *http.Request) {
	if config.Default.Auth.RemoteFile != "" {
		key := util.TryPathUnescape(r.FormValue("key"))
		if key == "" && config.Default.Auth.Basic {
			_, password, ok := r.BasicAuth()
			if ok {
				key = password
			}
		}
		result, err := keyhash.CheckList(config.RemoteAuthKeys, key, "")
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
		filename:      filepath.Base(grabURL.Path),
		allowZeroSize: true,
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

	resp, err := client.Do(req) //nolint:gosec
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
	upReq.size = resp.ContentLength
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

	upReq.deleteKey = util.TryPathUnescape(r.Header.Get("Linx-Delete-Key"))
	upReq.accessKey = util.TryPathUnescape(r.Header.Get(handlers.AccessKeyHeader))

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
	var randomize bool

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

	var exists, deleteKeyMatch bool
	var existingMeta backends.Metadata
	var err error
	if upReq.deleteKey == "" {
		exists, err = config.StorageBackend.Exists(ctx, upload.Filename)
		if err != nil {
			return upload, err
		}
	} else {
		existingMeta, err = config.StorageBackend.Head(ctx, upload.Filename)
		switch {
		case err == nil:
			if deleteKeyMatch, err = keyhash.CheckWithFallback(
				existingMeta.DeleteKey, upReq.deleteKey, existingMeta.Salt,
			); err != nil {
				return upload, err
			}
			exists = !deleteKeyMatch
		case errors.Is(err, backends.ErrNotFound):
			exists = false
		default:
			exists, err = config.StorageBackend.Exists(ctx, upload.Filename)
			if err != nil {
				return upload, err
			}
		}
	}

	if deleteKeyMatch && existingMeta.OriginalName != "" {
		// Keep the original filename when replacing an existing upload.
		upload.OriginalName = existingMeta.OriginalName
	}

	if !deleteKeyMatch && config.Default.ForceRandomFilename {
		randomize = true
		exists = true
	}

	origBarename := barename
	var counter int
	for exists {
		if randomize {
			barename = GenerateBarename()
		} else {
			counter++
			barename = origBarename + strconv.Itoa(counter)
		}
		upload.Filename = barename + "." + extension

		var err error
		exists, err = config.StorageBackend.Exists(ctx, upload.Filename)
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
	if _, err := fs.Stat(assets.Static(), strings.TrimPrefix(upload.Filename, "/")); err == nil || !os.IsNotExist(err) {
		return upload, ErrProhibitedFilename
	}

	// Get the rest of the metadata needed for storage
	var fileExpiry time.Time
	if upReq.expiry == 0 {
		fileExpiry = time.Time{}
	} else {
		fileExpiry = time.Now().Add(upReq.expiry)
	}

	salt := uniuri.NewLen(16)

	if upReq.deleteKey == "" {
		upReq.deleteKey = uniuri.NewLen(config.Default.RandomDeleteKeyLength)
	}
	hashedDeleteKey, err := keyhash.Hash(upReq.deleteKey, salt)
	if err != nil {
		return upload, err
	}
	storedAccessKey := upReq.accessKey
	if storedAccessKey != "" {
		if storedAccessKey, err = keyhash.Hash(storedAccessKey, salt); err != nil {
			return upload, err
		}
	}

	upload.Metadata, err = config.StorageBackend.Put(ctx, upReq.src, upload.Filename, upReq.size, backends.PutOptions{
		OriginalName: upload.OriginalName,
		Expiry:       fileExpiry,
		DeleteKey:    hashedDeleteKey,
		AccessKey:    storedAccessKey,
		Salt:         salt,
	})
	if err != nil {
		return upload, err
	}
	upload.Metadata.DeleteKey = upReq.deleteKey
	upload.Metadata.AccessKey = upReq.accessKey

	return upload, err
}

func HandleProcessError(w http.ResponseWriter, r *http.Request, err error) {
	_, isMaxBytes := errors.AsType[*http.MaxBytesError](err)
	switch {
	case isMaxBytes:
		handlers.ErrorMsg(w, r, http.StatusRequestEntityTooLarge, "File too large")
	case errors.Is(err, io.EOF):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Unexpected EOF")
	case errors.Is(err, backends.ErrFileEmpty):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Empty file")
	case errors.Is(err, ErrProhibitedFilename):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Prohibited filename")
	case errors.Is(err, io.ErrUnexpectedEOF):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Upload canceled")
	case errors.Is(err, backends.ErrSizeMismatch):
		handlers.ErrorMsg(w, r, http.StatusBadRequest, "Size mismatch")
	default:
		slog.Error("Upload failed", "error", err)
		handlers.Error(w, r, http.StatusInternalServerError)
	}
}

func GenerateBarename() string {
	return uniuri.NewLenChars(config.Default.RandomFilenameLength, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
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
		seconds, err := strconv.ParseInt(expStr, 10, 64)
		if err != nil {
			return config.Default.MaxExpiry.Duration
		}

		fileExpiry = time.Duration(seconds) * time.Second
	}

	fileExpiry = max(fileExpiry, 0)

	if config.Default.MaxExpiry.Duration == 0 {
		return fileExpiry
	} else if fileExpiry == 0 {
		return config.Default.MaxExpiry.Duration
	}

	return min(fileExpiry, config.Default.MaxExpiry.Duration)
}
