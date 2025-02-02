package handlers

import (
	"encoding/json"
	"html/template"
	"io"
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
	"gabe565.com/utils/bytefmt"
	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const maxDisplayFileSizeBytes = 1024 * 512

func FileDisplay(w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	var expiryHuman string
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := bytefmt.Encode(metadata.Size)
	extra := make(map[string]any)
	var lines []string

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	tpl := "display/file.html"
	switch {
	case strings.EqualFold("application/json", r.Header.Get("Accept")):
		directURL := headers.GetSelifURL(r, fileName)
		js, _ := json.Marshal(map[string]string{
			"filename":   fileName,
			"direct_url": directURL.String(),
			"expiry":     strconv.FormatInt(metadata.Expiry.Unix(), 10),
			"size":       strconv.FormatInt(metadata.Size, 10),
			"mimetype":   metadata.Mimetype,
			"sha256sum":  metadata.Sha256sum,
		})
		_, _ = w.Write(js)
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
		defer func() {
			_ = reader.Close()
		}()

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := io.ReadAll(reader)
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
		defer func() {
			_ = reader.Close()
		}()

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := io.ReadAll(reader)
			if err == nil {
				unsafe := blackfriday.Run(bytes)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				extra["Contents"] = template.HTML(html) //nolint:gosec
				tpl = "display/md.html"
			}
		}

	case strings.HasPrefix(metadata.Mimetype, "text/"), util.SupportedBinExtension(extension):
		metadata, reader, err := config.StorageBackend.Get(fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
		}
		defer func() {
			_ = reader.Close()
		}()

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := io.ReadAll(reader)
			if err == nil {
				extra["Extension"] = extension
				extra["LangHL"] = util.ExtensionToHlLang(extension)
				extra["Contents"] = string(bytes)
				tpl = "display/bin.html"
			}
		}
	}

	siteURL := headers.GetSelifURL(r, fileName)
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
		"SiteURL":     strings.TrimSuffix(siteURL.String(), "/"),
	}, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
	}
}
