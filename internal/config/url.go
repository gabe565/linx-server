package config

import (
	"net/url"
)

type URL struct {
	url.URL
}

func (u URL) String() string {
	return u.URL.String()
}

func (u *URL) Set(s string) error {
	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}
	u.URL = *parsed
	return nil
}

func (u *URL) Type() string {
	return typeString
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URL) UnmarshalText(text []byte) error {
	return u.Set(string(text))
}
