package genkey

import (
	"encoding/base64"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/scrypt"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genkey password",
		Short: "Generate auth file hashed keys",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
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
	checkKey, err := scrypt.Key([]byte(args[0]), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
	if err != nil {
		return err
	}

	_, err = io.WriteString(cmd.OutOrStdout(), base64.StdEncoding.EncodeToString(checkKey))
	return err
}
