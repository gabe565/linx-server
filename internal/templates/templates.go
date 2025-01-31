package templates

import (
	"io"
	"net/http"
	"strings"

	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/custompages"
	"github.com/flosch/pongo2"
)

func Render(tpl *pongo2.Template, context pongo2.Context, r *http.Request, writer io.Writer) error {
	if config.Default.SiteName == "" {
		parts := strings.Split(r.Host, ":")
		context["sitename"] = parts[0]
	} else {
		context["sitename"] = config.Default.SiteName
	}

	context["sitepath"] = config.Default.SitePath
	context["selifpath"] = config.Default.SelifPath
	context["custom_pages_names"] = custompages.Names

	var a string
	if config.Default.AuthFile == "" {
		a = "none"
	} else if config.Default.BasicAuth {
		a = "basic"
	} else {
		a = "header"
	}
	context["auth"] = a

	return tpl.ExecuteWriter(context, writer)
}
