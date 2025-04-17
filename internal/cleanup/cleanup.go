package cleanup

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gabe565.com/linx-server/internal/backends/localfs"
	"gabe565.com/linx-server/internal/expiry"
)

func Cleanup(filesDir string, metaDir string, noLogs bool) error {
	fileBackend := localfs.NewLocalfsBackend(metaDir, filesDir)

	files, err := fileBackend.List(context.Background())
	if err != nil {
		return err
	}

	var errs []error
	for _, filename := range files {
		metadata, err := fileBackend.Head(context.Background(), filename)
		if err != nil {
			if !noLogs {
				slog.Warn("Failed to find metadata for upload", "name", filename)
			}
		}

		if expiry.IsTSExpired(metadata.Expiry) {
			if !noLogs {
				slog.Info("Delete upload", "name", filename)
			}
			if err := fileBackend.Delete(context.Background(), filename); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func PeriodicCleanup(d time.Duration, filesDir string, metaDir string, noLogs bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		ticker.Reset(d)
		if err := Cleanup(filesDir, metaDir, noLogs); err != nil {
			slog.Error("Cleanup failed", "error", err)
		}
	}
}
