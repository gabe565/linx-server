package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/internal/backends"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/expiry"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/andreimarcu/linx-server/internal/templates"
	"github.com/andreimarcu/linx-server/internal/util"
	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/zenazn/goji/web"
)

const maxDisplayFileSizeBytes = 1024 * 512

func FileDisplay(c web.C, w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	var expiryHuman string
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(metadata.Size))
	extra := make(map[string]any)
	lines := []string{}

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename":   fileName,
			"direct_url": headers.GetSiteURL(r) + config.Default.SelifPath + fileName,
			"expiry":     strconv.FormatInt(metadata.Expiry.Unix(), 10),
			"size":       strconv.FormatInt(metadata.Size, 10),
			"mimetype":   metadata.Mimetype,
			"sha256sum":  metadata.Sha256sum,
		})
		w.Write(js)
		return
	}

	var tpl string

	if strings.HasPrefix(metadata.Mimetype, "image/") {
		tpl = "display/image.html"
	} else if strings.HasPrefix(metadata.Mimetype, "video/") {
		tpl = "display/video.html"
	} else if strings.HasPrefix(metadata.Mimetype, "audio/") {
		tpl = "display/audio.html"
	} else if metadata.Mimetype == "application/pdf" {
		tpl = "display/pdf.html"
	} else if extension == "story" {
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				extra["Contents"] = string(bytes)
				lines = strings.Split(string(bytes), "\n")
				tpl = "display/story.html"
			}
		}

	} else if extension == "md" {
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				unsafe := blackfriday.MarkdownCommon(bytes)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				extra["Contents"] = template.HTML(html)
				tpl = "display/md.html"
			}
		}

	} else if strings.HasPrefix(metadata.Mimetype, "text/") || util.SupportedBinExtension(extension) {
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				extra["Extension"] = extension
				extra["LangHL"] = util.ExtensionToHlLang(extension)
				extra["Contents"] = string(bytes)
				tpl = "display/bin.html"
			}
		}
	}

	// Catch other files
	if tpl == "" {
		tpl = "display/file.html"
	}

	err := templates.Render(tpl, map[string]any{
		"MIME":        metadata.Mimetype,
		"FileName":    fileName,
		"Size":        sizeHuman,
		"Expiry":      expiryHuman,
		"ExpiryList":  expiry.ListExpirationTimes(),
		"Extra":       extra,
		"ForceRandom": config.Default.ForceRandomFilename,
		"Lines":       lines,
		"Files":       metadata.ArchiveFiles,
		"SiteURL":     strings.TrimSuffix(headers.GetSiteURL(r), "/"),
	}, r, w)

	if err != nil {
		Oops(c, w, r, RespHTML, "")
	}
}
