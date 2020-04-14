package selectel

import (
	"testing"

	"github.com/selectel/mks-go/pkg/v1/node"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
	"github.com/stretchr/testify/assert"
)

func TestGetMKSClusterV1Endpoint(t *testing.T) {
	expectedEndpoints := map[string]string{
		ru1Region: ru1MKSClusterV1Endpoint,
		ru2Region: ru2MKSClusterV1Endpoint,
		ru3Region: ru3MKSClusterV1Endpoint,
		ru7Region: ru7MKSClusterV1Endpoint,
		ru8Region: ru8MKSClusterV1Endpoint,
	}

	for region, expected := range expectedEndpoints {
		actual := getMKSClusterV1Endpoint(region)
		assert.Equal(t, expected, actual)
	}
}

func TestExpandMKSClusterNodegroupsV1CreateOpts(t *testing.T) {
	opts := map[string]interface{}{
		"count":             1,
		"cpus":              1,
		"ram_mb":            1024,
		"volume_gb":         10,
		"volume_type":       "fast.ru-3a",
		"keypair_name":      "ssh-key",
		"availability_zone": "ru-3a",
	}

	expected := &nodegroup.CreateOpts{
		Count:            1,
		CPUs:             1,
		RAMMB:            1024,
		VolumeGB:         10,
		VolumeType:       "fast.ru-3a",
		KeypairName:      "ssh-key",
		AvailabilityZone: "ru-3a",
	}

	actual := expandMKSClusterNodegroupsV1CreateOpts(opts)
	assert.Equal(t, expected, actual)
}

func TestFlattenMKSClusterNodegroupsV1(t *testing.T) {
	r := resourceMKSClusterV1()
	d := r.TestResourceData()
	d.SetId("1")
	nodegroups := []map[string]interface{}{
		{
			"name":              "test-nodegroup-0",
			"id":                "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
			"availability_zone": "ru-3a",
			"volume_type":       "fast.ru-3a",
			"count":             1,
			"cpus":              1,
			"ram_mb":            1024,
			"volume_gb":         10,
		},
	}
	d.Set("nodegroups", nodegroups)

	views := []*nodegroup.View{
		{
			ID:               "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
			FlavorID:         "edc0b355-b540-495a-982f-efa28988ed5c",
			VolumeGB:         10,
			VolumeType:       "fast.ru-3a",
			LocalVolume:      false,
			AvailabilityZone: "ru-3a",
			Nodes: []*node.View{
				{
					ID:       "8d7bbe81-6521-4253-a9ba-f7e0bce7470c",
					Hostname: "test-cluster-0-node-xd9jk",
					IP:       "198.51.100.11",
				},
			},
		},
	}

	expected := []map[string]interface{}{
		{
			"name":              "test-nodegroup-0",
			"id":                "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
			"flavor_id":         "edc0b355-b540-495a-982f-efa28988ed5c",
			"count":             1,
			"cpus":              1,
			"ram_mb":            1024,
			"volume_gb":         10,
			"volume_type":       "fast.ru-3a",
			"local_volume":      false,
			"availability_zone": "ru-3a",
			"keypair_name":      "",
			"affinity_policy":   "",
			"nodes": []map[string]interface{}{
				{
					"id":       "8d7bbe81-6521-4253-a9ba-f7e0bce7470c",
					"hostname": "test-cluster-0-node-xd9jk",
					"ip":       "198.51.100.11",
				},
			},
		},
	}

	actual := flattenMKSClusterNodegroupsV1(d, views)
	assert.ElementsMatch(t, expected, actual)
}
