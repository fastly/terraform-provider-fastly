package fastly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPtr(t *testing.T) {
	v := ""
	assert.Equal(t, v, *strToPtr(v))
}

func TestIntPtr(t *testing.T) {
	v := 2
	assert.Equal(t, v, *intToPtr(v))
}

func TestBoolPtr(t *testing.T) {
	v := true
	assert.Equal(t, v, *boolToPtr(v))
}
