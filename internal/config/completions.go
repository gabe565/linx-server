package config

import (
	"errors"

	"gabe565.com/utils/bytefmt"
	"gabe565.com/utils/must"
	"github.com/spf13/cobra"
)

func RegisterBasicCompletions(cmd *cobra.Command) {
	must.Must(errors.Join(
		cmd.RegisterFlagCompletionFunc(FlagConfig, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"toml"}, cobra.ShellCompDirectiveFilterFileExt
		}),
		cmd.RegisterFlagCompletionFunc(FlagFilesPath, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}),
		cmd.RegisterFlagCompletionFunc(FlagMetaPath, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}),
	))
}

func RegisterServeCompletions(cmd *cobra.Command) {
	must.Must(errors.Join(
		cmd.RegisterFlagCompletionFunc(FlagBind, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"127.0.0.1:8080\tPrivate on port 8080", ":8080\tPublic on port 8080"}, cobra.ShellCompDirectiveNoFileComp
		}),
		cmd.RegisterFlagCompletionFunc(FlagSiteName, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(FlagSiteURL, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"https://"}, cobra.ShellCompDirectiveNoFileComp
		}),
		cmd.RegisterFlagCompletionFunc(FlagSelifPath, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"selif"}, cobra.ShellCompDirectiveNoFileComp
		}),
		cmd.RegisterFlagCompletionFunc(FlagMaxSize, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			sizes := []int64{
				10 * bytefmt.MiB,
				50 * bytefmt.MiB,
				100 * bytefmt.MiB,
				250 * bytefmt.MiB,
				1 * bytefmt.GiB,
				4 * bytefmt.GiB,
			}
			s := make([]string, 0, len(sizes))
			encoder := bytefmt.NewEncoder().SetUseSpace(false)
			for _, size := range sizes {
				s = append(s, encoder.Encode(size))
			}
			return s, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}),
		cmd.RegisterFlagCompletionFunc(FlagMaxExpiry, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{
				"1m\t1 minute",
				"5m\t5 minutes",
				"1h\t1 hour",
				"24h\t1 day",
				"168h\t1 week",
				"744h\t1 month",
				"8760h\t1 year",
			}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}),
		cmd.RegisterFlagCompletionFunc(FlagTLSCert, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"crt"}, cobra.ShellCompDirectiveFilterFileExt
		}),
		cmd.RegisterFlagCompletionFunc(FlagTLSKey, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"key"}, cobra.ShellCompDirectiveFilterFileExt
		}),
		cmd.RegisterFlagCompletionFunc(FlagS3Endpoint, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"https://"}, cobra.ShellCompDirectiveNoFileComp
		}),
		cmd.RegisterFlagCompletionFunc(FlagS3Region, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(FlagS3Bucket, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(FlagAuthCookieExpiry, cobra.NoFileCompletions),
		cmd.RegisterFlagCompletionFunc(FlagCustomPagesDir, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}),
		cmd.RegisterFlagCompletionFunc(FlagCleanupEvery, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"1h", "12h", "24h"}, cobra.ShellCompDirectiveNoFileComp
		}),
	))
}
