package util

import (
	"testing"

	"gabe565.com/linx-server/internal/backends"
	"github.com/stretchr/testify/assert"
)

func TestInferLang(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		meta     backends.Metadata
		want     string
	}{
		{
			name:     "prefers generated filename extension",
			fileName: "file.ts",
			meta: backends.Metadata{
				OriginalName: "file.py",
				Mimetype:     "text/plain",
			},
			want: "typescript",
		},
		{
			name:     "falls back to original filename extension",
			fileName: "file.bin",
			meta: backends.Metadata{
				OriginalName: "file.rb",
				Mimetype:     "application/octet-stream",
			},
			want: "ruby",
		},
		{
			name:     "dockerfile filename fallback",
			fileName: "random.bin",
			meta: backends.Metadata{
				OriginalName: "Dockerfile.dev",
				Mimetype:     "application/octet-stream",
			},
			want: "dockerfile",
		},
		{
			name:     "makefile filename fallback",
			fileName: "random.bin",
			meta: backends.Metadata{
				OriginalName: "Makefile",
				Mimetype:     "application/octet-stream",
			},
			want: "makefile",
		},
		{
			name:     "text mimetype fallback",
			fileName: "unknown.custom",
			meta: backends.Metadata{
				OriginalName: "unknown.custom",
				Mimetype:     "text/plain",
			},
			want: "text",
		},
		{
			name:     "unknown non text returns empty language",
			fileName: "unknown.custom",
			meta: backends.Metadata{
				OriginalName: "unknown.custom",
				Mimetype:     "application/octet-stream",
			},
			want: "",
		},
		{
			name:     "exact extension mapping can differ by case",
			fileName: "source.C",
			meta: backends.Metadata{
				Mimetype: "text/plain",
			},
			want: "cpp",
		},
		{
			name:     "lowercase fallback extension mapping",
			fileName: "source.JS",
			meta: backends.Metadata{
				Mimetype: "text/plain",
			},
			want: "javascript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferLang(tt.fileName, tt.meta)
			assert.Equal(t, tt.want, got)
		})
	}
}
