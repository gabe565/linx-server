package template

import (
	"bytes"
	"encoding/json"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
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

func ConfigBytes() ([]byte, error) {
	f, err := assets.Static().Open(manifest["src/fouc.js"].File)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, err
	}

	buf.WriteString("window.config=")
	if err := json.NewEncoder(&buf).Encode(NewConfig()); err != nil {
		return nil, err
	}
	buf.WriteByte(';')
	return buf.Bytes(), nil
}
