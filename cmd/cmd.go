package cmd

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"syscall"

	"gabe565.com/linx-server/cmd/cleanup"
	"gabe565.com/linx-server/cmd/genkey"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/server"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linx-server",
		Short: "Self-hosted file/media sharing website",
		RunE:  run,
	}
	cmd.AddCommand(
		cleanup.New(),
		genkey.New(),
	)
	config.Default.RegisterServeFlags(cmd)
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

	if config.Default.Fastcgi {
		var listener net.Listener
		if config.Default.Bind[0] == '/' {
			// UNIX path
			listener, err = net.ListenUnix("unix", &net.UnixAddr{Name: config.Default.Bind, Net: "unix"})
			if err != nil {
				return err
			}
			cleanup := func() {
				slog.Info("Removing FastCGI socket")
				os.Remove(config.Default.Bind)
			}
			defer cleanup()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				slog.Info("Got signal", "value", sig)
				cleanup()
				os.Exit(0)
			}()
		} else {
			listener, err = net.Listen("tcp", config.Default.Bind)
			if err != nil {
				return err
			}
		}

		log.Printf("Serving over fastcgi, bound on %s", config.Default.Bind)
		fcgi.Serve(listener, mux)
	} else if config.Default.TLSCert != "" {
		log.Printf("Serving over https, bound on %s", config.Default.Bind)
		err := http.ListenAndServeTLS(config.Default.Bind, config.Default.TLSCert, config.Default.TLSKey, mux)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Serving over http, bound on %s", config.Default.Bind)
		err := http.ListenAndServe(config.Default.Bind, mux)
		if err != nil {
			return err
		}
	}
	return nil
}
