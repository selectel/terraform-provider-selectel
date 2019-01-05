package selvpc

import (
	"testing"

	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
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
			NetworkID: "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			SubnetID:  "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			Region:    "ru-2",
		},
		{
			NetworkID: "954c6ebd-f923-4471-847a-e1be04af8952",
			SubnetID:  "4754c984-bb91-4221-820c-ae2b0f64dae0",
			Region:    "ru-3",
		},
	}
	expectedSubnetsMaps := []map[string]interface{}{
		{
			"network_id": "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			"subnet_id":  "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			"region":     "ru-2",
		},
		{
			"network_id": "954c6ebd-f923-4471-847a-e1be04af8952",
			"subnet_id":  "4754c984-bb91-4221-820c-ae2b0f64dae0",
			"region":     "ru-3",
		},
	}

	actualSubnetsMaps := subnetsMapsFromStructs(subnetsStructs)

	assert.ElementsMatch(t, expectedSubnetsMaps, actualSubnetsMaps)
}
