package keyhash

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckList(t *testing.T) {
	stored := []string{
		KeyPrefix + "vhvZ/PT1jeTbTAJ8JdoxddqFtebSxdVb0vwPlYO+4HM=",
		KeyPrefix + "vFpNprT9wbHgwAubpvRxYCCpA2FQMAK6hFqPvAGrdZo=",
	}

	ok, err := CheckList(stored, "", "", base64.StdEncoding)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "thisisnotvalid", "", base64.StdEncoding)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "haPVipRnGJ0QovA9nyqK", "", base64.StdEncoding)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestHashAndCheckRoundTrip(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	hash, err := Hash(key, salt, base64.StdEncoding)
	require.NoError(t, err)
	assert.True(t, IsValidHash(hash, base64.StdEncoding))

	urlHash, err := Hash(key, salt, base64.RawURLEncoding)
	require.NoError(t, err)
	assert.True(t, IsValidHash(urlHash, base64.RawURLEncoding))

	assert.NotEqual(t, hash, urlHash)

	ok, err := Check(hash, key, salt, base64.StdEncoding)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Check(urlHash, key, salt, base64.RawURLEncoding)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Check(hash, "wrong", salt, base64.StdEncoding)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = Check(hash, key, "wrongsalt", base64.StdEncoding)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckWithFallback(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	hash, err := Hash(key, salt, base64.StdEncoding)
	require.NoError(t, err)

	urlHash, err := Hash(key, salt, base64.RawURLEncoding)
	require.NoError(t, err)

	ok, err := CheckWithFallback(hash, key, salt)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckWithFallback(urlHash, key, salt)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckWithFallback("plaintext", "wrong", "")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckListInvalidHash(t *testing.T) {
	_, err := CheckList([]string{KeyPrefix + "not-base64!"}, "anything", "", base64.StdEncoding)
	require.Error(t, err)
}
