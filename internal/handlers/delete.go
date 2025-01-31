package handlers

import (
	"fmt"
	"net/http"

	"github.com/andreimarcu/linx-server/internal/backends"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/zenazn/goji/web"
)

func Delete(c web.C, w http.ResponseWriter, r *http.Request) {
	requestKey := r.Header.Get("Linx-Delete-Key")

	filename := c.URLParams["name"]

	// Ensure that file exists and delete key is correct
	metadata, err := config.StorageBackend.Head(filename)
	if err == backends.NotFoundErr {
		NotFound(c, w, r) // 404 - file doesn't exist
		return
	} else if err != nil {
		Unauthorized(c, w, r) // 401 - no metadata available
		return
	}

	if metadata.DeleteKey == requestKey {
		err := config.StorageBackend.Delete(filename)
		if err != nil {
			Oops(c, w, r, RespPLAIN, "Could not delete")
			return
		}

		fmt.Fprintf(w, "DELETED")
		return

	} else {
		Unauthorized(c, w, r) // 401 - wrong delete key
		return
	}
}
