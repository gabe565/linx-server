package handlers

import (
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
	metadata, err := config.StorageBackend.Head(filename)
	if err == backends.NotFoundErr {
		NotFound(w, r) // 404 - file doesn't exist
		return
	} else if err != nil {
		Unauthorized(w, r) // 401 - no metadata available
		return
	}

	if metadata.DeleteKey == requestKey {
		err := config.StorageBackend.Delete(filename)
		if err != nil {
			Oops(w, r, RespPLAIN, "Could not delete")
			return
		}

		fmt.Fprintf(w, "DELETED")
		return

	} else {
		Unauthorized(w, r) // 401 - wrong delete key
		return
	}
}
