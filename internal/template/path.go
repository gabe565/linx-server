package template

import (
	"path"

	"gabe565.com/linx-server/internal/config"
)

func SitePath(p string) string {
	if config.Default.SiteURL.Path == "" {
		return p
	}
	return path.Join(config.Default.SiteURL.Path, p)
}
