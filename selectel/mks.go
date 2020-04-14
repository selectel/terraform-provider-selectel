package selectel

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/cluster"
	"github.com/selectel/mks-go/pkg/v1/node"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
)

const (
	ru1Region = "ru-1"
	ru2Region = "ru-2"
	ru3Region = "ru-3"
	ru7Region = "ru-7"
	ru8Region = "ru-8"

	ru1MKSClusterV1Endpoint = "https://ru-1.mks.selcloud.ru/v1"
	ru2MKSClusterV1Endpoint = "https://ru-2.mks.selcloud.ru/v1"
	ru3MKSClusterV1Endpoint = "https://ru-3.mks.selcloud.ru/v1"
	ru7MKSClusterV1Endpoint = "https://ru-7.mks.selcloud.ru/v1"
	ru8MKSClusterV1Endpoint = "https://ru-8.mks.selcloud.ru/v1"
)

func getMKSClusterV1Endpoint(region string) (endpoint string) {
	switch region {
	case ru1Region:
		endpoint = ru1MKSClusterV1Endpoint
	case ru2Region:
		endpoint = ru2MKSClusterV1Endpoint
	case ru3Region:
		endpoint = ru3MKSClusterV1Endpoint
	case ru7Region:
		endpoint = ru7MKSClusterV1Endpoint
	case ru8Region:
		endpoint = ru8MKSClusterV1Endpoint
	}

	return
}

func hashMKSClusterNodegroupsV1(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["name"].(string)))
}

func expandMKSClusterNodegroupsV1CreateOpts(v interface{}) *nodegroup.CreateOpts {
	m := v.(map[string]interface{})
	opts := nodegroup.CreateOpts{}
	if _, ok := m["count"]; ok {
		opts.Count = m["count"].(int)
	}
	if _, ok := m["flavor_id"]; ok {
		opts.FlavorID = m["flavor_id"].(string)
	}
	if _, ok := m["cpus"]; ok {
		opts.CPUs = m["cpus"].(int)
	}
	if _, ok := m["ram_mb"]; ok {
		opts.RAMMB = m["ram_mb"].(int)
	}
	if _, ok := m["volume_gb"]; ok {
		opts.VolumeGB = m["volume_gb"].(int)
	}
	if _, ok := m["volume_type"]; ok {
		opts.VolumeType = m["volume_type"].(string)
	}
	if _, ok := m["keypair_name"]; ok {
		opts.KeypairName = m["keypair_name"].(string)
	}
	if _, ok := m["affinity_policy"]; ok {
		opts.AffinityPolicy = m["affinity_policy"].(string)
	}
	if _, ok := m["availability_zone"]; ok {
		opts.AvailabilityZone = m["availability_zone"].(string)
	}

	return &opts
}

func waitForMKSClusterV1ActiveState(
	ctx context.Context, client *v1.ServiceClient, clusterID string, timeout time.Duration) error {
	pending := []string{
		string(cluster.StatusPendingCreate),
		string(cluster.StatusPendingUpdate),
		string(cluster.StatusPendingResize),
	}
	target := []string{
		string(cluster.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    mksClusterV1StateRefreshFunc(ctx, client, clusterID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the cluster %s to become 'ACTIVE': %s",
			clusterID, err)
	}

	return nil
}

func mksClusterV1StateRefreshFunc(
	ctx context.Context, client *v1.ServiceClient, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, _, err := cluster.Get(ctx, client, clusterID)
		if err != nil {
			return nil, "", err
		}

		return c, string(c.Status), nil
	}
}

func flattenMKSClusterNodegroupsV1(d *schema.ResourceData, views []*nodegroup.View) []interface{} {
	nodegroups := d.Get("nodegroups").(*schema.Set).List()
	for i, v := range nodegroups {
		m := v.(map[string]interface{})
		if _, ok := m["id"]; ok {
			for _, view := range views {
				if m["id"] == view.ID {
					m["cluster_id"] = view.ClusterID
					m["flavor_id"] = view.FlavorID
					m["volume_gb"] = view.VolumeGB
					m["volume_type"] = view.VolumeType
					m["local_volume"] = view.LocalVolume
					m["availability_zone"] = view.AvailabilityZone
					m["nodes"] = flattenMKSClusterNodesV1(view.Nodes)

					if view.CreatedAt != nil {
						m["created_at"] = view.CreatedAt.String()
					}
					if view.UpdatedAt != nil {
						m["updated_at"] = view.UpdatedAt.String()
					}
				}
				nodegroups[i] = m
			}
		}
	}

	return nodegroups
}

func flattenMKSClusterNodesV1(views []*node.View) []map[string]interface{} {
	nodes := make([]map[string]interface{}, len(views))
	for i, view := range views {
		nodes[i] = make(map[string]interface{})
		nodes[i]["id"] = view.ID
		nodes[i]["hostname"] = view.Hostname
		nodes[i]["ip"] = view.IP
		nodes[i]["nodegroup_id"] = view.NodegroupID

		if view.CreatedAt != nil {
			nodes[i]["created_at"] = view.CreatedAt.String()
		}
		if view.UpdatedAt != nil {
			nodes[i]["updated_at"] = view.UpdatedAt.String()
		}
	}

	return nodes
}
