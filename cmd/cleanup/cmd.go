package cleanup

import (
	"gabe565.com/linx-server/internal/backends/localfs"
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

		ValidArgsFunction: cobra.NoFileCompletions,
	}
	config.Default.RegisterBasicFlags(cmd)
	config.RegisterBasicCompletions(cmd)
	cmd.Flags().Lookup(config.FlagNoLogs).Usage = "Disable logging of deleted files"
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if err := config.Default.Load(cmd); err != nil {
		return err
	}

	cmd.SilenceUsage = true

	return cleanup.Cleanup(cmd.Context(),
		localfs.New(config.Default.MetaPath, config.Default.FilesPath),
		config.Default.NoLogs,
	)
}
