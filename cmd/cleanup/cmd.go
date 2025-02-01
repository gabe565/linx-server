package cleanup

import (
	"gabe565.com/linx-server/internal/cleanup"
	"gabe565.com/linx-server/internal/config"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Manually clean up expired files",
		Args:  cobra.NoArgs,
		RunE:  run,
	}
	config.Default.RegisterBasicFlags(cmd)
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if err := config.Default.Load(cmd); err != nil {
		return err
	}

	cmd.SilenceUsage = true

	return cleanup.Cleanup(config.Default.FilesDir, config.Default.MetaDir, config.Default.NoLogs)
}
