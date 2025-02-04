package cmd

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gabe565.com/linx-server/cmd/cleanup"
	"gabe565.com/linx-server/cmd/genkey"
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
		cleanup.New(),
		genkey.New(),
	)
	config.Default.RegisterServeFlags(cmd)
	config.RegisterBasicCompletions(cmd)
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

	mux, err := server.Setup()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	if config.Default.Fastcgi {
		var listener net.Listener
		if config.Default.Bind[0] == '/' {
			// UNIX path
			listener, err = net.ListenUnix("unix", &net.UnixAddr{Name: config.Default.Bind, Net: "unix"})
			if err != nil {
				return err
			}
			defer func() {
				slog.Info("Removing FastCGI socket")
				_ = os.Remove(config.Default.Bind)
			}()
		} else {
			listener, err = net.Listen("tcp", config.Default.Bind)
			if err != nil {
				return err
			}
		}

		group.Go(func() error {
			<-ctx.Done()
			return listener.Close()
		})

		group.Go(func() error {
			log.Printf("Serving over fastcgi, bound on %s", config.Default.Bind)
			return fcgi.Serve(listener, mux)
		})
	} else {
		srv := &http.Server{
			Addr:              config.Default.Bind,
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
		}

		if config.Default.TLS.Cert != "" {
			group.Go(func() error {
				log.Printf("Serving over https, bound on %s", config.Default.Bind)
				return srv.ListenAndServeTLS(config.Default.TLS.Cert, config.Default.TLS.Key)
			})
		} else {
			group.Go(func() error {
				log.Printf("Serving over http, bound on %s", config.Default.Bind)
				return srv.ListenAndServe()
			})
		}

		group.Go(func() error {
			<-ctx.Done()
			const timeout = 10 * time.Second
			slog.Info("Gracefully shutting down", "timeout", timeout.String())
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return srv.Shutdown(ctx)
		})
	}
	err = group.Wait()
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		err = nil
	}
	return err
}
