package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBool(t *testing.T) {
	type args struct {
		s          string
		defaultVal bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"no", args{"no", true}, false},
		{"n", args{"n", true}, false},
		{"yes", args{"yes", false}, true},
		{"y", args{"y", false}, true},
		{"parse bool", args{"true", false}, true},
		{"invalid default false", args{"abc", false}, false},
		{"invalid default true", args{"abc", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseBool(tt.args.s, tt.args.defaultVal))
		})
	}
}
