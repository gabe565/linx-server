package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
	"github.com/minio/sha256-simd"
)

type Config struct {
	SiteName        string           `json:"site_name"`
	MaxSize         int64            `json:"max_size"`
	ForceRandom     bool             `json:"force_random"`
	Auth            bool             `json:"auth"`
	ExpirationTimes []ExpirationTime `json:"expiration_times"`
	CustomPages     []string         `json:"custom_pages,omitzero"`
}

type ExpirationTime struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewConfig() Config {
	expirationTimes := expiry.ListExpirationTimes()
	conf := Config{
		SiteName:        config.Default.SiteName,
		ForceRandom:     config.Default.ForceRandomFilename,
		MaxSize:         int64(config.Default.MaxSize),
		Auth:            config.Default.Auth.Basic || config.Default.Auth.File != "",
		ExpirationTimes: make([]ExpirationTime, 0, len(expirationTimes)),
		CustomPages:     config.CustomPages,
	}
	for _, t := range expirationTimes {
		conf.ExpirationTimes = append(conf.ExpirationTimes, ExpirationTime{
			Name:  t.Human,
			Value: t.Duration.String(),
		})
	}
	return conf
}

func ConfigString() string {
	conf, _ := json.Marshal(NewConfig())
	return "window.config=" + string(bytes.TrimSpace(conf))
}

func ConfigHash() string {
	if config.ComputedHash == "" {
		hash := sha256.Sum256([]byte(ConfigString()))
		config.ComputedHash = "'sha256-" + base64.StdEncoding.EncodeToString(hash[:]) + "'"
	}
	return config.ComputedHash
}
