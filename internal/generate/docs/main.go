package main

import (
	"os"

	"gabe565.com/linx-server/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	const output = "./docs"

	if err := os.RemoveAll(output); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(output, 0o777); err != nil {
		panic(err)
	}

	root := cmd.New()
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, output); err != nil {
		panic(err)
	}
}
