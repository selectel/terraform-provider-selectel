package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkingFloatingIPV2DataSourceSelectors(t *testing.T) {
	s := dataSourceVPCFloatingIPV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "floatingip_id", "address", "description",
		"pool", "port_id", "fixed_ip", "status", "tags",
	}, selectors)

	for _, name := range []string{
		"blocked", "qos_policy_id", "router_id", "floating_network_id", "all_tags",
	} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed)
		assert.False(t, s[name].Optional, name+" must be returned, not selectable")
	}
}

func TestNetworkingFloatingIPV2PoolID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := networkingFloatingIPV2PoolID(nil, "public")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match any external network")
	})

	t.Run("unique", func(t *testing.T) {
		id, err := networkingFloatingIPV2PoolID([]network.Network{{ID: "network-id"}}, "public")
		require.NoError(t, err)
		assert.Equal(t, "network-id", id)
	})

	t.Run("ambiguous", func(t *testing.T) {
		_, err := networkingFloatingIPV2PoolID(
			[]network.Network{{ID: "one"}, {ID: "two"}}, "public",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "matches 2 external networks")
	})
}

func TestNetworkingFloatingIPV2DataSourceResult(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := networkingFloatingIPV2DataSourceResult(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no networking floating ip found")
	})

	t.Run("unique", func(t *testing.T) {
		result, err := networkingFloatingIPV2DataSourceResult(
			[]floatingip.FloatingIP{{ID: "floating-ip-id"}},
		)
		require.NoError(t, err)
		assert.Equal(t, "floating-ip-id", result.ID)
	})

	t.Run("ambiguous", func(t *testing.T) {
		_, err := networkingFloatingIPV2DataSourceResult(
			[]floatingip.FloatingIP{{ID: "one"}, {ID: "two"}},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "2 networking floating ips found")
	})
}

func TestNetworkingFloatingIPV2DataSourceFlattenPreservesTagSelector(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceVPCFloatingIPV2().Schema, map[string]interface{}{
		"project_id": "project-id",
		"region":     "ru-3",
		"tags":       []interface{}{"selector"},
	})

	flattenNetworkingFloatingIPV2(d, &floatingip.FloatingIP{
		FloatingNetworkID: "network-id",
		FloatingIPAddress: "203.0.113.10",
		Tags:              []string{"selector", "external"},
	})

	assert.Equal(t, "203.0.113.10", d.Get("address"))
	assert.Equal(t, "network-id", d.Get("floating_network_id"))
	assert.ElementsMatch(t, []interface{}{"selector"}, d.Get("tags").(*schema.Set).List())
	assert.ElementsMatch(t, []interface{}{"selector", "external"}, d.Get("all_tags").(*schema.Set).List())
}
