package fastly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultUintToZero(t *testing.T) {
	assert.Equal(t, uint(0), uintOrDefault(nil))
}

func TestDefaultUint(t *testing.T) {
	v := uint(10)
	assert.Equal(t, v, uintOrDefault(&v))
}
