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

	srv := &http.Server{
		Addr:              config.Default.Bind,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		var err error
		if config.Default.TLS.Cert != "" {
			slog.Info("Serving over https", "address", config.Default.Bind)
			err = srv.ListenAndServeTLS(config.Default.TLS.Cert, config.Default.TLS.Key)
		} else {
			slog.Info("Serving over http", "address", config.Default.Bind)
			err = srv.ListenAndServe()
		}
		if err != nil {
			errCh <- err
		}
	}()

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if config.Default.CleanupEvery.Duration > 0 {
		if backend, ok := config.StorageBackend.(backends.ListBackend); ok {
			go func() {
				cleanup.PeriodicCleanup(ctx, backend, config.Default.CleanupEvery.Duration, config.Default.NoLogs)
			}()
		}
	}

	select {
	case <-ctx.Done():
		timeout := config.Default.GracefulShutdown.Duration

		ctx, cancelTimeout := context.WithTimeout(context.Background(), timeout)
		defer cancelTimeout()

		ctx, cancelSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		defer cancelSignal()

		slog.Info("Gracefully stopping server", "timeout", timeout)
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
