package template

import (
	"net/http"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	. "maragu.dev/gomponents"      //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

func Index(r *http.Request, opts ...OptionFunc) Node {
	u := headers.GetSiteURL(r)
	u.Path = r.URL.Path

	options := Options{
		Title:       config.Default.SiteName,
		Description: "Self-hosted file/media sharing website.",
		OpenGraph: map[string]string{
			OpenGraphSiteName: config.Default.SiteName,
			OpenGraphURL:      u.String(),
			OpenGraphType:     "website",
		},
	}

	for _, o := range opts {
		o(&options)
	}

	return Doctype(
		HTML(
			Head(
				Meta(Charset("UTF-8")),
				Link(Rel("icon"), Href("/favicon.ico")),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
				options.Components(),
				Script(Attr("integrity", ConfigHash()), Raw(ConfigString())),
				ImportAssets(),
			),
			Body(
				Div(ID("app")),
			),
		),
	)
}
