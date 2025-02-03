package config

import (
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/utils/bytefmt"
)

type HeaderList []string

func (h *HeaderList) String() string {
	return strings.Join(*h, ",")
}

func (h *HeaderList) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func (h *HeaderList) Type() string {
	return typeString
}

type Config struct {
	Bind      string `toml:"bind"`
	FilesPath string `toml:"files-path" comment:"Path to files directory"`
	MetaPath  string `toml:"meta-path" comment:"Path to metadata directory"`
	SiteName  string `toml:"site-name"`
	SiteURL   URL    `toml:"site-url"`
	SelifPath string `toml:"selif-path" comment:"Path relative to site base url where files are accessed directly"`
	Fastcgi   bool   `toml:"fastcgi" comment:"Serve through fastcgi"`

	MaxSize               Bytes    `toml:"max-size" comment:"Maximum upload file size in bytes"`
	MaxExpiry             Duration `toml:"max-expiry" comment:"Maximum expiration time. A value of 0 means no expiry."`
	AllowHotlink          bool     `toml:"allow-hotlink" comment:"Allow hot-linking of files"`
	RemoteUploads         bool     `toml:"remote-uploads" comment:"Enable remote uploads (/upload?url=https://...)"`
	NoDirectAgents        bool     `toml:"no-direct-agents" comment:"Disable serving files directly for wget/curl user agents"`
	ForceRandomFilename   bool     `toml:"force-random-filename" comment:"Force all uploads to use a random filename"`
	AccessKeyCookieExpiry uint64   `toml:"access-key-cookie-expiry" comment:"Expiration time for access key cookies in seconds (set 0 to use session cookies)"`
	NoLogs                bool     `toml:"no-logs" comment:"Remove stdout output for each request"`

	BasicAuth      bool   `toml:"basic-auth" comment:"Allow logging in with basic auth password"`
	AuthFile       string `toml:"auth-file" comment:"Path to a file containing newline-separated scrypted auth keys"`
	RemoteAuthFile string `toml:"remote-auth-file" comment:"Path to a file containing newline-separated scrypted auth keys for remote uploads"`

	CleanupEvery Duration `toml:"cleanup-every" comment:"How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed."`

	TLSCert string `toml:"tls-cert" comment:"HTTPS configuration"`
	TLSKey  string `toml:"tls-key"`

	S3Endpoint       string `toml:"s3-endpoint" comment:"AWS S3 configuration"`
	S3Region         string `toml:"s3-region"`
	S3Bucket         string `toml:"s3-bucket"`
	S3ForcePathStyle bool   `toml:"s3-force-path-style" comment:"Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)"`

	CustomPagesDir string `toml:"custom-pages-dir" comment:"Path to directory containing .md files to render as custom pages"`

	RealIP                    bool       `toml:"real-ip" comment:"Use X-Real-IP/X-Forwarded-For headers"`
	AddHeaders                HeaderList `toml:"add-headers" comment:"Add arbitrary headers to the response"`
	ContentSecurityPolicy     string     `toml:"content-security-policy" comment:"Value of default Content-Security-Policy header"`
	FileContentSecurityPolicy string     `toml:"file-content-security-policy" comment:"Value of Content-Security-Policy header for file access"`
	ReferrerPolicy            string     `toml:"referrer-policy" comment:"Value of default Referrer-Policy header"`
	FileReferrerPolicy        string     `toml:"file-referrer-policy" comment:"Value of Referrer-Policy header for file access"`
	XFrameOptions             string     `toml:"x-frame-options" comment:"Value of X-Frame-Options header"`
}

func New() *Config {
	c := &Config{
		Bind:                      "127.0.0.1:8080",
		FilesPath:                 "data/files",
		MetaPath:                  "data/meta",
		SelifPath:                 "selif",
		MaxSize:                   4 * bytefmt.GiB,
		CleanupEvery:              Duration{time.Hour},
		ContentSecurityPolicy:     "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		FileContentSecurityPolicy: "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		ReferrerPolicy:            "same-origin",
		FileReferrerPolicy:        "same-origin",
		XFrameOptions:             "SAMEORIGIN",
	}
	if os.Getenv("LINX_DEFAULTS") == "container" {
		c.Bind = ":8080"
		c.FilesPath = "/data/files"
		c.MetaPath = "/data/meta"
		c.SiteName = "linx"
	}
	return c
}

//nolint:gochecknoglobals
var (
	Default            = New()
	StorageBackend     backends.StorageBackend
	Templates          map[string]*template.Template
	TimeStarted        time.Time
	TimeStartedStr     string
	RemoteAuthKeys     []string
	MetaStorageBackend backends.MetaStorageBackend
)

func getDefaultFile() (string, error) {
	const configDir, configFile = "linx-server", "config.toml"
	var dir string
	switch runtime.GOOS {
	case "darwin":
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			dir = filepath.Join(xdgConfigHome, configDir)
			break
		}
		fallthrough
	default:
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(dir, configDir)
	}
	return filepath.Join(dir, configFile), nil
}

func (c *Config) MaxExpirySeconds() uint64 {
	return uint64(c.MaxExpiry.Seconds())
}

const typeString = "string"
