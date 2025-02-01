package assets

import "embed"

//go:embed static
var Static embed.FS

//go:embed templates
var Template embed.FS
