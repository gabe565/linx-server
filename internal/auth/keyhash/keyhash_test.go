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

	ok, err := CheckList(stored, "")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "thisisnotvalid")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = CheckList(stored, "haPVipRnGJ0QovA9nyqK")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestHashAndCheckRoundTrip(t *testing.T) {
	hash, err := Hash("supersecret")
	require.NoError(t, err)
	assert.True(t, IsValidHash(hash))

	ok, err := Check(hash, "supersecret")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Check(hash, "wrong")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckWithFallback(t *testing.T) {
	ok, err := CheckWithFallback("plaintext", "plaintext")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckWithFallback("plaintext", "wrong")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckListInvalidHash(t *testing.T) {
	_, err := CheckList([]string{KeyPrefix + "not-base64!"}, "anything")
	require.Error(t, err)
}
