package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

var ErrNoCustomPages = errors.New("no custom pages found")

func ListCustomPages(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	customPages := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if strings.EqualFold(ext, ".md") {
			customPages = append(customPages, strings.TrimSuffix(file.Name(), ext))
		}
	}

	if len(customPages) == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoCustomPages)
	}
	return customPages, nil
}

func CustomPage(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name") + ".md"

		root, err := os.OpenRoot(dir)
		if err != nil {
			Error(w, r, http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = root.Close()
		}()

		w.Header().Set("Cache-Control", "public, no-cache")
		http.ServeFileFS(w, r, root.FS(), name)
	}
}
