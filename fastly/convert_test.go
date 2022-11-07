package fastly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultIntToZero(t *testing.T) {
	assert.Equal(t, int(0), intOrDefault(nil))
}

func TestDefaultInt(t *testing.T) {
	v := int(10)
	assert.Equal(t, v, intOrDefault(&v))
}
