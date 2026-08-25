package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStringSlice(t *testing.T) {
	assert.Equal(t, []string{"one", "two"}, ToStringSlice([]any{"one", "two"}))
	assert.Equal(t, []string{"one"}, ToStringSlice([]any{"one", 42, nil}))
	assert.Empty(t, ToStringSlice(nil))
}
