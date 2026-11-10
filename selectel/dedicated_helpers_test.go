package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePrivateSubnet(t *testing.T) {
	testCases := []struct {
		cidr          string
		expectedError bool
	}{
		// Valid private subnets.
		{"10.0.0.0/8", false},
		{"10.100.123.0/24", false},
		{"10.255.255.255/32", false},
		{"172.16.0.0/12", false},
		{"172.20.100.0/24", false},
		{"172.31.255.255/32", false},
		{"192.168.0.0/16", false},
		{"192.168.100.0/24", false},
		{"192.168.255.255/32", false},

		// Invalid private subnets (public addresses).
		{"8.8.8.8/32", true},
		{"1.1.1.1/32", true},
		{"11.0.0.0/8", true},
		{"172.32.0.0/12", true},
		{"192.169.0.0/16", true},
		{"172.15.0.0/12", true},

		// Invalid CIDR formats.
		{"invalid-cidr", true},
		{"10.0.0.0", true},
		{"10.0.0.0/", true},
		{"/8", true},
		{"", true},
	}

	for _, testCase := range testCases {
		err := validatePrivateSubnet(testCase.cidr)
		if testCase.expectedError {
			assert.Error(t, err, "expected error for CIDR: %s", testCase.cidr)
		} else {
			assert.NoError(t, err, "unexpected error for CIDR: %s", testCase.cidr)
		}
	}
}
