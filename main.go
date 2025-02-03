package main

import (
	"os"

	"gabe565.com/linx-server/cmd"
	"gabe565.com/utils/cobrax"
)

//go:generate npm --prefix assets/static install
//go:generate npm --prefix assets/static run build

var version = "beta"

func main() {
	root := cmd.New(cobrax.WithVersion(version))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
