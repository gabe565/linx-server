package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gabe565.com/linx-server/internal/config"
)

// Burn deletes the file and its claim.
func Burn(ctx context.Context, fileName string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Minute)
	defer cancel()

	if err := config.StorageBackend.Delete(ctx, fileName); err != nil {
		slog.Error("Failed to delete burn-after-read file", "path", fileName, "error", err)
	}
}

// prepareBurnRequest forces a full response, since there won't be a second one.
func prepareBurnRequest(r *http.Request) {
	for _, h := range []string{"Range", "If-Range", "If-Match", "If-None-Match", "If-Modified-Since", "If-Unmodified-Since"} {
		r.Header.Del(h)
	}
}
