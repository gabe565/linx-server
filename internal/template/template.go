package template

import (
	"io"
	"net/http"
	"path"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	. "maragu.dev/gomponents"      //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

func Index(r *http.Request, opts ...OptionFunc) Node {
	u := headers.GetSiteURL(r)
	u.Path = path.Join(u.Path, r.URL.Path)

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
				Link(Rel("icon"), Href(SitePath("favicon.ico"))),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
				options.Components(),
				func() Node {
					conf, err := ConfigBytes()
					if err != nil {
						return NodeFunc(func(io.Writer) error {
							return err
						})
					}
					return Script(NodeFunc(func(w io.Writer) error {
						_, err := w.Write(conf)
						return err
					}))
				}(),
				ImportAssets(r),
			),
			Body(
				Div(ID("app")),
			),
		),
	)
}
