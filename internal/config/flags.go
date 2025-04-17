package config

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	FlagConfig              = "config"
	FlagBind                = "bind"
	FlagFilesPath           = "files-path"
	FlagMetaPath            = "meta-path"
	FlagNoLogs              = "no-logs"
	FlagAuthBasic           = "auth-basic"
	FlagAllowHotlink        = "allow-hotlink"
	FlagSiteName            = "site-name"
	FlagSiteURL             = "site-url"
	FlagSelifPath           = "selif-path"
	FlagMaxSize             = "max-size"
	FlagMaxExpiry           = "max-expiry"
	FlagTLSCert             = "tls-cert"
	FlagTLSKey              = "tls-key"
	FlagRealIP              = "real-ip"
	FlagFastcgi             = "fastcgi"
	FlagRemoteUploads       = "remote-uploads"
	FlagAuthFile            = "auth-file"
	FlagAuthRemoteFile      = "auth-remote-file"
	FlagNoDirectAgents      = "no-direct-agents"
	FlagS3Endpoint          = "s3-endpoint"
	FlagS3Region            = "s3-region"
	FlagS3Bucket            = "s3-bucket"
	FlagS3ForcePathStyle    = "s3-force-path-style"
	FlagForceRandomFilename = "force-random-filename"
	FlagAuthCookieExpiry    = "auth-cookie-expiry"
	FlagCustomPagesDir      = "custom-pages-path"
	FlagCleanupEvery        = "cleanup-every"
)

func (c *Config) RegisterBasicFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	confPath := os.Getenv("LINX_CONFIG")
	if confPath == "" {
		confPath, _ = getDefaultFile()
		if home, err := os.UserHomeDir(); err == nil && home != "/" {
			confPath = strings.Replace(confPath, home, "$HOME", 1)
		}
	}
	fs.StringP(FlagConfig, "c", confPath, "Path to the config file")
	fs.StringVar(&c.FilesPath, FlagFilesPath, c.FilesPath, "Path to files directory")
	fs.StringVar(&c.MetaPath, FlagMetaPath, c.MetaPath, "Path to metadata directory")
	fs.BoolVar(&c.NoLogs, FlagNoLogs, c.NoLogs, "Remove logging of each request")
}

func (c *Config) RegisterServeFlags(cmd *cobra.Command) {
	c.RegisterBasicFlags(cmd)

	fs := cmd.Flags()
	fs.StringVar(&c.Bind, FlagBind, c.Bind, "Host to bind to")
	fs.BoolVar(&c.Auth.Basic, FlagAuthBasic, c.Auth.Basic, "Allow logging in with basic auth password")
	fs.BoolVar(&c.AllowHotlink, FlagAllowHotlink, c.AllowHotlink, "Allow hot-linking of files")
	fs.StringVar(&c.SiteName, FlagSiteName, c.SiteName, "Name of the site")
	fs.Var(&c.SiteURL, FlagSiteURL, "Site base url")
	fs.StringVar(&c.SelifPath, FlagSelifPath, c.SelifPath,
		"Path relative to site base url where files are accessed directly",
	)
	fs.Var(&c.MaxSize, FlagMaxSize, "Maximum upload file size in bytes")
	fs.DurationVar(&c.MaxExpiry.Duration, FlagMaxExpiry, c.MaxExpiry.Duration,
		"Maximum expiration time. A value of 0 means no expiry.",
	)
	fs.StringVar(&c.TLS.Cert, FlagTLSCert, c.TLS.Cert, "Path to ssl certificate (for https)")
	fs.StringVar(&c.TLS.Key, FlagTLSKey, c.TLS.Key, "Path to ssl key (for https)")
	fs.BoolVar(&c.Header.RealIP, FlagRealIP, c.Header.RealIP, "Use X-Real-IP/X-Forwarded-For headers")
	fs.BoolVar(&c.Fastcgi, FlagFastcgi, c.Fastcgi, "Serve through fastcgi")
	fs.BoolVar(&c.RemoteUploads, FlagRemoteUploads, c.RemoteUploads, "Enable remote uploads (/upload?url=https://...)")
	fs.StringVar(&c.Auth.File, FlagAuthFile, c.Auth.File,
		"Path to a file containing newline-separated scrypted auth keys",
	)
	fs.StringVar(&c.Auth.RemoteFile, FlagAuthRemoteFile, c.Auth.RemoteFile,
		"Path to a file containing newline-separated scrypted auth keys for remote uploads",
	)
	fs.BoolVar(&c.NoDirectAgents, FlagNoDirectAgents, c.NoDirectAgents,
		"Disable serving files directly for wget/curl user agents",
	)
	fs.StringVar(&c.S3.Endpoint, FlagS3Endpoint, c.S3.Endpoint, "S3 endpoint")
	fs.StringVar(&c.S3.Region, FlagS3Region, c.S3.Region, "S3 region")
	fs.StringVar(&c.S3.Bucket, FlagS3Bucket, c.S3.Bucket, "S3 bucket to use for files and metadata")
	fs.BoolVar(&c.S3.ForcePathStyle, FlagS3ForcePathStyle, c.S3.ForcePathStyle,
		"Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)",
	)
	fs.BoolVar(&c.ForceRandomFilename, FlagForceRandomFilename, c.ForceRandomFilename,
		"Force all uploads to use a random filename",
	)
	fs.DurationVar(&c.Auth.CookieExpiry.Duration, FlagAuthCookieExpiry, c.Auth.CookieExpiry.Duration,
		"Expiration time for access key cookies in seconds (set 0 to use session cookies)",
	)
	fs.StringVar(&c.CustomPagesDir, FlagCustomPagesDir, c.CustomPagesDir,
		"Path to directory containing .md files to render as custom pages",
	)
	fs.DurationVar(&c.CleanupEvery.Duration, FlagCleanupEvery, c.CleanupEvery.Duration,
		"How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed.",
	)
}
