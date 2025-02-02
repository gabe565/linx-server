package config

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	FlagConfig                    = "config"
	FlagBind                      = "bind"
	FlagFiles                     = "files-dir"
	FlagMeta                      = "meta-dir"
	FlagNoLogs                    = "no-logs"
	FlagBasicAuth                 = "basic-auth"
	FlagAllowHotlink              = "allow-hotlink"
	FlagSiteName                  = "site-name"
	FlagSiteURL                   = "site-url"
	FlagSelifPath                 = "selif-path"
	FlagMaxSize                   = "max-size"
	FlagMaxExpiry                 = "max-expiry"
	FlagTLSCert                   = "tls-cert"
	FlagTLSKey                    = "tls-key"
	FlagRealIP                    = "real-ip"
	FlagFastcgi                   = "fastcgi"
	FlagRemoteUploads             = "remote-uploads"
	FlagAuthFile                  = "auth-file"
	FlagRemoteAuthFile            = "remote-auth-file"
	FlagContentSecurityPolicy     = "content-security-policy"
	FlagFileContentSecurityPolicy = "file-content-security-policy"
	FlagReferrerPolicy            = "referrer-policy"
	FlagFileReferrerPolicy        = "file-referrer-policy"
	FlagXFrameOptions             = "x-frame-options"
	FlagAddHeader                 = "add-header"
	FlagNoDirectAgents            = "no-direct-agents"
	FlagS3Endpoint                = "s3-endpoint"
	FlagS3Region                  = "s3-region"
	FlagS3Bucket                  = "s3-bucket"
	FlagS3ForcePathStyle          = "s3-force-path-style"
	FlagForceRandomFilename       = "force-random-filename"
	FlagAccessKeyCookieExpiry     = "access-key-cookie-expiry"
	FlagCustomPagesDir            = "custom-pages-path"
	FlagCleanupEvery              = "cleanup-every"
)

func (c *Config) RegisterBasicFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	confPath, _ := getDefaultFile()
	if home, err := os.UserHomeDir(); err == nil {
		confPath = strings.Replace(confPath, home, "$HOME", 1)
	}
	fs.StringP(FlagConfig, "c", confPath, "Path to the config file")
	fs.StringVar(&c.FilesDir, FlagFiles, c.FilesDir, "Path to files directory")
	fs.StringVar(&c.MetaDir, FlagMeta, c.MetaDir, "Path to metadata directory")
	fs.BoolVar(&c.NoLogs, FlagNoLogs, c.NoLogs, "Remove logging of each request")
}

func (c *Config) RegisterServeFlags(cmd *cobra.Command) {
	c.RegisterBasicFlags(cmd)

	fs := cmd.Flags()
	fs.StringVar(&c.Bind, FlagBind, c.Bind, "Host to bind to")
	fs.BoolVar(&c.BasicAuth, FlagBasicAuth, c.BasicAuth, "Allow logging in with basic auth password")
	fs.BoolVar(&c.AllowHotlink, FlagAllowHotlink, c.AllowHotlink, "Allow hot-linking of files")
	fs.StringVar(&c.SiteName, FlagSiteName, c.SiteName, "Name of the site")
	fs.Var(&c.SiteURL, FlagSiteURL, "Site base url")
	fs.StringVar(&c.SelifPath, FlagSelifPath, c.SelifPath, "Path relative to site base url where files are accessed directly")
	fs.Var(&c.MaxSize, FlagMaxSize, "Maximum upload file size in bytes")
	fs.DurationVar(&c.MaxExpiry.Duration, FlagMaxExpiry, c.MaxExpiry.Duration, "Maximum expiration time. A value of 0 means no expiry.")
	fs.StringVar(&c.TLSCert, FlagTLSCert, c.TLSCert, "Path to ssl certificate (for https)")
	fs.StringVar(&c.TLSKey, FlagTLSKey, c.TLSKey, "Path to ssl key (for https)")
	fs.BoolVar(&c.RealIP, FlagRealIP, c.RealIP, "Use X-Real-IP/X-Forwarded-For headers")
	fs.BoolVar(&c.Fastcgi, FlagFastcgi, c.Fastcgi, "Serve through fastcgi")
	fs.BoolVar(&c.RemoteUploads, FlagRemoteUploads, c.RemoteUploads, "Enable remote uploads")
	fs.StringVar(&c.AuthFile, FlagAuthFile, c.AuthFile, "Path to a file containing newline-separated scrypted auth keys")
	fs.StringVar(&c.RemoteAuthFile, FlagRemoteAuthFile, c.RemoteAuthFile, "Path to a file containing newline-separated scrypted auth keys for remote uploads")
	fs.StringVar(&c.ContentSecurityPolicy, FlagContentSecurityPolicy, c.ContentSecurityPolicy, "Value of default Content-Security-Policy header")
	fs.StringVar(&c.FileContentSecurityPolicy, FlagFileContentSecurityPolicy, c.FileContentSecurityPolicy, "Value of Content-Security-Policy header for file access")
	fs.StringVar(&c.ReferrerPolicy, FlagReferrerPolicy, c.ReferrerPolicy, "Value of default Referrer-Policy header")
	fs.StringVar(&c.FileReferrerPolicy, FlagFileReferrerPolicy, c.FileReferrerPolicy, "Value of Referrer-Policy header for file access")
	fs.StringVar(&c.XFrameOptions, FlagXFrameOptions, c.XFrameOptions, "Value of X-Frame-Options header")
	fs.Var(&c.AddHeaders, FlagAddHeader, "Add an arbitrary header to the response. This option can be used multiple times.")
	fs.BoolVar(&c.NoDirectAgents, FlagNoDirectAgents, c.NoDirectAgents, "Disable serving files directly for wget/curl user agents")
	fs.StringVar(&c.S3Endpoint, FlagS3Endpoint, c.S3Endpoint, "S3 endpoint")
	fs.StringVar(&c.S3Region, FlagS3Region, c.S3Region, "S3 region")
	fs.StringVar(&c.S3Bucket, FlagS3Bucket, c.S3Bucket, "S3 bucket to use for files and metadata")
	fs.BoolVar(&c.S3ForcePathStyle, FlagS3ForcePathStyle, c.S3ForcePathStyle, "Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)")
	fs.BoolVar(&c.ForceRandomFilename, FlagForceRandomFilename, c.ForceRandomFilename, "Force all uploads to use a random filename")
	fs.Uint64Var(&c.AccessKeyCookieExpiry, FlagAccessKeyCookieExpiry, c.AccessKeyCookieExpiry, "Expiration time for access key cookies in seconds (set 0 to use session cookies)")
	fs.StringVar(&c.CustomPagesDir, FlagCustomPagesDir, c.CustomPagesDir, "Path to directory containing .md files to render as custom pages")
	fs.DurationVar(&c.CleanupEvery.Duration, FlagCleanupEvery, c.CleanupEvery.Duration, "How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed.")
}
