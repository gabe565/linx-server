package util

import (
	"strconv"
	"strings"
)

func ParseBool(s string, defaultVal bool) bool {
	s = strings.ToLower(s)
	switch s {
	case "yes", "y":
		return true
	case "no", "n":
		return false
	default:
		if v, err := strconv.ParseBool(s); err == nil {
			return v
		}
		return defaultVal
	}
}
