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
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	src, err := CheckAccessKey(req, &backends.Metadata{})
	require.NoError(t, err)
	assert.Equal(t, AccessKeySourceNone, src)
}

func TestCheckAccessKeyHeaderValid(t *testing.T) {
	stored, err := keyhash.Hash("supersecret")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AccessKeyHeader, "supersecret")

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored})
	require.NoError(t, err)
	assert.Equal(t, AccessKeySourceHeader, src)
}

func TestCheckAccessKeyCookieHasPriority(t *testing.T) {
	stored, err := keyhash.Hash("supersecret")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AccessKeyHeader, Value: url.PathEscape("wrong")})
	req.Header.Set(AccessKeyHeader, "supersecret")

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored})
	require.ErrorIs(t, err, errInvalidAccessKey)
	assert.Equal(t, AccessKeySourceCookie, src)
}

func TestCheckAccessKeyHeaderHasPriorityOverForm(t *testing.T) {
	stored, err := keyhash.Hash("supersecret")
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/?"+AccessKeyParam+"=supersecret",
		strings.NewReader("access_key=supersecret"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(AccessKeyHeader, "wrong")

	src, err := CheckAccessKey(req, &backends.Metadata{AccessKey: stored})
	require.ErrorIs(t, err, errInvalidAccessKey)
	assert.Equal(t, AccessKeySourceHeader, src)
}
