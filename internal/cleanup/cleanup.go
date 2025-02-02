package cleanup

import (
	"context"
	"errors"
	"log"
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
				log.Printf("Failed to find metadata for %s", filename)
			}
		}

		if expiry.IsTSExpired(metadata.Expiry) {
			if !noLogs {
				log.Printf("Delete %s", filename)
			}
			if err := fileBackend.Delete(context.Background(), filename); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func PeriodicCleanup(minutes time.Duration, filesDir string, metaDir string, noLogs bool) {
	c := time.Tick(minutes)
	for range c {
		if err := Cleanup(filesDir, metaDir, noLogs); err != nil {
			slog.Error("Cleanup failed", "error", err)
		}
	}
}
