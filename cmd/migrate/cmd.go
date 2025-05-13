package migrate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/utils/must"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	FlagFrom    = "from"
	FlagTo      = "to"
	Concurrency = "concurrency"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate uploads to a new storage backend",
		Args:  cobra.NoArgs,
		RunE:  run,
	}
	config.Default.RegisterBasicFlags(cmd)
	config.RegisterBasicCompletions(cmd)

	cmd.Flags().StringP(FlagFrom, "f", "", "Source backend (one of s3, local)")
	must.Must(cmd.MarkFlagRequired(FlagFrom))

	cmd.Flags().StringP(FlagTo, "t", "", "Destination backend (one of s3, local)")
	must.Must(cmd.MarkFlagRequired(FlagTo))

	cmd.Flags().Int(Concurrency, 4, "Number of uploads to migrate in parallel")

	cmd.Flags().Lookup(config.FlagNoLogs).Usage = "Disable logging of migrated files"

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if err := config.Default.Load(cmd); err != nil {
		return err
	}

	cmd.SilenceUsage = true

	srcName := must.Must2(cmd.Flags().GetString(FlagFrom))
	srcBackend, err := newBackend(cmd.Context(), srcName)
	if err != nil {
		return err
	}

	dstName := must.Must2(cmd.Flags().GetString(FlagTo))
	dstBackend, err := newBackend(cmd.Context(), dstName)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)
	concurrency := must.Must2(cmd.Flags().GetInt(Concurrency))
	group.SetLimit(concurrency)

	for path, err := range srcBackend.List(ctx) {
		group.Go(func() error {
			if err != nil {
				return fmt.Errorf("failed to list uploads: %w", err)
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			meta, r, err := srcBackend.Get(ctx, path)
			if err != nil {
				return fmt.Errorf("failed to get upload: %w", err)
			}
			defer func() {
				_ = r.Close()
			}()

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if _, err := dstBackend.Put(ctx, r, path, meta.Size, backends.PutOptions{
				OriginalName: meta.OriginalName,
				Expiry:       meta.Expiry,
				DeleteKey:    meta.DeleteKey,
				AccessKey:    meta.AccessKey,
			}); err != nil {
				return fmt.Errorf("failed to put upload: %w", err)
			}

			if !config.Default.NoLogs {
				slog.Info("Migrated upload", "name", path)
			}
			return nil
		})
	}

	err = group.Wait()
	return err
}

var ErrUnknownBackend = errors.New("unknown backend")

func newBackend(ctx context.Context, name string) (backends.ListBackend, error) {
	switch name {
	case "s3":
		return config.Default.NewS3Backend(ctx)
	case "local":
		return config.Default.NewLocalBackend()
	}
	return nil, fmt.Errorf("%w: %s", ErrUnknownBackend, name)
}
