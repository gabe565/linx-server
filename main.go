package main

import (
	"os"

	"gabe565.com/linx-server/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
