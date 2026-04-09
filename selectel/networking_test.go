package selectel

import (
	"testing"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/subnets"
	"github.com/stretchr/testify/assert"
)

func TestGetPrefixLengthFromCIDR(t *testing.T) {
	testingData := map[string]int{
		"192.0.2.100/29":   29,
		"192.0.2.200/28":   28,
		"203.0.113.10/24":  24,
		"203.0.113.129/25": 25,
	}

	for cidr, expected := range testingData {
		actual, err := getPrefixLengthFromCIDR(cidr)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}

func TestGetIPVersionFromPrefixLength(t *testing.T) {
	testingData := map[int]string{
		29: string(selvpcclient.IPv4),
		28: string(selvpcclient.IPv4),
		48: string(selvpcclient.IPv6),
		64: string(selvpcclient.IPv6),
		24: string(selvpcclient.IPv4),
		25: string(selvpcclient.IPv4),
	}

	for prefixLength, expected := range testingData {
		actual := getIPVersionFromPrefixLength(prefixLength)

		assert.Equal(t, expected, actual)
	}
}

func TestSubnetsMapsFromStructs(t *testing.T) {
	subnetsStructs := []subnets.Subnet{
		{
			NetworkID:     "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			SubnetID:      "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			Region:        "ru-2",
			CIDR:          "192.168.200.0/24",
			VLANID:        1003,
			ProjectID:     "b63ab68796e34858befb8fa2a8b1e12a",
			VTEPIPAddress: "10.10.0.101",
		},
		{
			NetworkID:     "954c6ebd-f923-4471-847a-e1be04af8952",
			SubnetID:      "4754c984-bb91-4221-820c-ae2b0f64dae0",
			Region:        "ru-3",
			CIDR:          "192.168.200.0/24",
			VLANID:        1003,
			ProjectID:     "b63ab68796e34858befb8fa2a8b1e12a",
			VTEPIPAddress: "10.10.0.201",
		},
	}
	expectedSubnetsMaps := []map[string]interface{}{
		{
			"network_id":      "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			"subnet_id":       "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			"region":          "ru-2",
			"cidr":            "192.168.200.0/24",
			"vlan_id":         1003,
			"project_id":      "b63ab68796e34858befb8fa2a8b1e12a",
			"vtep_ip_address": "10.10.0.101",
		},
		{
			"network_id":      "954c6ebd-f923-4471-847a-e1be04af8952",
			"subnet_id":       "4754c984-bb91-4221-820c-ae2b0f64dae0",
			"region":          "ru-3",
			"cidr":            "192.168.200.0/24",
			"vlan_id":         1003,
			"project_id":      "b63ab68796e34858befb8fa2a8b1e12a",
			"vtep_ip_address": "10.10.0.201",
		},
	}

	actualSubnetsMaps := subnetsMapsFromStructs(subnetsStructs)

	assert.ElementsMatch(t, expectedSubnetsMaps, actualSubnetsMaps)
}

func TestValidatePrivateSubnet(t *testing.T) {
	type testCase struct {
		cidr          string
		expectedError bool
	}
	testCases := []testCase{
		// Valid private subnets
		{"10.0.0.0/8", false},
		{"10.100.123.0/24", false},
		{"10.255.255.255/32", false},
		{"172.16.0.0/12", false},
		{"172.20.100.0/24", false},
		{"172.31.255.255/32", false},
		{"192.168.0.0/16", false},
		{"192.168.100.0/24", false},
		{"192.168.255.255/32", false},

		// Invalid private subnets (public addresses)
		{"8.8.8.8/32", true},
		{"1.1.1.1/32", true},
		{"11.0.0.0/8", true},
		{"172.32.0.0/12", true},
		{"192.169.0.0/16", true},
		{"172.15.0.0/12", true},

		// Invalid CIDR formats
		{"invalid-cidr", true},
		{"10.0.0.0", true},
		{"10.0.0.0/", true},
		{"/8", true},
		{"", true},
	}

	for _, tc := range testCases {
		err := validatePrivateSubnet(tc.cidr)
		if tc.expectedError {
			assert.Error(t, err, "Expected error for CIDR: %s", tc.cidr)
		} else {
			assert.NoError(t, err, "Unexpected error for CIDR: %s", tc.cidr)
		}
	}
}
