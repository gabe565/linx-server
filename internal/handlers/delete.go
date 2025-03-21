package handlers

import (
	"errors"
	"fmt"
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
			NotFound(w, r) // 404 - file doesn't exist
			return
		}
		Unauthorized(w, r) // 401 - no metadata available
		return
	}

	if metadata.DeleteKey == requestKey {
		err := config.StorageBackend.Delete(r.Context(), filename)
		if err != nil {
			Oops(w, r, RespPLAIN, "Could not delete")
			return
		}

		_, _ = fmt.Fprintf(w, "DELETED")
		return
	}

	Unauthorized(w, r) // 401 - wrong delete key
}
