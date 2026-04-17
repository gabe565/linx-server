package cleanup

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gabe565.com/linx-server/internal/backends"
)

func Cleanup(ctx context.Context, backend backends.ListBackend, noLogs bool) error {
	var errs []error
	for filename, err := range backend.List(ctx) {
		switch {
		case err != nil:
			errs = append(errs, err)
			continue
		case ctx.Err() != nil:
			errs = append(errs, ctx.Err())
			return errors.Join(errs...)
		}

		metadata, err := backend.Head(ctx, filename)
		if err != nil {
			if !noLogs {
				slog.Warn("Failed to find metadata for upload", "name", filename)
			}
			continue
		}

		if metadata.Expired() {
			if !noLogs {
				slog.Info("Delete upload", "name", filename)
			}
			if err := backend.Delete(ctx, filename); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func PeriodicCleanup(ctx context.Context, backend backends.ListBackend, d time.Duration, noLogs bool) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		if err := Cleanup(ctx, backend, noLogs); err != nil {
			slog.Error("Cleanup failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
