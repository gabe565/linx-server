package keyhash

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	KeyPrefix        = "scrypt$"
	KeyPrefixURLSafe = "scrypt."
	scryptSalt       = "linx-server"
	scryptN          = 16384
	scryptR          = 8
	scryptP          = 1
	scryptKeyLen     = 32
)

func Hash(key, salt string, urlSafe bool) (string, error) {
	if salt == "" {
		salt = scryptSalt
	}
	hashed, err := scrypt.Key([]byte(key), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return "", err
	}

	prefix, encoding := getEncoding(urlSafe)
	return prefix + encoding.EncodeToString(hashed), nil
}

func IsValidHash(key string, urlSafe bool) bool {
	prefix, encoding := getEncoding(urlSafe)

	if len(key) <= len(prefix)+scryptKeyLen {
		return false
	}

	raw, found := strings.CutPrefix(key, prefix)
	if !found {
		return false
	}

	_, err := encoding.DecodeString(raw)
	return err == nil
}

func Check(stored, request, salt string, urlSafe bool) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}
	requestHash, err := scrypt.Key([]byte(request), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	prefix, encoding := getEncoding(urlSafe)
	raw := strings.TrimPrefix(stored, prefix)

	storedHash, err := encoding.DecodeString(raw)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(storedHash, requestHash) == 1, nil
}

func CheckList(stored []string, request, salt string, urlSafe bool) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}

	requestHash, err := scrypt.Key([]byte(request), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	prefix, encoding := getEncoding(urlSafe)

	for _, entry := range stored {
		raw := strings.TrimPrefix(entry, prefix)

		storedHash, err := encoding.DecodeString(raw)
		if err != nil {
			return false, err
		}

		if subtle.ConstantTimeCompare(storedHash, requestHash) == 1 {
			return true, nil
		}
	}

	return false, nil
}

func CheckWithFallback(stored, request, salt string) (bool, error) {
	switch {
	case IsValidHash(stored, true):
		return Check(stored, request, salt, true)
	case IsValidHash(stored, false):
		return Check(stored, request, salt, false)
	default:
		return stored == request, nil
	}
}

func getEncoding(urlSafe bool) (string, *base64.Encoding) {
	if urlSafe {
		return KeyPrefixURLSafe, base64.RawURLEncoding
	}
	return KeyPrefix, base64.StdEncoding
}
