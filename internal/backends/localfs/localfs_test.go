package localfs

import (
	"strings"
	"testing"

	"gabe565.com/linx-server/internal/backends"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackend_Claim(t *testing.T) {
	b := New(t.TempDir(), t.TempDir())
	_, err := b.Put(t.Context(), strings.NewReader("content"), "test.txt", 0, backends.PutOptions{})
	require.NoError(t, err)

	claimed, err := b.Claim(t.Context(), "test.txt")
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = b.Claim(t.Context(), "test.txt")
	require.NoError(t, err)
	assert.False(t, claimed)

	require.NoError(t, b.Delete(t.Context(), "test.txt"))

	claimed, err = b.Claim(t.Context(), "test.txt")
	require.NoError(t, err)
	assert.False(t, claimed, "a deleted file should not be claimable")

	_, err = b.Put(t.Context(), strings.NewReader("content"), "test.txt", 0, backends.PutOptions{})
	require.NoError(t, err)

	claimed, err = b.Claim(t.Context(), "test.txt")
	require.NoError(t, err)
	assert.True(t, claimed, "claims should be removed with the file")
}
