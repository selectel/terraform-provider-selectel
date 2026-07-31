package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingPortV2SchemaOwnership covers REQ-048 (SCN-048-01): the
// attributes the cloud's policy reserves for admin are read-only, and the two
// binding fields the SDK does not return are absent entirely.
func TestNetworkingPortV2SchemaOwnership(t *testing.T) {
	s := resourceNetworkingPortV2().Schema

	for _, name := range []string{"project_id", "region", "network_id"} {
		t.Run(name+" is required and ForceNew", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		})
	}

	for _, name := range []string{"port_security_enabled", "qos_policy_id", "dhcp_blocked", "blocked"} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional, name+" must not be a configuration argument")
			assert.False(t, s[name].Required)
		})
	}

	// binding.host_id and binding.profile are admin-only and are not returned by
	// the SDK, so the block exposes only vnic_type.
	t.Run("the binding block exposes only vnic_type", func(t *testing.T) {
		require.Contains(t, s, "binding")
		nested, ok := s["binding"].Elem.(*schema.Resource)
		require.True(t, ok)

		keys := make([]string, 0, len(nested.Schema))
		for name := range nested.Schema {
			keys = append(keys, name)
		}
		assert.ElementsMatch(t, []string{"vnic_type"}, keys)
	})

	// REQ-042: mac_address, device_id and device_owner have no update path, so a
	// change replaces the port.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "network_id", "mac_address", "device_id", "device_owner",
		}, forceNew)
	})

	// REQ-044: the two mutually exclusive pairs are rejected by the schema.
	t.Run("mutually exclusive attributes conflict", func(t *testing.T) {
		assert.Equal(t, []string{"no_fixed_ip"}, s["fixed_ip"].ConflictsWith)
		assert.Equal(t, []string{"fixed_ip"}, s["no_fixed_ip"].ConflictsWith)
		assert.Equal(t, []string{"no_security_groups"}, s["security_group_ids"].ConflictsWith)
		assert.Equal(t, []string{"security_group_ids"}, s["no_security_groups"].ConflictsWith)
	})

	// The actual state of both sets is reported separately from the configured
	// value, because the API may add to it.
	t.Run("actual sets are exported separately", func(t *testing.T) {
		for _, name := range []string{"all_fixed_ips", "all_security_group_ids", "all_tags"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional)
		}
	})
}

// TestNetworkingPortV2VNICTypeValidation covers SCN-044-04: a vnic_type outside
// the accepted set is rejected before the API is called.
func TestNetworkingPortV2VNICTypeValidation(t *testing.T) {
	nested, ok := resourceNetworkingPortV2().Schema["binding"].Elem.(*schema.Resource)
	require.True(t, ok)

	validate := nested.Schema["vnic_type"].ValidateFunc
	require.NotNil(t, validate)

	for _, value := range networkingPortV2VNICTypes {
		t.Run("accepts "+value, func(t *testing.T) {
			_, errs := validate(value, "vnic_type")
			assert.Empty(t, errs)
		})
	}

	t.Run("rejects an unknown vnic_type", func(t *testing.T) {
		_, errs := validate("not-a-vnic-type", "vnic_type")
		assert.NotEmpty(t, errs)
	})
}

// TestNetworkingPortV2ConfigValidation covers SCN-044-01 and SCN-044-02.
func TestNetworkingPortV2ConfigValidation(t *testing.T) {
	newData := func(raw map[string]interface{}) *schema.ResourceData {
		return schema.TestResourceDataRaw(t, resourceNetworkingPortV2().Schema, raw)
	}

	t.Run("no_fixed_ip with a non-empty fixed_ip is rejected", func(t *testing.T) {
		err := validateNetworkingPortV2Config(newData(map[string]interface{}{
			"no_fixed_ip": true,
			"fixed_ip":    []interface{}{map[string]interface{}{"subnet_id": "s-1"}},
		}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no_fixed_ip")
	})

	t.Run("no_security_groups with a non-empty set is rejected", func(t *testing.T) {
		err := validateNetworkingPortV2Config(newData(map[string]interface{}{
			"no_security_groups": true,
			"security_group_ids": []interface{}{"sg-1"},
		}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no_security_groups")
	})

	t.Run("a valid configuration passes", func(t *testing.T) {
		err := validateNetworkingPortV2Config(newData(map[string]interface{}{
			"network_id": "n-1",
			"fixed_ip":   []interface{}{map[string]interface{}{"subnet_id": "s-1"}},
		}))

		assert.NoError(t, err)
	})
}

func TestNetworkingPortV2Expanders(t *testing.T) {
	t.Run("fixed ips round trip", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{"subnet_id": "s-1", "ip_address": "192.168.199.10"},
		}

		fixedIPs := expandNetworkingPortV2FixedIPs(raw)

		require.Len(t, fixedIPs, 1)
		assert.Equal(t, port.FixedIP{SubnetID: "s-1", IPAddress: "192.168.199.10"}, fixedIPs[0])
		assert.Equal(t, raw, flattenNetworkingPortV2FixedIPs(fixedIPs))
	})

	t.Run("allowed address pairs round trip", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{"ip_address": "192.168.199.20", "mac_address": "fa:16:3e:00:00:01"},
		}

		pairs := expandNetworkingPortV2AllowedAddressPairs(raw)

		require.Len(t, pairs, 1)
		assert.Equal(t, "192.168.199.20", pairs[0].IPAddress)
		assert.Equal(t, raw, flattenNetworkingPortV2AllowedAddressPairs(pairs))
	})

	t.Run("extra dhcp options round trip with the ip version", func(t *testing.T) {
		raw := []interface{}{
			map[string]interface{}{"name": "tftp-server", "value": "10.0.0.1", "ip_version": 4},
		}

		options := expandNetworkingPortV2ExtraDHCPOptions(raw)

		require.Len(t, options, 1)
		assert.Equal(t, "tftp-server", options[0].Name)
		require.NotNil(t, options[0].IPVersion)
		assert.Equal(t, networkingSubnetV2IPVersion4, *options[0].IPVersion)
		assert.Equal(t, raw, flattenNetworkingPortV2ExtraDHCPOptions(options))
	})

	t.Run("dns assignment is flattened", func(t *testing.T) {
		assignments := []port.DNSAssignment{
			{IPAddress: "192.168.199.10", Hostname: "host-1", FQDN: "host-1.example.com."},
		}

		assert.Equal(t, []interface{}{map[string]interface{}{
			"ip_address": "192.168.199.10",
			"hostname":   "host-1",
			"fqdn":       "host-1.example.com.",
		}}, flattenNetworkingPortV2DNSAssignment(assignments))
	})
}

// TestNetworkingPortV2VNICTypeReader checks the single-block accessor used for
// the binding block.
func TestNetworkingPortV2VNICTypeReader(t *testing.T) {
	s := resourceNetworkingPortV2().Schema

	t.Run("reads the configured vnic type", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{
			"binding": []interface{}{map[string]interface{}{"vnic_type": "direct"}},
		})

		assert.Equal(t, "direct", networkingPortV2VNICType(d))
	})

	t.Run("returns an empty value without a binding block", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{})

		assert.Empty(t, networkingPortV2VNICType(d))
	})
}

// TestNetworkingPortV2DataSourceSelectors covers REQ-047: the partial-match
// selectors are applied locally, and no admin-managed attribute is selectable.
func TestNetworkingPortV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingPortV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "port_id", "name", "description", "admin_state_up",
		"network_id", "device_owner", "mac_address", "device_id", "status",
		"fixed_ip", "security_group_ids", "tags", "dns_name",
	}, selectors)

	for _, name := range []string{"port_security_enabled", "qos_policy_id", "dhcp_blocked", "blocked"} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed)
		assert.False(t, s[name].Optional, name+" must be returned, not selectable")
	}
}

// TestNetworkingPortV2Filter covers the two partial-match rules of REQ-047:
// fixed_ip matches any of the port's addresses, and security_group_ids matches a
// port sharing at least one group.
func TestNetworkingPortV2Filter(t *testing.T) {
	candidates := []port.Port{
		{
			ID:             "p-1",
			FixedIPs:       []port.FixedIP{{SubnetID: "s-1", IPAddress: "192.168.199.10"}},
			SecurityGroups: []string{"sg-1", "sg-2"},
		},
		{
			ID:             "p-2",
			FixedIPs:       []port.FixedIP{{SubnetID: "s-1", IPAddress: "192.168.199.11"}},
			SecurityGroups: []string{"sg-3"},
		},
	}

	newData := func(raw map[string]interface{}) *schema.ResourceData {
		return schema.TestResourceDataRaw(t, dataSourceNetworkingPortV2().Schema, raw)
	}

	t.Run("fixed_ip matches one of several addresses", func(t *testing.T) {
		matched := filterNetworkingPortV2(newData(map[string]interface{}{
			"fixed_ip": "192.168.199.11",
		}), candidates)

		require.Len(t, matched, 1)
		assert.Equal(t, "p-2", matched[0].ID)
	})

	t.Run("an address no port has matches nothing", func(t *testing.T) {
		matched := filterNetworkingPortV2(newData(map[string]interface{}{
			"fixed_ip": "192.168.199.99",
		}), candidates)

		assert.Empty(t, matched)
	})

	t.Run("security_group_ids matches an intersection, not an exact set", func(t *testing.T) {
		matched := filterNetworkingPortV2(newData(map[string]interface{}{
			"security_group_ids": []interface{}{"sg-2", "sg-9"},
		}), candidates)

		require.Len(t, matched, 1)
		assert.Equal(t, "p-1", matched[0].ID)
	})

	t.Run("both selectors apply together", func(t *testing.T) {
		matched := filterNetworkingPortV2(newData(map[string]interface{}{
			"fixed_ip":           "192.168.199.10",
			"security_group_ids": []interface{}{"sg-3"},
		}), candidates)

		assert.Empty(t, matched, "the port matching the address has none of the groups")
	})

	t.Run("without local selectors everything passes through", func(t *testing.T) {
		matched := filterNetworkingPortV2(newData(map[string]interface{}{}), candidates)

		assert.Len(t, matched, 2)
	})
}
