package genkey

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/scrypt"
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

const (
	scryptSalt   = "linx-server"
	scryptN      = 16384
	scryptr      = 8
	scryptp      = 1
	scryptKeyLen = 32
)

func run(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	checkKey, err := scrypt.Key([]byte(args[0]), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
	if err != nil {
		return err
	}

	buf := make([]byte, 0, base64.StdEncoding.EncodedLen(len(checkKey))+1)
	buf = base64.StdEncoding.AppendEncode(buf, checkKey)
	buf = append(buf, '\n')

	_, err = cmd.OutOrStdout().Write(buf)
	return err
}
