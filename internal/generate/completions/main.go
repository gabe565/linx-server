package main

import (
	"bytes"
	"io"
	"os"
	"time"

	"gabe565.com/linx-server/cmd"
	"gabe565.com/utils/cobrax"
	flag "github.com/spf13/pflag"
)

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)

	var dateParam string
	flags.StringVar(&dateParam, "date", time.Now().Format(time.RFC3339), "Build date")

	if err := flags.Parse(os.Args); err != nil {
		panic(err)
	}

	date, err := time.Parse(time.RFC3339, dateParam)
	if err != nil {
		panic(err)
	}

	if err := os.RemoveAll("completions"); err != nil {
		panic(err)
	}

	if err := os.Mkdir("completions", 0o777); err != nil {
		panic(err)
	}

	root, err := os.OpenRoot("completions")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = root.Close()
	}()

	rootCmd := cmd.New()
	name := rootCmd.Name()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	for _, shell := range []cobrax.Shell{cobrax.Bash, cobrax.Zsh, cobrax.Fish} {
		if err := cobrax.GenCompletion(rootCmd, shell); err != nil {
			panic(err)
		}

		outPath := name + "." + string(shell)

		f, err := root.Create(outPath)
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(f, &buf); err != nil {
			panic(err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}

		if err := root.Chtimes(outPath, date, date); err != nil {
			panic(err)
		}
	}
}
