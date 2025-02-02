package expiry

import (
	"context"
	"time"

	"gabe565.com/linx-server/internal/config"
	"github.com/dustin/go-humanize"
)

//nolint:gochecknoglobals
var defaultExpiryList = []time.Duration{
	time.Minute,
	5 * time.Minute,
	time.Hour,
	24 * time.Hour,
	7 * 24 * time.Hour,
	31 * 24 * time.Hour,
	365 * 24 * time.Hour,
}

type ExpirationTime struct {
	Seconds uint64
	Human   string
}

// Determine if the given filename is expired
func IsFileExpired(ctx context.Context, filename string) (bool, error) {
	metadata, err := config.StorageBackend.Head(ctx, filename)
	if err != nil {
		return false, err
	}

	return IsTSExpired(metadata.Expiry), nil
}

// Return a list of expiration times and their humanized versions
func ListExpirationTimes() []ExpirationTime {
	epoch := time.Now()
	actualExpiryInList := false
	var expiryList []ExpirationTime

	for _, expiryEntry := range defaultExpiryList {
		if config.Default.MaxExpiry.Duration == 0 || expiryEntry <= config.Default.MaxExpiry.Duration {
			if expiryEntry == config.Default.MaxExpiry.Duration {
				actualExpiryInList = true
			}

			expiryList = append(expiryList, ExpirationTime{
				Seconds: uint64(expiryEntry.Seconds()),
				Human:   humanize.RelTime(epoch, epoch.Add(expiryEntry), "", ""),
			})
		}
	}

	if config.Default.MaxExpiry.Duration == 0 {
		expiryList = append(expiryList, ExpirationTime{
			0,
			"never",
		})
	} else if !actualExpiryInList {
		expiryList = append(expiryList, ExpirationTime{
			Seconds: uint64(config.Default.MaxExpiry.Seconds()),
			Human:   humanize.RelTime(epoch, epoch.Add(config.Default.MaxExpiry.Duration), "", ""),
		})
	}

	return expiryList
}

//nolint:gochecknoglobals
var NeverExpire = time.Unix(0, 0)

// Determine if a file with expiry set to "ts" has expired yet
func IsTSExpired(ts time.Time) bool {
	now := time.Now()
	return ts != NeverExpire && now.After(ts)
}
