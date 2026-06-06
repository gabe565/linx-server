package csp_test

import (
	"testing"

	"gabe565.com/linx-server/internal/csp"
	"github.com/stretchr/testify/assert"
)

func TestCSPString(t *testing.T) {
	tests := []struct {
		name string
		in   csp.CSP
		want string
	}{
		{"nil", nil, ""},
		{"empty", csp.CSP{}, ""},
		{
			"single directive, no sources",
			csp.CSP{"upgrade-insecure-requests": nil},
			"upgrade-insecure-requests;",
		},
		{
			"single directive, one source",
			csp.CSP{"default-src": {csp.Self}},
			"default-src 'self';",
		},
		{
			"single directive, multiple sources",
			csp.CSP{"img-src": {csp.Self, csp.Data}},
			"img-src 'self' data:;",
		},
		{
			"multiple directives are sorted alphabetically",
			csp.CSP{
				"style-src":   {csp.Self, csp.UnsafeInline},
				"default-src": {csp.None},
				"img-src":     {csp.Self, csp.Data},
			},
			"default-src 'none'; img-src 'self' data:; style-src 'self' 'unsafe-inline';",
		},
		{
			"arbitrary URL source preserved verbatim",
			csp.CSP{"img-src": {csp.Self, "https://s3.example.com"}},
			"img-src 'self' https://s3.example.com;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.String())
		})
	}
}
