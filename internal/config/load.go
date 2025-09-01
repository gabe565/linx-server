package config

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

func (c *Config) Load(cmd *cobra.Command) error {
	k := koanf.New(".")

	// Load default config
	if err := k.Load(structs.Provider(c, "toml"), nil); err != nil {
		return err
	}

	// Find config file
	cfgFile := must.Must2(cmd.Flags().GetString(FlagConfig))
	if cfgFile == "" {
		var err error
		cfgFile, err = getDefaultFile()
		if err != nil {
			return err
		}
	}
	if strings.Contains(cfgFile, "$HOME") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		cfgFile = strings.Replace(cfgFile, "$HOME", home, 1)
	}

	// Load config file
	cfgContents, err := os.ReadFile(cfgFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(cfgContents) != 0 {
		if err := k.Load(rawbytes.Provider(cfgContents), TOMLParser{}); err != nil {
			return err
		}
	}

	// Load envs
	const envPrefix = "LINX_"
	nested := []string{"tls", "auth", "s3", "limit", "header"}
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(k, envPrefix)
			k = strings.ToLower(k)
			k = strings.ReplaceAll(k, "_", "-")
			for _, name := range nested {
				if strings.HasPrefix(k, name) {
					k = strings.Replace(k, name+"-", name+".", 1)
					break
				}
			}
			return k, v
		},
	}), nil); err != nil {
		return err
	}

	// Load flags
	if err := k.Load(posflag.ProviderWithValue(cmd.Flags(), ".", k, func(key string, value string) (string, any) {
		key = strings.ReplaceAll(key, "-", "_")
		return key, value
	}), nil); err != nil {
		return err
	}

	return k.UnmarshalWithConf("", c, koanf.UnmarshalConf{Tag: "toml"})
}
