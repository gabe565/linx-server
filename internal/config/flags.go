package config

import (
	"gabe565.com/utils/must"
	"github.com/spf13/cobra"
)

const (
	FlagBind                = "bind"
	FlagFiles               = "files"
	FlagMeta                = "meta"
	FlagNoLogs              = "no-logs"
	FlagBasicAuth           = "basic-auth"
	FlagAllowHotlink        = "allow-hotlink"
	FlagSiteName            = "site-name"
	FlagSiteURL             = "site-url"
	FlagSelifPath           = "selif-path"
	FlagMaxSize             = "max-size"
	FlagMaxExpiry           = "max-expiry"
	FlagTLSCert             = "tls-cert"
	FlagTLSKey              = "tls-key"
	FlagRealIp              = "real-ip"
	FlagFastcgi             = "fastcgi"
	FlagRemoteUploads       = "remote-uploads"
	FlagAuthFile            = "auth-file"
	FlagRemoteAuthFile      = "remote-auth-file"
	FlagCSP                 = "csp"
	FlagFileCSP             = "file-csp"
	FlagReferrerPolicy      = "referrer-policy"
	FlagFileReferrerPolicy  = "file-referrer-policy"
	FlagXFrameOptions       = "x-frame-options"
	FlagAddHeader           = "add-header"
	FlagNoDirectAgents      = "no-direct-agents"
	FlagS3Endpoint          = "s3-endpoint"
	FlagS3Region            = "s3-region"
	FlagS3Bucket            = "s3-bucket"
	FlagS3ForcePathStyle    = "s3-force-path-style"
	FlagForceRandomFilename = "force-random-filename"
	FlagAccessCookieExpiry  = "access-cookie-expiry"
	FlagCustomPagesDir      = "custom-pages-path"
	FlagCleanupEveryMinutes = "cleanup-every-minutes"
)

func (c *Config) RegisterBasicFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.StringVar(&c.FilesDir, FlagFiles, c.FilesDir, "Path to files directory")
	fs.StringVar(&c.MetaDir, FlagMeta, c.MetaDir, "Path to metadata directory")
	fs.BoolVar(&c.NoLogs, FlagNoLogs, c.NoLogs, "Remove stdout output for each request")

	// Deprecated
	fs.StringVar(&c.FilesDir, "filespath", c.FilesDir, "")
	must.Must(fs.MarkDeprecated("filespath", "use --files instead"))
	fs.StringVar(&c.MetaDir, "metapath", c.MetaDir, "")
	must.Must(fs.MarkDeprecated("metapath", "use --meta instead"))
	fs.BoolVar(&c.NoLogs, "nologs", c.NoLogs, "")
	must.Must(fs.MarkDeprecated("nologs", "use --quiet instead"))
}

func (c *Config) RegisterServeFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.StringVar(&c.Bind, FlagBind, c.Bind, "Host to bind to")
	fs.BoolVar(&c.BasicAuth, FlagBasicAuth, c.BasicAuth, "Allow logging by basic auth password")
	fs.BoolVar(&c.AllowHotlink, FlagAllowHotlink, c.AllowHotlink, "Allow hot-linking of files")
	fs.StringVar(&c.SiteName, FlagSiteName, c.SiteName, "Name of the site")
	fs.StringVar(&c.SiteURL, FlagSiteURL, c.SiteURL, "Site base url")
	fs.StringVar(&c.SelifPath, FlagSelifPath, c.SelifPath, "Path relative to site base url where files are accessed directly")
	fs.Var(&c.MaxSize, FlagMaxSize, "Maximum upload file size in bytes (default 4GB)")
	fs.Uint64Var(&c.MaxExpiry, FlagMaxExpiry, c.MaxExpiry, "Maximum expiration time in seconds (default is 0, which is no expiry)")
	fs.StringVar(&c.TLSCert, FlagTLSCert, c.TLSCert, "Path to ssl certificate (for https)")
	fs.StringVar(&c.TLSKey, FlagTLSKey, c.TLSKey, "Path to ssl key (for https)")
	fs.BoolVar(&c.RealIp, FlagRealIp, c.RealIp, "Use X-Real-IP/X-Forwarded-For headers as original host")
	fs.BoolVar(&c.Fastcgi, FlagFastcgi, c.Fastcgi, "Serve through fastcgi")
	fs.BoolVar(&c.RemoteUploads, FlagRemoteUploads, c.RemoteUploads, "Enable remote uploads")
	fs.StringVar(&c.AuthFile, FlagAuthFile, c.AuthFile, "Path to a file containing newline-separated scrypted auth keys")
	fs.StringVar(&c.RemoteAuthFile, FlagRemoteAuthFile, c.RemoteAuthFile, "Path to a file containing newline-separated scrypted auth keys for remote uploads")
	fs.StringVar(&c.ContentSecurityPolicy, FlagCSP, c.ContentSecurityPolicy, "Value of default Content-Security-Policy header")
	fs.StringVar(&c.FileContentSecurityPolicy, FlagFileCSP, c.FileContentSecurityPolicy, "Value of Content-Security-Policy header for file access")
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
	fs.Uint64Var(&c.AccessKeyCookieExpiry, FlagAccessCookieExpiry, c.AccessKeyCookieExpiry, "Expiration time for access key cookies in seconds (set 0 to use session cookies)")
	fs.StringVar(&c.CustomPagesDir, FlagCustomPagesDir, c.CustomPagesDir, "Path to directory containing .md files to render as custom pages")
	fs.Uint64Var(&c.CleanupEveryMinutes, FlagCleanupEveryMinutes, c.CleanupEveryMinutes, "How often to clean up expired files in minutes (default is 0, which means files will be cleaned up as they are accessed)")

	// Deprecated
	fs.BoolVar(&c.BasicAuth, "basicauth", c.BasicAuth, "")
	must.Must(fs.MarkDeprecated("basicauth", "use --basic-auth instead"))
	fs.BoolVar(&c.AllowHotlink, "allowhotlink", c.AllowHotlink, "")
	must.Must(fs.MarkDeprecated("allowhotlink", "use --allow-hotlink instead"))
	fs.StringVar(&c.SiteName, "sitename", c.SiteName, "")
	must.Must(fs.MarkDeprecated("sitename", "use --site-name instead"))
	fs.StringVar(&c.SiteURL, "siteurl", c.SiteURL, "")
	must.Must(fs.MarkDeprecated("siteurl", "use --site-url instead"))
	fs.StringVar(&c.SelifPath, "selifpath", c.SelifPath, "")
	must.Must(fs.MarkDeprecated("selifpath", "use --selif-path instead"))
	fs.Var(&c.MaxSize, "maxsize", "")
	must.Must(fs.MarkDeprecated("maxsize", "use --max-size instead"))
	fs.Uint64Var(&c.MaxExpiry, "maxexpiry", c.MaxExpiry, "")
	must.Must(fs.MarkDeprecated("maxexpiry", "use --max-expiry instead"))
	fs.StringVar(&c.TLSCert, "certfile", c.TLSCert, "")
	must.Must(fs.MarkDeprecated("certfile", "use --tls-cert instead"))
	fs.StringVar(&c.TLSKey, "keyfile", c.TLSKey, "")
	must.Must(fs.MarkDeprecated("keyfile", "use --tls-key instead"))
	fs.BoolVar(&c.RealIp, "realip", c.RealIp, "")
	must.Must(fs.MarkDeprecated("realip", "use --real-ip instead"))
	fs.BoolVar(&c.RemoteUploads, "remoteuploads", c.RemoteUploads, "")
	must.Must(fs.MarkDeprecated("remoteuploads", "use --remote-uploads instead"))
	fs.StringVar(&c.AuthFile, "authfile", c.AuthFile, "")
	must.Must(fs.MarkDeprecated("authfile", "use --auth-file instead"))
	fs.StringVar(&c.RemoteAuthFile, "remoteauthfile", c.RemoteAuthFile, "")
	must.Must(fs.MarkDeprecated("remoteauthfile", "use --remote-auth-file instead"))
	fs.StringVar(&c.ContentSecurityPolicy, "contentsecuritypolicy", c.ContentSecurityPolicy, "")
	must.Must(fs.MarkDeprecated("contentsecuritypolicy", "use --csp instead"))
	fs.StringVar(&c.FileContentSecurityPolicy, "filecontentsecuritypolicy", c.FileContentSecurityPolicy, "")
	must.Must(fs.MarkDeprecated("filecontentsecuritypolicy", "use --file-csp instead"))
	fs.StringVar(&c.ReferrerPolicy, "referrerpolicy", c.ReferrerPolicy, "")
	must.Must(fs.MarkDeprecated("referrerpolicy", "use --referrer-policy instead"))
	fs.StringVar(&c.FileReferrerPolicy, "filereferrerpolicy", c.FileReferrerPolicy, "")
	must.Must(fs.MarkDeprecated("filereferrerpolicy", "use --file-referrer-policy instead"))
	fs.StringVar(&c.XFrameOptions, "xframeoptions", c.XFrameOptions, "")
	must.Must(fs.MarkDeprecated("xframeoptions", "use --x-frame-options instead"))
	fs.Var(&c.AddHeaders, "addheader", "")
	must.Must(fs.MarkDeprecated("addheader", "use --add-header instead"))
	fs.BoolVar(&c.NoDirectAgents, "nodirectagents", c.NoDirectAgents, "")
	must.Must(fs.MarkDeprecated("nodirectagents", "use --no-direct-agents instead"))
	fs.StringVar(&c.CustomPagesDir, "custompagespath", c.CustomPagesDir, "")
	must.Must(fs.MarkDeprecated("custompagespath", "use --custom-pages-path instead"))
}
