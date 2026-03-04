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

func Hash(key, salt string, encoding *base64.Encoding) (string, error) {
	if salt == "" {
		salt = scryptSalt
	}
	hashed, err := scrypt.Key([]byte(key), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return "", err
	}

	return KeyPrefix + encoding.EncodeToString(hashed), nil
}

func IsValidHash(key string, encoding *base64.Encoding) bool {
	if len(key) <= len(KeyPrefix)+scryptKeyLen {
		return false
	}

	raw, found := strings.CutPrefix(key, KeyPrefix)
	if !found {
		return false
	}

	_, err := encoding.DecodeString(raw)
	return err == nil
}

func Check(stored, request, salt string, encoding *base64.Encoding) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}
	requestHash, err := scrypt.Key([]byte(request), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	raw := strings.TrimPrefix(stored, KeyPrefix)
	storedHash, err := encoding.DecodeString(raw)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(storedHash, requestHash) == 1, nil
}

func CheckList(stored []string, request, salt string, encoding *base64.Encoding) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}

	requestHash, err := scrypt.Key([]byte(request), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return false, err
	}

	for _, entry := range stored {
		raw := strings.TrimPrefix(entry, KeyPrefix)

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
	if strings.HasPrefix(stored, KeyPrefix) {
		encoding := determineEncoding(stored)
		return Check(stored, request, salt, encoding)
	}

	return stored == request, nil
}

func determineEncoding(key string) *base64.Encoding {
	if strings.ContainsAny(key, "+/=") {
		return base64.StdEncoding
	}
	return base64.RawURLEncoding
}
