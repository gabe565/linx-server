package main

import (
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"gabe565.com/linx-server/cmd"
	"github.com/spf13/cobra/doc"
	flag "github.com/spf13/pflag"
)

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)

	var version string
	flags.StringVar(&version, "version", "beta", "Version")

	var dateParam string
	flags.StringVar(&dateParam, "date", time.Now().Format(time.RFC3339), "Build date")

	if err := flags.Parse(os.Args); err != nil {
		panic(err)
	}

	if err := os.RemoveAll("manpages"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("manpages", 0o755); err != nil {
		panic(err)
	}

	root, err := os.OpenRoot("manpages")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = root.Close()
	}()

	rootCmd := cmd.New()
	rootName := rootCmd.Name()

	date, err := time.Parse(time.RFC3339, dateParam)
	if err != nil {
		panic(err)
	}

	header := doc.GenManHeader{
		Title:   strings.ToUpper(rootName),
		Section: "1",
		Date:    &date,
		Source:  rootName + " " + version,
		Manual:  "User Commands",
	}

	if err := doc.GenManTree(rootCmd, &header, "manpages"); err != nil {
		panic(err)
	}

	if err := fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		in, err := root.Open(path)
		if err != nil {
			return err
		}

		outPath := path + ".gz"

		out, err := root.Create(outPath)
		if err != nil {
			return err
		}
		gz := gzip.NewWriter(out)

		if _, err := io.Copy(gz, in); err != nil {
			return err
		}

		if err := in.Close(); err != nil {
			return err
		}
		if err := root.Remove(path); err != nil {
			return err
		}

		if err := gz.Close(); err != nil {
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}

		if err := root.Chtimes(outPath, date, date); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}
}
