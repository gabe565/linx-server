package template

import (
	"slices"

	"gabe565.com/linx-server/internal/config"
	. "maragu.dev/gomponents"      //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

const (
	OpenGraphTitle       = "og:title"
	OpenGraphDescription = "og:description"
	OpenGraphSiteName    = "og:site_name"
	OpenGraphURL         = "og:url"
	OpenGraphType        = "og:type"
)

type Options struct {
	Title       string
	Description string
	OpenGraph   OpenGraph
}

func (o Options) Components() Group {
	group := make(Group, 0, 3)
	if o.Title != "" {
		group = append(group, TitleEl(Text(o.Title)))
		if _, ok := o.OpenGraph[OpenGraphTitle]; !ok {
			o.OpenGraph[OpenGraphTitle] = o.Title
		}
	}
	if o.Description != "" {
		group = append(group, Meta(Name("description"), Content(o.Description)))
		if _, ok := o.OpenGraph[OpenGraphDescription]; !ok {
			o.OpenGraph[OpenGraphDescription] = o.Description
		}
	}
	group = append(group, o.OpenGraph.Components())
	return group
}

type OpenGraph map[string]string

func (o OpenGraph) Components() Group {
	fields := make([]string, 0, len(o))
	for field := range o {
		fields = append(fields, field)
	}
	slices.Sort(fields)

	components := make(Group, 0, len(o))
	for _, field := range fields {
		components = append(components,
			Meta(Attr("property", field), Content(o[field])),
		)
	}
	return components
}

type OptionFunc func(o *Options)

func WithTitle(title string) OptionFunc {
	return func(o *Options) {
		if title != "" && config.Default.SiteName != "" {
			title += " · "
		}
		title += config.Default.SiteName

		o.Title = title
	}
}

func WithDescription(description string) OptionFunc {
	return func(o *Options) {
		o.Description = description
	}
}

func WithOpenGraph(k, v string) OptionFunc {
	return func(o *Options) {
		if o.OpenGraph == nil {
			o.OpenGraph = make(map[string]string)
		}
		o.OpenGraph[k] = v
	}
}
