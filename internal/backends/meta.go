package backends

import (
	"errors"
	"strconv"
	"time"
)

type Metadata struct {
	OriginalName string
	DeleteKey    string
	AccessKey    string
	Salt         string
	Checksum     string
	Mimetype     string
	Size         int64
	ModTime      time.Time
	Expiry       time.Time
	ArchiveFiles []string
}

var ErrBadMetadata = errors.New("corrupted metadata")

func (m Metadata) Etag() string {
	return strconv.Quote(
		m.Checksum + "-" + strconv.FormatInt(m.ModTime.Unix(), 36),
	)
}

func (m Metadata) Expired() bool {
	return !m.Expiry.IsZero() && m.Expiry.Before(time.Now())
}
