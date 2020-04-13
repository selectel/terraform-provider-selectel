package selectel

import (
	"testing"
	"time"

	"github.com/selectel/mks-go/pkg/v1/node"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterV1Endpoint(t *testing.T) {
	expectedEndpoints := map[string]string{
		ru1Region: ru1Endpoint,
		ru2Region: ru2Endpoint,
		ru3Region: ru3Endpoint,
		ru7Region: ru7Endpoint,
		ru8Region: ru8Endpoint,
	}

	for region, expected := range expectedEndpoints {
		actual := getClusterV1Endpoint(region)
		assert.Equal(t, expected, actual)
	}
}

func TestExpandNodegroupCreateOpts(t *testing.T) {
	opts := map[string]interface{}{
		"count":             1,
		"flavor_id":         "edc0b355-b540-495a-982f-efa28988ed5c",
		"cpus":              1,
		"ram_mb":            1024,
		"volume_gb":         10,
		"volume_type":       "fast.ru-3a",
		"local_volume":      false,
		"keypair_name":      "ssh-key",
		"availability_zone": "ru-3a",
	}

	expected := &nodegroup.CreateOpts{
		Count:            1,
		FlavorID:         "edc0b355-b540-495a-982f-efa28988ed5c",
		CPUs:             1,
		RAMMB:            1024,
		VolumeGB:         10,
		VolumeType:       "fast.ru-3a",
		LocalVolume:      false,
		KeypairName:      "ssh-key",
		AvailabilityZone: "ru-3a",
	}

	actual := expandNodegroupCreateOpts(opts)
	assert.Equal(t, expected, actual)
}

func TestFlattenNodegroups(t *testing.T) {
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

	ts, _ := time.Parse(time.RFC3339, "0001-01-01 00:00:00 +0000 UTC")
	views := []*nodegroup.View{
		{
			ID:               "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
			CreatedAt:        &ts,
			UpdatedAt:        &ts,
			ClusterID:        "9c4c1f09-1c4c-44cc-a848-a9a0ba648ffa",
			FlavorID:         "edc0b355-b540-495a-982f-efa28988ed5c",
			VolumeGB:         10,
			VolumeType:       "fast.ru-3a",
			LocalVolume:      false,
			AvailabilityZone: "ru-3a",
			Nodes: []*node.View{
				{
					ID:          "8d7bbe81-6521-4253-a9ba-f7e0bce7470c",
					CreatedAt:   &ts,
					UpdatedAt:   &ts,
					Hostname:    "test-cluster-0-node-xd9jk",
					IP:          "198.51.100.11",
					NodegroupID: "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
				},
			},
		},
	}

	expected := []map[string]interface{}{
		{
			"name":              "test-nodegroup-0",
			"id":                "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
			"created_at":        "0001-01-01 00:00:00 +0000 UTC",
			"updated_at":        "0001-01-01 00:00:00 +0000 UTC",
			"cluster_id":        "9c4c1f09-1c4c-44cc-a848-a9a0ba648ffa",
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
					"id":           "8d7bbe81-6521-4253-a9ba-f7e0bce7470c",
					"created_at":   "0001-01-01 00:00:00 +0000 UTC",
					"updated_at":   "0001-01-01 00:00:00 +0000 UTC",
					"hostname":     "test-cluster-0-node-xd9jk",
					"ip":           "198.51.100.11",
					"nodegroup_id": "be49545a-3a6d-447c-8e90-fd40ee1c3a3a",
				},
			},
		},
	}

	actual := flattenNodegroups(d, views)
	assert.ElementsMatch(t, expected, actual)
}
