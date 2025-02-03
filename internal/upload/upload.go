package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"github.com/dchest/uniuri"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
)

//nolint:gochecknoglobals
var (
	ErrFileTooLarge = errors.New("file too large")
	fileDenylist    = []string{
		"favicon.ico",
		"crossdomain.xml",
	}
)

// Describes metadata directly from the user request
type Request struct {
	src            io.Reader
	size           int64
	filename       string
	expiry         time.Duration // Seconds until expiry, 0 = never
	deleteKey      string        // Empty string if not defined
	randomBarename bool
	accessKey      string // Empty string if not defined
}

// Metadata associated with a file as it would actually be stored
type Upload struct {
	Filename string // Final filename on disk
	Metadata backends.Metadata
}

func POSTHandler(w http.ResponseWriter, r *http.Request) {
	siteURL := headers.GetSiteURL(r).String()
	if !csrf.StrictReferrerCheck(r, siteURL, []string{"Linx-Delete-Key", "Linx-Expiry", "Linx-Randomize", "X-Requested-With"}) {
		handlers.BadRequest(w, r, handlers.RespAUTO, "")
		return
	}

	upReq := Request{}
	HeaderProcess(r, &upReq)

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, headers, err := r.FormFile("file")
		if err != nil {
			handlers.Oops(w, r, handlers.RespHTML, "Could not upload file.")
			return
		}
		defer file.Close()

		upReq.src = file
		upReq.size = headers.Size
		upReq.filename = headers.Filename
	} else {
		if r.PostFormValue("content") == "" {
			handlers.BadRequest(w, r, handlers.RespAUTO, "Empty file")
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

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			if errors.Is(err, ErrFileTooLarge) || errors.Is(err, backends.ErrFileEmpty) {
				handlers.BadRequest(w, r, handlers.RespJSON, err.Error())
				return
			}
			handlers.Oops(w, r, handlers.RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = w.Write(js)
	} else {
		if err != nil {
			if errors.Is(err, ErrFileTooLarge) || errors.Is(err, backends.ErrFileEmpty) {
				handlers.BadRequest(w, r, handlers.RespHTML, err.Error())
				return
			}
			handlers.Oops(w, r, handlers.RespHTML, "Could not upload file: "+err.Error())
			return
		}

		http.Redirect(w, r, headers.GetFileURL(r, upload.Filename).String(), http.StatusSeeOther)
	}
}

func PUTHandler(w http.ResponseWriter, r *http.Request) {
	upReq := Request{}
	HeaderProcess(r, &upReq)

	defer r.Body.Close()
	upReq.filename = chi.URLParam(r, "name")
	upReq.src = http.MaxBytesReader(w, r.Body, int64(config.Default.MaxSize))

	upload, err := Process(r.Context(), upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			if errors.Is(err, ErrFileTooLarge) || errors.Is(err, backends.ErrFileEmpty) {
				handlers.BadRequest(w, r, handlers.RespJSON, err.Error())
				return
			}
			handlers.Oops(w, r, handlers.RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = w.Write(js)
	} else {
		if err != nil {
			if errors.Is(err, ErrFileTooLarge) || errors.Is(err, backends.ErrFileEmpty) {
				handlers.BadRequest(w, r, handlers.RespPLAIN, err.Error())
				return
			}
			handlers.Oops(w, r, handlers.RespPLAIN, "Could not upload file: "+err.Error())
			return
		}

		fmt.Fprintf(w, "%s\n", headers.GetFileURL(r, upload.Filename))
	}
}

const InputYes = "yes"

func Remote(w http.ResponseWriter, r *http.Request) {
	if config.Default.RemoteAuthFile != "" {
		key := r.FormValue("key")
		if key == "" && config.Default.BasicAuth {
			_, password, ok := r.BasicAuth()
			if ok {
				key = password
			}
		}
		result, err := apikeys.CheckAuth(config.RemoteAuthKeys, key)
		if err != nil || !result {
			if config.Default.BasicAuth {
				rs := ""
				if config.Default.SiteName != "" {
					rs = " realm=" + strconv.Quote(config.Default.SiteName)
				}
				w.Header().Set("WWW-Authenticate", `Basic`+rs)
			}
			handlers.Unauthorized(w, r)
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
		handlers.Oops(w, r, handlers.RespAUTO, "Invalid URL: "+err.Error())
		return
	}
	directURL := r.FormValue("direct_url") == InputYes

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, grabURL.String(), nil)
	if err != nil {
		handlers.Oops(w, r, handlers.RespAUTO, err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		handlers.Oops(w, r, handlers.RespAUTO, "Could not retrieve URL")
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	upReq.filename = filepath.Base(grabURL.Path)
	upReq.src = http.MaxBytesReader(w, resp.Body, int64(config.Default.MaxSize))
	upReq.deleteKey = r.FormValue("deletekey")
	upReq.accessKey = r.FormValue(handlers.ParamName)
	upReq.randomBarename = r.FormValue("randomize") == InputYes
	upReq.expiry = ParseExpiry(r.FormValue("expiry"))

	upload, err := Process(r.Context(), upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			handlers.Oops(w, r, handlers.RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := GenerateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = w.Write(js)
	} else {
		if err != nil {
			handlers.Oops(w, r, handlers.RespHTML, "Could not upload file: "+err.Error())
			return
		}

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
		return upload, ErrFileTooLarge
	}

	// Determine the appropriate filename
	barename, extension := BarePlusExt(upReq.filename)
	randomize := false

	// Randomize the "barename" (filename without extension) if needed
	if upReq.randomBarename || len(barename) == 0 {
		barename = GenerateBarename()
		randomize = true
	}

	var header []byte
	if len(extension) == 0 {
		// Pull the first 512 bytes off for use in MIME detection
		header = make([]byte, 512)
		n, _ := upReq.src.Read(header)
		if n == 0 {
			return upload, backends.ErrFileEmpty
		}
		header = header[:n]

		// Determine the type of file from header
		kind := mimetype.Detect(header)
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
		fileExpiry = expiry.NeverExpire
	} else {
		fileExpiry = time.Now().Add(upReq.expiry)
	}

	if upReq.deleteKey == "" {
		upReq.deleteKey = uniuri.NewLen(30)
	}

	upload.Metadata, err = config.StorageBackend.Put(ctx, upload.Filename, io.MultiReader(bytes.NewReader(header), upReq.src), fileExpiry, upReq.deleteKey, upReq.accessKey)
	if err != nil {
		return upload, err
	}

	return upload, err
}

func GenerateBarename() string {
	return uniuri.NewLenChars(8, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

func GenerateJSONresponse(upload Upload, r *http.Request) []byte {
	js, _ := json.Marshal(map[string]string{
		"url":        headers.GetFileURL(r, upload.Filename).String(),
		"direct_url": headers.GetSelifURL(r, upload.Filename).String(),
		"filename":   upload.Filename,
		"delete_key": upload.Metadata.DeleteKey,
		"access_key": upload.Metadata.AccessKey,
		"expiry":     strconv.FormatInt(upload.Metadata.Expiry.Unix(), 10),
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

	fileExpiry, err := strconv.ParseUint(expStr, 10, 64)
	if err != nil {
		return config.Default.MaxExpiry.Duration
	}

	if config.Default.MaxExpiry.Duration > 0 && (fileExpiry > config.Default.MaxExpirySeconds() || fileExpiry == 0) {
		fileExpiry = config.Default.MaxExpirySeconds()
	}
	return time.Duration(fileExpiry) * time.Second //nolint:gosec
}
