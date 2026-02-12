package keyhash

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	KeyPrefix    = "scrypt$"
	scryptSalt   = "linx-server"
	scryptN      = 16384
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
)

func Hash(key string) (string, error) {
	hashed, err := scrypt.Key([]byte(key), []byte(scryptSalt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return "", err
	}

	return KeyPrefix + base64.StdEncoding.EncodeToString(hashed), nil
}

func IsValidHash(key string) bool {
	if len(key) <= len(KeyPrefix)+scryptKeyLen {
		return false
	}

	raw, found := strings.CutPrefix(key, KeyPrefix)
	if !found {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(raw)
	return err == nil
}

func Check(stored, request string) (bool, error) {
	requestHash, err := scrypt.Key([]byte(request), []byte(scryptSalt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	raw := strings.TrimPrefix(stored, KeyPrefix)
	storedHash, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(storedHash, requestHash) == 1, nil
}

func CheckList(stored []string, request string) (bool, error) {
	requestHash, err := scrypt.Key([]byte(request), []byte(scryptSalt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	for _, entry := range stored {
		raw := strings.TrimPrefix(entry, KeyPrefix)

		storedHash, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return false, err
		}

		if subtle.ConstantTimeCompare(storedHash, requestHash) == 1 {
			return true, nil
		}
	}

	return false, nil
}

func CheckWithFallback(stored, request string) (bool, error) {
	if strings.HasPrefix(stored, KeyPrefix) {
		return Check(stored, request)
	}

	return stored == request, nil
}
