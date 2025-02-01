package custompages

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

//nolint:gochecknoglobals
var (
	CustomPages map[string]string
	Names       map[string]string
)

func InitializeCustomPages(customPagesDir string) {
	files, err := os.ReadDir(customPagesDir)
	if err != nil {
		log.Fatal("Error reading the custom pages directory: ", err)
	}
	if len(files) == 0 {
		return
	}

	CustomPages = make(map[string]string, len(files))
	Names = make(map[string]string, len(files))

	for _, file := range files {
		fileName := file.Name()

		if len(fileName) <= 3 {
			continue
		}

		if strings.EqualFold(fileName[len(fileName)-3:], ".md") {
			contents, err := os.ReadFile(path.Join(customPagesDir, fileName))
			if err != nil {
				log.Fatalf("Error reading file %s", fileName)
			}

			unsafe := blackfriday.Run(contents)
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

			fileName := fileName[0 : len(fileName)-3]
			CustomPages[fileName] = string(html)
			Names[fileName] = strings.ReplaceAll(fileName, "_", " ")
		}
	}
}
