package expiry

import (
	"time"

	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/dustin/go-humanize"
)

var defaultExpiryList = []uint64{
	60,
	300,
	3600,
	86400,
	604800,
	2419200,
	31536000,
}

type ExpirationTime struct {
	Seconds uint64
	Human   string
}

// Determine if the given filename is expired
func IsFileExpired(filename string) (bool, error) {
	metadata, err := config.StorageBackend.Head(filename)
	if err != nil {
		return false, err
	}

	return IsTsExpired(metadata.Expiry), nil
}

// Return a list of expiration times and their humanized versions
func ListExpirationTimes() []ExpirationTime {
	epoch := time.Now()
	actualExpiryInList := false
	var expiryList []ExpirationTime

	for _, expiryEntry := range defaultExpiryList {
		if config.Default.MaxExpiry == 0 || expiryEntry <= config.Default.MaxExpiry {
			if expiryEntry == config.Default.MaxExpiry {
				actualExpiryInList = true
			}

			duration := time.Duration(expiryEntry) * time.Second
			expiryList = append(expiryList, ExpirationTime{
				Seconds: expiryEntry,
				Human:   humanize.RelTime(epoch, epoch.Add(duration), "", ""),
			})
		}
	}

	if config.Default.MaxExpiry == 0 {
		expiryList = append(expiryList, ExpirationTime{
			0,
			"never",
		})
	} else if actualExpiryInList == false {
		duration := time.Duration(config.Default.MaxExpiry) * time.Second
		expiryList = append(expiryList, ExpirationTime{
			config.Default.MaxExpiry,
			humanize.RelTime(epoch, epoch.Add(duration), "", ""),
		})
	}

	return expiryList
}

var NeverExpire = time.Unix(0, 0)

// Determine if a file with expiry set to "ts" has expired yet
func IsTsExpired(ts time.Time) bool {
	now := time.Now()
	return ts != NeverExpire && now.After(ts)
}
