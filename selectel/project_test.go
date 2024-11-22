package selectel

import (
	"testing"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
)

func TestFlattenVPCProjectV2Theme(t *testing.T) {
	testCases := []struct {
		have projects.Theme
		want map[string]string
	}{
		{
			have: projects.Theme{},
			want: nil,
		},
		{
			have: projects.Theme{
				Color: "some_color",
			},
			want: map[string]string{
				"color": "some_color",
			},
		},
		{
			have: projects.Theme{
				Logo: "some_logo",
			},
			want: map[string]string{
				"logo": "some_logo",
			},
		},
		{
			have: projects.Theme{
				Color: "another_color",
				Logo:  "another_logo",
			},
			want: map[string]string{
				"color": "another_color",
				"logo":  "another_logo",
			},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.want, flattenVPCProjectV2Theme(testCase.have))
	}
}
