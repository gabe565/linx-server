package template

import (
	"encoding/json"
	"io"
	"net/http"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/config"
	. "maragu.dev/gomponents"      //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

type ManifestEntry struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

type ManifestMap map[string]ManifestEntry

func (m ManifestMap) Import(name string) Node {
	if entry, ok := m[name]; ok {
		g := make(Group, 0, len(entry.CSS)+1)
		g = append(g, Script(Type("module"), CrossOrigin(""), Src(SitePath(entry.File))))
		for _, entry := range entry.CSS {
			g = append(g, Link(Rel("stylesheet"), CrossOrigin(""), Href(SitePath(entry))))
		}
		return g
	}
	return NodeFunc(func(io.Writer) error { return nil })
}

func (m ManifestMap) Preload(name string) Node {
	if entry, ok := m[name]; ok {
		g := make(Group, 0, len(entry.Imports)+1)
		g = append(g, Link(Rel("modulepreload"), CrossOrigin(""), Href(SitePath(entry.File))))
		for _, srcPath := range entry.Imports {
			if entry, ok := m[srcPath]; ok {
				g = append(g, Link(Rel("modulepreload"), CrossOrigin(""), Href(SitePath(entry.File))))
			}
		}
		return g
	}
	return NodeFunc(func(io.Writer) error { return nil })
}

//nolint:gochecknoglobals
var manifest ManifestMap

func LoadManifest() error {
	f, err := assets.Static().Open(".vite/manifest.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	if err := json.NewDecoder(f).Decode(&manifest); err != nil {
		return err
	}

	return nil
}

func ImportAssets(r *http.Request) Node {
	if u := config.Default.ViteURL; u != "" {
		return Group{
			Script(Type("module"), Src(u+"/@vite/client")),
			Script(Type("module"), Src(u+"/src/main.js")),
		}
	}

	var preload Node
	if r.URL.Path == "/api" {
		preload = manifest.Preload("src/views/APIView.vue")
	}

	return Group{preload, manifest.Import("src/main.js")}
}
