package acls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashIDs(t *testing.T) {
	assert.Equal(t, hashIDs([]string{"a", "b"}), hashIDs([]string{"b", "a"}), "order should not affect the hash")
	assert.NotEqual(t, hashIDs([]string{"a", "b"}), hashIDs([]string{"a", "c"}), "different sets should hash differently")
}
