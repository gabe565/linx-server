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

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/templates"
	"gabe565.com/linx-server/internal/util"
	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

const maxDisplayFileSizeBytes = 1024 * 512

func FileDisplay(w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	var expiryHuman string
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(metadata.Size))
	extra := make(map[string]any)
	lines := []string{}

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	tpl := "display/file.html"
	switch {
	case strings.EqualFold("application/json", r.Header.Get("Accept")):
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
	case strings.HasPrefix(metadata.Mimetype, "image/"):
		tpl = "display/image.html"
	case strings.HasPrefix(metadata.Mimetype, "video/"):
		tpl = "display/video.html"
	case strings.HasPrefix(metadata.Mimetype, "audio/"):
		tpl = "display/audio.html"
	case metadata.Mimetype == "application/pdf":
		tpl = "display/pdf.html"
	case extension == "story":
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				extra["Contents"] = string(bytes)
				lines = strings.Split(string(bytes), "\n")
				tpl = "display/story.html"
			}
		}

	case extension == "md":
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
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

	case strings.HasPrefix(metadata.Mimetype, "text/"), util.SupportedBinExtension(extension):
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
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
		Oops(w, r, RespHTML, "")
	}
}
