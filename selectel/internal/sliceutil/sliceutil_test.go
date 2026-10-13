package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnion(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, Union([]string{"a", "b"}, []string{"b", "c"}))
	assert.Equal(t, []string{"a"}, Union([]string{"a"}, nil))
	assert.Equal(t, []string{"a"}, Union(nil, []string{"a", "a"}))
	assert.Empty(t, Union[string](nil, nil))
}

func TestSubtract(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, Subtract([]string{"a", "b", "c"}, []string{"b"}))
	assert.Equal(t, []string{"a"}, Subtract([]string{"a"}, nil))
	assert.Empty(t, Subtract([]string{"a"}, []string{"a"}))
}

func TestIntersect(t *testing.T) {
	assert.Equal(t, []string{"b", "c"}, Intersect([]string{"a", "b", "c"}, []string{"c", "b"}))
	assert.Empty(t, Intersect([]string{"a"}, nil))
	assert.Empty(t, Intersect(nil, []string{"a"}))
}
