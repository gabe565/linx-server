package config

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/utils/bytefmt"
)

type Config struct {
	Bind        string `toml:"bind"`
	FilesPath   string `toml:"files-path"             comment:"Path to files directory"`
	MetaPath    string `toml:"meta-path"              comment:"Path to metadata directory"`
	SiteName    string `toml:"site-name"`
	SiteURL     URL    `toml:"site-url"`
	FrontendURL string `toml:"frontend-url,omitempty"`
	SelifPath   string `toml:"selif-path"             comment:"Path relative to site base url where files are accessed directly"`
	Fastcgi     bool   `toml:"fastcgi"                comment:"Serve through fastcgi"`

	MaxSize             Bytes    `toml:"max-size"              comment:"Maximum upload file size in bytes"`
	MaxExpiry           Duration `toml:"max-expiry"            comment:"Maximum expiration time (a value of 0s means no expiry)"`
	AllowHotlink        bool     `toml:"allow-hotlink"         comment:"Allow hot-linking of files"`
	RemoteUploads       bool     `toml:"remote-uploads"        comment:"Enable remote uploads (/upload?url=https://...)"`
	NoDirectAgents      bool     `toml:"no-direct-agents"      comment:"Disable serving files directly for wget/curl user agents"`
	ForceRandomFilename bool     `toml:"force-random-filename" comment:"Force all uploads to use a random filename"`
	NoLogs              bool     `toml:"no-logs"               comment:"Remove stdout output for each request"`
	NoTorrent           bool     `toml:"no-torrent"            comment:"Disable the torrent file endpoint"`

	CleanupEvery Duration `toml:"cleanup-every" comment:"How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed."`

	CustomPagesDir string `toml:"custom-pages-dir" comment:"Path to directory containing .md files to render as custom pages"`

	TLS    TLS    `toml:"tls"    comment:"TLS (HTTPS) configuration"`
	Auth   Auth   `toml:"auth"`
	S3     S3     `toml:"s3"     comment:"S3-compatible storage configuration"`
	Limit  Limit  `toml:"limit"  comment:"Configure rate limits"`
	Header Header `toml:"header" comment:"Modify request/response headers"`
}

type TLS struct {
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

type Auth struct {
	CookieExpiry Duration `toml:"cookie-expiry" comment:"Expiration time for access key cookies (set to 0s to use session cookies)"`
	Basic        bool     `toml:"basic"         comment:"Allow logging in with basic auth password"`
	File         string   `toml:"file"          comment:"Path to a file containing newline-separated scrypted auth keys"`
	RemoteFile   string   `toml:"remote-file"   comment:"Path to a file containing newline-separated scrypted auth keys for remote uploads"`
}

type S3 struct {
	Endpoint       string `toml:"endpoint"`
	Region         string `toml:"region"`
	Bucket         string `toml:"bucket"`
	ForcePathStyle bool   `toml:"force-path-style" comment:"Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)"`
}

type Limit struct {
	UploadMaxRequests int      `toml:"upload-max-requests"`
	UploadInterval    Duration `toml:"upload-interval"`

	FileMaxRequests int      `toml:"file-max-requests"`
	FileInterval    Duration `toml:"file-interval"`
}

type Header struct {
	RealIP                    bool              `toml:"real-ip"                      comment:"Use X-Real-IP/X-Forwarded-For headers"`
	AddHeaders                map[string]string `toml:"add-headers,inline"`
	ContentSecurityPolicy     string            `toml:"content-security-policy"`
	FileContentSecurityPolicy string            `toml:"file-content-security-policy"`
	ReferrerPolicy            string            `toml:"referrer-policy"`
	FileReferrerPolicy        string            `toml:"file-referrer-policy"`
	XFrameOptions             string            `toml:"x-frame-options"`
}

func New() *Config {
	c := &Config{
		Bind:         "127.0.0.1:8080",
		FilesPath:    "data/files",
		MetaPath:     "data/meta",
		SelifPath:    "selif",
		MaxSize:      4 * bytefmt.GiB,
		CleanupEvery: Duration{time.Hour},
		Limit: Limit{
			UploadMaxRequests: 5,
			UploadInterval:    Duration{15 * time.Second},
			FileMaxRequests:   20,
			FileInterval:      Duration{10 * time.Second},
		},
		Header: Header{
			AddHeaders:                map[string]string{},
			ContentSecurityPolicy:     "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
			FileContentSecurityPolicy: "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
			ReferrerPolicy:            "same-origin",
			FileReferrerPolicy:        "same-origin",
			XFrameOptions:             "SAMEORIGIN",
		},
	}
	if os.Getenv("LINX_DEFAULTS") == "container" {
		c.Bind = ":8080"
		c.FilesPath = "/data/files"
		c.MetaPath = "/data/meta"
		c.SiteName = "Linx"
	}
	return c
}

//nolint:gochecknoglobals
var (
	Default        = New()
	StorageBackend backends.StorageBackend
	TimeStarted    time.Time
	TimeStartedStr string
	RemoteAuthKeys []string
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
