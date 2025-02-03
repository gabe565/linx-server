package assets

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"strings"

	"gabe565.com/utils/must"
)

//go:embed static/dist
var static embed.FS

func Static() fs.FS {
	return must.Must2(fs.Sub(static, "static/dist"))
}

//go:embed templates
var templates embed.FS

func Templates() fs.FS {
	return must.Must2(fs.Sub(templates, "templates"))
}

type ManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

type ManifestMap map[string]ManifestEntry

func (m ManifestMap) ImportCSS(sitePath, name string) template.HTML {
	if entry, ok := m[name]; ok {
		var s strings.Builder
		for _, e := range entry.CSS {
			s.WriteString(`<link href="` + sitePath + e + `" rel="stylesheet" type="text/css">`)
		}
		return template.HTML(s.String()) //nolint:gosec
	}
	return ""
}

func (m ManifestMap) ImportJS(sitePath, name string) template.HTML {
	if entry, ok := m[name]; ok {
		return template.HTML(`<script type="module" src="` + sitePath + entry.File + `"></script>`) //nolint:gosec
	}
	return ""
}

func (m ManifestMap) PreloadJS(sitePath, name string) template.HTML {
	if entry, ok := m[name]; ok {
		return template.HTML(`<script rel="modulepreload" src="` + sitePath + entry.File + `"></script>`) //nolint:gosec
	}
	return ""
}

//nolint:gochecknoglobals
var Manifest ManifestMap

func LoadManifest() error {
	f, err := Static().Open("manifest.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	return json.NewDecoder(f).Decode(&Manifest)
}
