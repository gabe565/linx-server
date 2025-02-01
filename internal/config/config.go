package config

import (
	"html/template"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
)

type HeaderList []string

func (h *HeaderList) String() string {
	return strings.Join(*h, ",")
}

func (h *HeaderList) Set(value string) error {
	*h = append(*h, value)
	return nil
}

type Config struct {
	Bind                      string
	FilesDir                  string
	MetaDir                   string
	SiteName                  string
	SiteURL                   string
	SitePath                  string
	SelifPath                 string
	CertFile                  string
	KeyFile                   string
	ContentSecurityPolicy     string
	FileContentSecurityPolicy string
	ReferrerPolicy            string
	FileReferrerPolicy        string
	XFrameOptions             string
	MaxSize                   int64
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

var (
	Default            Config
	StorageBackend     backends.StorageBackend
	Templates          map[string]*template.Template
	TimeStarted        time.Time
	TimeStartedStr     string
	RemoteAuthKeys     []string
	MetaStorageBackend backends.MetaStorageBackend
)
