package reflect

import (
	"encoding/json"
	"reflect"
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

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		FieldA          string `json:"field_a"`
		FieldB          int    `json:"field_b"`
		FieldC          bool   `json:"field_c"`
		unexportedField string // This should be ignored by JSON marshaling
	}

	type NestedStruct struct {
		Name     string       `json:"name"`
		Value    int          `json:"value"`
		Children []TestStruct `json:"children"`
	}

	tests := []struct {
		name    string
		input   interface{}
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "SimpleStruct",
			input: TestStruct{
				FieldA:          "test_value",
				FieldB:          42,
				FieldC:          true,
				unexportedField: "should_be_ignored",
			},
			want: map[string]interface{}{
				"field_a": "test_value",
				"field_b": 42.0, // JSON unmarshals numbers as float64
				"field_c": true,
			},
			wantErr: false,
		},
		{
			name: "NestedStruct",
			input: NestedStruct{
				Name:  "parent",
				Value: 100,
				Children: []TestStruct{
					{
						FieldA: "child1",
						FieldB: 1,
						FieldC: true,
					},
					{
						FieldA: "child2",
						FieldB: 2,
						FieldC: false,
					},
				},
			},
			want: map[string]interface{}{
				"name":  "parent",
				"value": 100.0, // JSON unmarshals numbers as float64
				"children": []interface{}{
					map[string]interface{}{
						"field_a": "child1",
						"field_b": 1.0,
						"field_c": true,
					},
					map[string]interface{}{
						"field_a": "child2",
						"field_b": 2.0,
						"field_c": false,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyStruct",
			input:   struct{}{},
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "StructWithSlice",
			input: struct {
				Array []int  `json:"array"`
				Name  string `json:"name"`
			}{
				Array: []int{1, 2, 3},
				Name:  "test",
			},
			want: map[string]interface{}{
				"array": []interface{}{1.0, 2.0, 3.0}, // Numbers become float64 in JSON unmarshaling
				"name":  "test",
			},
			wantErr: false,
		},
		{
			name: "PointerToStruct",
			input: &TestStruct{
				FieldA: "pointer_value",
				FieldB: 999,
				FieldC: false,
			},
			want: map[string]interface{}{
				"field_a": "pointer_value",
				"field_b": 999.0,
				"field_c": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
