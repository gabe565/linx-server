package genkey

import (
	"io"

	"gabe565.com/linx-server/internal/auth/keyhash"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genkey password",
		Short: "Generate auth file hashed keys",
		Args:  cobra.ExactArgs(1),
		RunE:  run,

		ValidArgsFunction: cobra.NoFileCompletions,
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	checkKey, err := keyhash.Hash(args[0], "")
	if err != nil {
		return err
	}

	_, err = io.WriteString(cmd.OutOrStdout(), checkKey+"\n")
	return err
}
