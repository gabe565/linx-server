package backends

import (
	"errors"
	"time"
)

type Metadata struct {
	OriginalName string
	DeleteKey    string
	AccessKey    string
	Sha256sum    string
	Mimetype     string
	Size         int64
	ModTime      time.Time
	Expiry       time.Time
	ArchiveFiles []string
}

var ErrBadMetadata = errors.New("corrupted metadata")
