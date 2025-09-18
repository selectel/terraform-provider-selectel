package reflect

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func jsonToMap(t *testing.T, s string) map[string]interface{} {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	require.NoError(t, err)
	return m
}

func TestIsSetContainsSubset(t *testing.T) {
	tests := []struct {
		name   string
		set    string
		subset string
		want   bool
	}{
		{
			name:   "SimpleMatch",
			set:    `{"a": 1, "b": 2}`,
			subset: `{"a": 1}`,
			want:   true,
		},
		{
			name:   "KeyNotFound",
			set:    `{"a": 1}`,
			subset: `{"b": 2}`,
			want:   false,
		},
		{
			name:   "NestedObjects",
			set:    `{"a": {"b": 2, "c": 3}}`,
			subset: `{"a": {"b": 2}}`,
			want:   true,
		},
		{
			name:   "Arrays",
			set:    `{"arr": [1, 2, 3]}`,
			subset: `{"arr": [2, 3]}`,
			want:   true,
		},
		{
			name:   "ArraysWithObjects",
			set:    `{"arr": [{"a": 1, "b": 2}, {"a": 2}]}`,
			subset: `{"arr": [{"a": 1}]}`,
			want:   true,
		},
		{
			name:   "ValueMismatch",
			set:    `{"a": 1}`,
			subset: `{"a": 2}`,
			want:   false,
		},
		{
			name:   "NestedArrays",
			set:    `{"a": [[1,2],[3,4]]}`,
			subset: `{"a": [[1]]}`,
			want:   true,
		},
		{
			name:   "EmptySubset",
			set:    `{"a": 1}`,
			subset: `{}`,
			want:   true,
		},
		{
			name:   "EmptySet",
			set:    `{}`,
			subset: `{"a": 1}`,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := jsonToMap(t, tt.set)
			subset := jsonToMap(t, tt.subset)
			got := IsSetContainsSubset(subset, set)
			if got != tt.want {
				t.Errorf("IsSetContainsSubset(%v, %v) = %v, want %v", tt.subset, tt.set, got, tt.want)
			}
		})
	}
}
