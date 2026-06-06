package keyhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckList(t *testing.T) {
	stored := []string{
		KeyPrefix + "vhvZ/PT1jeTbTAJ8JdoxddqFtebSxdVb0vwPlYO+4HM=",
		KeyPrefix + "vFpNprT9wbHgwAubpvRxYCCpA2FQMAK6hFqPvAGrdZo=",
	}

	ok, err := CheckList(stored, "", "", nil, false)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "thisisnotvalid", "", nil, false)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "haPVipRnGJ0QovA9nyqK", "", nil, false)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestHashAndCheckRoundTrip(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	hash, err := Hash(key, salt, nil, false)
	require.NoError(t, err)
	assert.True(t, IsValidHash(hash, false))

	urlHash, err := Hash(key, salt, nil, true)
	require.NoError(t, err)
	assert.True(t, IsValidHash(urlHash, true))

	assert.NotEqual(t, hash, urlHash)

	ok, err := Check(hash, key, salt, nil, false)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Check(urlHash, key, salt, nil, true)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Check(hash, "wrong", salt, nil, false)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = Check(hash, key, "wrongsalt", nil, false)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHashAndCheckRoundTripWithAppKey(t *testing.T) {
	const key, salt = "supersecret", "mysalt"
	appKeys := []string{"server-secret"}

	hash, err := Hash(key, salt, appKeys, false)
	require.NoError(t, err)

	ok, err := Check(hash, key, salt, appKeys, false)
	require.NoError(t, err)
	assert.True(t, ok)

	// Wrong app key fails (and the no-pepper fallback also doesn't match).
	ok, err = Check(hash, key, salt, []string{"different-secret"}, false)
	require.NoError(t, err)
	assert.False(t, ok)

	// No app keys configured: only the no-pepper attempt runs, which fails
	// since the stored hash was peppered.
	ok, err = Check(hash, key, salt, nil, false)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckLegacyHashWithAppKeyConfigured(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	// Legacy hash created without an app key.
	legacyHash, err := Hash(key, salt, nil, false)
	require.NoError(t, err)

	// Verification with app keys configured still succeeds via the
	// no-pepper fallback.
	ok, err := Check(legacyHash, key, salt, []string{"server-secret"}, false)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRemoveAppKey(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	// File hashed with a pepper.
	oldHash, err := Hash(key, salt, []string{"retired-key"}, false)
	require.NoError(t, err)

	// Admin unsets app-key but keeps the old in app-previous-keys.
	retired := []string{"", "retired-key"}

	// New hashes are unpeppered.
	newHash, err := Hash(key, salt, retired, false)
	require.NoError(t, err)
	ok, err := Check(newHash, key, salt, nil, false)
	require.NoError(t, err)
	assert.True(t, ok, "newly-created hash should be unpeppered")

	// Old peppered hashes still verify via the previous-key fallback.
	ok, err = Check(oldHash, key, salt, retired, false)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckAppKeyRotation(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	// Hash created with the old app key.
	oldHash, err := Hash(key, salt, []string{"old-key"}, false)
	require.NoError(t, err)

	// After rotation, "new-key" is primary, "old-key" kept as previous.
	rotated := []string{"new-key", "old-key"}

	ok, err := Check(oldHash, key, salt, rotated, false)
	require.NoError(t, err)
	assert.True(t, ok, "old hash should still verify against the previous key")

	// New hashes use the new primary.
	newHash, err := Hash(key, salt, rotated, false)
	require.NoError(t, err)

	ok, err = Check(newHash, key, salt, rotated, false)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckWithFallback(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	hash, err := Hash(key, salt, nil, false)
	require.NoError(t, err)

	urlHash, err := Hash(key, salt, nil, true)
	require.NoError(t, err)

	ok, err := CheckWithFallback(hash, key, salt, nil)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckWithFallback(urlHash, key, salt, nil)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckWithFallback("plaintext", "wrong", "", nil)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckListInvalidHash(t *testing.T) {
	_, err := CheckList([]string{KeyPrefix + "not-base64!"}, "anything", "", nil, false)
	require.Error(t, err)
}
