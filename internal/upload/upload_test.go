package upload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBarePlusExt(t *testing.T) {
	barename, extension := BarePlusExt("test.jpg.gz")
	assert.Equal(t, "test", barename)
	assert.Equal(t, "jpg.gz", extension)

	barename, extension = BarePlusExt("test.gz")
	assert.Equal(t, "test", barename)
	assert.Equal(t, "gz", extension)

	barename, extension = BarePlusExt("test.tar.gz")
	assert.Equal(t, "test", barename)
	assert.Equal(t, "tar.gz", extension)
}
