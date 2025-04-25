package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
)

type ConfigResponse struct {
	SiteName        string           `json:"site_name"`
	MaxSize         int64            `json:"max_size"`
	ForceRandom     bool             `json:"force_random"`
	Auth            bool             `json:"auth"`
	ExpirationTimes []ExpirationTime `json:"expiration_times"`
}

type ExpirationTime struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func Config(w http.ResponseWriter, r *http.Request) {
	expirationTimes := expiry.ListExpirationTimes()
	conf := ConfigResponse{
		SiteName:        config.Default.SiteName,
		ForceRandom:     config.Default.ForceRandomFilename,
		MaxSize:         int64(config.Default.MaxSize),
		Auth:            config.Default.Auth.Basic || config.Default.Auth.File != "",
		ExpirationTimes: make([]ExpirationTime, 0, len(expirationTimes)),
	}
	for _, t := range expirationTimes {
		conf.ExpirationTimes = append(conf.ExpirationTimes, ExpirationTime{
			Name:  t.Human,
			Value: t.Duration.String(),
		})
	}

	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(conf)
	http.ServeContent(w, r, "", config.TimeStarted, bytes.NewReader(buf.Bytes()))
}
