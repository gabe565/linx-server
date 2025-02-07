package handlers

import (
	"encoding/csv"
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

const maxDisplayFileSizeBytes = 512 * bytefmt.KiB

func FileDisplay(w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	var expiryHuman string
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = strings.TrimSpace(humanize.RelTime(time.Now(), metadata.Expiry, "", ""))
	}
	sizeHuman := bytefmt.NewEncoder().SetPrecision(0).Encode(metadata.Size)
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
	case extension == "md":
		metadata, reader, err := config.StorageBackend.Get(r.Context(), fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
			return
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
	case extension == "csv":
		metadata, reader, err := config.StorageBackend.Get(r.Context(), fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
			return
		}
		defer func() {
			_ = reader.Close()
		}()

		if metadata.Size < maxDisplayFileSizeBytes {
			reader := csv.NewReader(reader)
			var content [][]string
			var columns int
			var err error
			for {
				var record []string
				reader.FieldsPerRecord = 0
				if record, err = reader.Read(); err != nil {
					break
				}
				if columns < len(record) {
					columns = len(record)
				}
				content = append(content, record)
			}
			if err == io.EOF {
				for i, record := range content {
					if len(record) != columns {
						content[i] = append(record, make([]string, columns-len(record))...)
					}
				}
				extra["Contents"] = content
				extra["Columns"] = columns
				tpl = "display/csv.html"
			}
		}
	case strings.HasPrefix(metadata.Mimetype, "text/"), util.SupportedBinExtension(extension):
		metadata, reader, err := config.StorageBackend.Get(r.Context(), fileName)
		if err != nil {
			Oops(w, r, RespHTML, err.Error())
			return
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
		return
	}
}
