package assets

import (
	"embed"
	"io/fs"

	"gabe565.com/utils/must"
)

//go:embed static/dist
var static embed.FS

func Static() fs.FS {
	return must.Must2(fs.Sub(static, "static/dist"))
}
