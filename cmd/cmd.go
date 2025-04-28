package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cleanupCmd "gabe565.com/linx-server/cmd/cleanup"
	"gabe565.com/linx-server/cmd/genkey"
	"gabe565.com/linx-server/cmd/migrate"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/cleanup"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/server"
	"gabe565.com/utils/cobrax"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func New(options ...cobrax.Option) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linx-server",
		Short: "Self-hosted file/media sharing website",
		Args:  cobra.NoArgs,
		RunE:  run,

		ValidArgsFunction: cobra.NoFileCompletions,
	}
	cmd.AddCommand(
		cleanupCmd.New(),
		genkey.New(),
		migrate.New(),
	)
	config.Default.RegisterServeFlags(cmd)
	config.RegisterServeCompletions(cmd)
	for _, option := range options {
		option(cmd)
	}
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if err := config.Default.Load(cmd); err != nil {
		return err
	}

	cmd.SilenceUsage = true

	slog.Info("Linx Server", "version", cobrax.GetVersion(cmd), "commit", cobrax.GetCommit(cmd))

	mux, err := server.Setup()
	if err != nil {
		return err
	}

	config.StorageBackend, err = config.Default.NewStorageBackend(cmd.Context())
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	srv := &http.Server{
		Addr:              config.Default.Bind,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		<-ctx.Done()
		timeout := config.Default.GracefulShutdown.Duration
		slog.Info("Gracefully shutting down", "timeout", timeout)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return srv.Shutdown(ctx)
	})

	group.Go(func() error {
		if config.Default.TLS.Cert != "" {
			slog.Info("Serving over https", "address", config.Default.Bind)
			return srv.ListenAndServeTLS(config.Default.TLS.Cert, config.Default.TLS.Key)
		}
		slog.Info("Serving over http", "address", config.Default.Bind)
		return srv.ListenAndServe()
	})

	if config.Default.CleanupEvery.Duration > 0 {
		if backend, ok := config.StorageBackend.(backends.ListBackend); ok {
			group.Go(func() error {
				cleanup.PeriodicCleanup(ctx, backend, config.Default.CleanupEvery.Duration, config.Default.NoLogs)
				return nil
			})
		}
	}

	err = group.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}
