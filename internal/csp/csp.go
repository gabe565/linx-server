// Package csp builds Content-Security-Policy header values.
package csp

import (
	"maps"
	"slices"
	"strings"
)

// Common source keywords. The single quotes are part of the CSP spec and need
// to be in the emitted header, so we bake them into the constants.
const (
	Self         = "'self'"
	None         = "'none'"
	UnsafeInline = "'unsafe-inline'"
	Data         = "data:"
)

// CSP represents a Content-Security-Policy as directive name -> source list.
type CSP map[string][]string

func (c CSP) String() string {
	var b strings.Builder
	for i, k := range slices.Sorted(maps.Keys(c)) {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		for _, src := range c[k] {
			b.WriteByte(' ')
			b.WriteString(src)
		}
		b.WriteByte(';')
	}
	return b.String()
}
