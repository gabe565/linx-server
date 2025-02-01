package config

import (
	"html/template"
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
	return "string"
}

type Config struct {
	Bind                      string
	FilesDir                  string
	MetaDir                   string
	SiteName                  string
	SiteURL                   string
	SitePath                  string
	SelifPath                 string
	TLSCert                   string
	TLSKey                    string
	ContentSecurityPolicy     string
	FileContentSecurityPolicy string
	ReferrerPolicy            string
	FileReferrerPolicy        string
	XFrameOptions             string
	MaxSize                   Bytes
	MaxExpiry                 uint64
	RealIp                    bool
	NoLogs                    bool
	AllowHotlink              bool
	Fastcgi                   bool
	RemoteUploads             bool
	BasicAuth                 bool
	AuthFile                  string
	RemoteAuthFile            string
	AddHeaders                HeaderList
	NoDirectAgents            bool
	S3Endpoint                string
	S3Region                  string
	S3Bucket                  string
	S3ForcePathStyle          bool
	ForceRandomFilename       bool
	AccessKeyCookieExpiry     uint64
	CustomPagesDir            string
	CleanupEveryMinutes       uint64
}

func New() *Config {
	return &Config{
		Bind:                      "127.0.0.1:8080",
		FilesDir:                  "files",
		MetaDir:                   "meta",
		SelifPath:                 "selif",
		ContentSecurityPolicy:     "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		FileContentSecurityPolicy: "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		ReferrerPolicy:            "same-origin",
		FileReferrerPolicy:        "same-origin",
		XFrameOptions:             "SAMEORIGIN",
		MaxSize:                   4 * bytefmt.GiB,
	}
}

var (
	Default            = New()
	StorageBackend     backends.StorageBackend
	Templates          map[string]*template.Template
	TimeStarted        time.Time
	TimeStartedStr     string
	RemoteAuthKeys     []string
	MetaStorageBackend backends.MetaStorageBackend
)
