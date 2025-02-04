package config

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/knadh/koanf/providers/env"
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
	nested := []string{"limit"}
	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		s = strings.TrimPrefix(s, envPrefix)
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", "-")
		for _, name := range nested {
			if strings.HasPrefix(s, name) {
				s = strings.Replace(s, name+"-", name+".", 1)
				break
			}
		}
		return s
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
