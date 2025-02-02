package templates

import (
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/custompages"
	"github.com/Masterminds/sprig/v3"
)

func Load(fsys fs.FS) (map[string]*template.Template, error) {
	t := make(map[string]*template.Template)
	funcMap := sprig.HtmlFuncMap()

	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		name := strings.TrimPrefix(path, "templates/")

		switch name {
		case "base.html", "display/display.html":
			return nil
		}

		paths := []string{"templates/base.html"}
		if strings.HasPrefix(name, "display/") {
			paths = append(paths, "templates/display/display.html")
		}
		paths = append(paths, path)

		t[name], err = template.New(name).Funcs(funcMap).ParseFS(fsys, paths...)
		return err
	}); err != nil {
		return nil, err
	}

	return t, nil
}

func Render(name string, data map[string]any, r *http.Request, writer io.Writer) error {
	if data == nil {
		data = make(map[string]any)
	}

	if config.Default.SiteName == "" {
		parts := strings.Split(r.Host, ":")
		data["SiteName"] = parts[0]
	} else {
		data["SiteName"] = config.Default.SiteName
	}

	data["SitePath"] = config.Default.SiteURL.Path
	data["SelifPath"] = config.Default.SelifPath
	data["CustomPagesNames"] = custompages.Names

	switch {
	case config.Default.AuthFile == "":
		data["Auth"] = "none"
	case config.Default.BasicAuth:
		data["Auth"] = "basic"
	default:
		data["Auth"] = "header"
	}

	err := config.Templates[name].ExecuteTemplate(writer, filepath.Base(name), data)
	if err != nil {
		slog.Error("Render failed", "error", err)
	}
	return err
}
