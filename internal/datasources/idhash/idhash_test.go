package idhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashIDs(t *testing.T) {
	assert.Equal(t, HashIDs([]string{"a", "b"}), HashIDs([]string{"b", "a"}), "order should not affect the hash")
	assert.NotEqual(t, HashIDs([]string{"a", "b"}), HashIDs([]string{"a", "c"}), "different sets should hash differently")
}
