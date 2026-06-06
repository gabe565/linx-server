package keyhash

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"iter"
	"slices"
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

func applyAppKey(key, appKey string) []byte {
	if appKey == "" {
		return []byte(key)
	}
	mac := hmac.New(sha256.New, []byte(appKey))
	mac.Write([]byte(key))
	return mac.Sum(nil)
}

func Hash(key, salt string, appKeys []string, urlSafe bool) (string, error) {
	if salt == "" {
		salt = scryptSalt
	}
	var appKey string
	if len(appKeys) > 0 {
		appKey = appKeys[0]
	}
	hashed, err := scrypt.Key(applyAppKey(key, appKey), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
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

func candidateKeys(appKeys []string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, k := range appKeys {
			if k != "" {
				if !yield(k) {
					return
				}
			}
		}
		yield("")
	}
}

func Check(stored, request, salt string, appKeys []string, urlSafe bool) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}

	prefix, encoding := getEncoding(urlSafe)
	raw := strings.TrimPrefix(stored, prefix)
	storedHash, err := encoding.DecodeString(raw)
	if err != nil {
		return false, err
	}

	for ak := range candidateKeys(appKeys) {
		requestHash, err := scrypt.Key(applyAppKey(request, ak), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
		if err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare(storedHash, requestHash) == 1 {
			return true, nil
		}
	}
	return false, nil
}

func CheckList(stored []string, request, salt string, appKeys []string, urlSafe bool) (bool, error) {
	if salt == "" {
		salt = scryptSalt
	}

	prefix, encoding := getEncoding(urlSafe)
	candidates := slices.Collect(candidateKeys(appKeys))

	// Memoize scrypt output per app-key candidate across all entries.
	hashes := make([][]byte, len(candidates))

	for _, entry := range stored {
		raw := strings.TrimPrefix(entry, prefix)
		storedHash, err := encoding.DecodeString(raw)
		if err != nil {
			return false, err
		}

		for i, ak := range candidates {
			if hashes[i] == nil {
				h, err := scrypt.Key(applyAppKey(request, ak), []byte(salt), scryptN, scryptR, scryptP, scryptKeyLen)
				if err != nil {
					return false, err
				}
				hashes[i] = h
			}
			if subtle.ConstantTimeCompare(storedHash, hashes[i]) == 1 {
				return true, nil
			}
		}
	}
	return false, nil
}

func CheckWithFallback(stored, request, salt string, appKeys []string) (bool, error) {
	switch {
	case IsValidHash(stored, true):
		return Check(stored, request, salt, appKeys, true)
	case IsValidHash(stored, false):
		return Check(stored, request, salt, appKeys, false)
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
