package util

import (
	"net/url"
)

func EncodeContentDisposition(mediatype, filename string) string {
	return mediatype + "; filename*=UTF-8''" + url.PathEscape(filename)
}
