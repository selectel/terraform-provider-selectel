package selvpc

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/crossregionsubnets"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
	"github.com/stretchr/testify/assert"
)

func TestExpandResellV2Regions(t *testing.T) {
	r := resourceResellKeypairV2()
	d := r.TestResourceData()
	d.SetId("1")
	regions := []interface{}{"ru-1", "ru-2", "ru-3"}
	d.Set("regions", regions)

	expected := []string{"ru-1", "ru-2", "ru-3"}

	actual := expandResellV2Regions(d.Get("regions").(*schema.Set))

	assert.ElementsMatch(t, expected, actual)
}

func TestExpandResellV2CrossRegionOpts(t *testing.T) {
	r := resourceResellCrossRegionSubnetV2()
	d := r.TestResourceData()
	d.SetId("1")
	regions := []interface{}{
		map[string]interface{}{
			"region": "ru-1",
		},
		map[string]interface{}{
			"region": "ru-3",
		},
	}
	d.Set("regions", regions)

	expected := []crossregionsubnets.CrossRegionOpt{
		{
			Region: "ru-1",
		},
		{
			Region: "ru-3",
		},
	}

	actual, err := expandResellV2CrossRegionOpts(d.Get("regions").(*schema.Set))

	assert.Empty(t, err)
	assert.ElementsMatch(t, expected, actual)
}

func TestRegionsMapsFromSubnetsStructs(t *testing.T) {
	subnetsStructs := []subnets.Subnet{
		{
			NetworkID:     "912bd5d0-cb11-4a7f-af7c-ea84c8e7db2e",
			SubnetID:      "4912cca9-cad2-49c1-a69a-929cd4cf9559",
			Region:        "ru-1",
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
	expected := []map[string]interface{}{
		{
			"region": "ru-1",
		},
		{
			"region": "ru-3",
		},
	}

	actual := regionsMapsFromSubnetsStructs(subnetsStructs)

	assert.ElementsMatch(t, expected, actual)
}
