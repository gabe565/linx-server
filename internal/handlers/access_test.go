package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gabe565.com/linx-server/internal/auth/keyhash"
	"gabe565.com/linx-server/internal/backends"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAccessKeyNoProtection(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	src, err := CheckAccessKey(req, &backends.Metadata{})
	require.NoError(t, err)
	assert.Equal(t, AccessKeySourceNone, src)
}

func TestCheckAccessKeyHeaderValid(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	stored, err := keyhash.Hash(key, salt, true)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set(AccessKeyHeader, key)

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored, Salt: salt})
	require.NoError(t, err)
	assert.Equal(t, AccessKeySourceHeader, src)
}

func TestCheckAccessKeyCookieHasPriority(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	stored, err := keyhash.Hash(key, salt, true)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AccessKeyHeader, Value: url.PathEscape("wrong")})
	req.Header.Set(AccessKeyHeader, key)

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored, Salt: salt})
	require.ErrorIs(t, err, errInvalidAccessKey)
	assert.Equal(t, AccessKeySourceCookie, src)
}

func TestCheckAccessKeyHeaderHasPriorityOverForm(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	stored, err := keyhash.Hash(key, salt, true)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(),
		http.MethodPost,
		"/?"+AccessKeyParam+"="+key,
		strings.NewReader("access_key="+key),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(AccessKeyHeader, "wrong")

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored, Salt: salt})
	require.ErrorIs(t, err, errInvalidAccessKey)
	assert.Equal(t, AccessKeySourceHeader, src)
}

func TestCheckAccessKeyStdBase64Fallback(t *testing.T) {
	const key, salt = "supersecret", "mysalt"

	stored, err := keyhash.Hash(key, salt, false)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AccessKeyHeader, Value: url.PathEscape("wrong")})
	req.Header.Set(AccessKeyHeader, key)

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored, Salt: salt})
	require.ErrorIs(t, err, errInvalidAccessKey)
	assert.Equal(t, AccessKeySourceCookie, src)
}
