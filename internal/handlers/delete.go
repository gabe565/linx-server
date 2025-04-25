package handlers

import (
	"errors"
	"io"
	"net/http"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"github.com/go-chi/chi/v5"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	requestKey := r.Header.Get("Linx-Delete-Key")

	filename := chi.URLParam(r, "name")

	// Ensure that file exists and delete key is correct
	metadata, err := config.StorageBackend.Head(r.Context(), filename)
	if err != nil {
		if errors.Is(err, backends.ErrNotFound) {
			ErrorMsg(w, r, http.StatusNotFound, "File not found") // 404 - file doesn't exist
		} else {
			Error(w, r, http.StatusUnauthorized) // 401 - no metadata available
		}
		return
	}

	if metadata.DeleteKey != requestKey {
		Error(w, r, http.StatusUnauthorized) // 401 - wrong delete key
		return
	}

	if err := config.StorageBackend.Delete(r.Context(), filename); err != nil {
		Error(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Vary", "Accept, Linx-Delete-Key")
	_, _ = io.WriteString(w, "DELETED\n")
}
