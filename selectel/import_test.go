package selectel

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testVPCImportProjectID = "8ac3d7f5b1a4c2e9b0d6f4a1c7e2b5d8"
	testVPCImportRegion    = "ru-9"
)

func vpcImportTestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}

func TestResourceVPCImportProjectRegion(t *testing.T) {
	t.Run("sets both fields from the provider configuration", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
		config := &Config{
			ProjectID: testVPCImportProjectID,
			Region:    testVPCImportRegion,
		}

		err := importProjectRegion(d, config)

		require.NoError(t, err)
		assert.Equal(t, testVPCImportProjectID, d.Get("project_id").(string))
		assert.Equal(t, testVPCImportRegion, d.Get("region").(string))
	})

	t.Run("fails when the provider has no project_id", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
		config := &Config{Region: testVPCImportRegion}

		err := importProjectRegion(d, config)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "INFRA_PROJECT_ID")
		assert.Empty(t, d.Get("project_id").(string))
	})

	t.Run("fails when the provider has no region", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
		config := &Config{ProjectID: testVPCImportProjectID}

		err := importProjectRegion(d, config)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "INFRA_REGION")
	})
}

func vpcImportTestIdentitySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":         {Type: schema.TypeString, RequiredForImport: true},
		"project_id": {Type: schema.TypeString, RequiredForImport: true},
		"region":     {Type: schema.TypeString, RequiredForImport: true},
	}
}

func TestResourceVPCImportProjectRegionByIdentity(t *testing.T) {
	t.Run("takes id, project_id and region from the identity", func(t *testing.T) {
		d := schema.TestResourceDataWithIdentityRaw(t, vpcImportTestSchema(), vpcImportTestIdentitySchema(), map[string]string{
			"id":         "b311ce58-2658-46b5-b733-7a0f418703f2",
			"project_id": testVPCImportProjectID,
			"region":     testVPCImportRegion,
		})

		err := importProjectRegion(d, &Config{})

		require.NoError(t, err)
		assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", d.Id())
		assert.Equal(t, testVPCImportProjectID, d.Get("project_id").(string))
		assert.Equal(t, testVPCImportRegion, d.Get("region").(string))
	})

	t.Run("falls back to the provider configuration when the identity is empty", func(t *testing.T) {
		d := schema.TestResourceDataWithIdentityRaw(t, vpcImportTestSchema(), vpcImportTestIdentitySchema(), map[string]string{})
		config := &Config{ProjectID: testVPCImportProjectID, Region: testVPCImportRegion}

		err := importProjectRegion(d, config)

		require.NoError(t, err)
		assert.Equal(t, testVPCImportProjectID, d.Get("project_id").(string))
		assert.Equal(t, testVPCImportRegion, d.Get("region").(string))
	})
}

func TestSetProjectRegionIdentity(t *testing.T) {
	t.Run("stores id, project_id and region in the identity", func(t *testing.T) {
		d := schema.TestResourceDataWithIdentityRaw(t, vpcImportTestSchema(), vpcImportTestIdentitySchema(), map[string]string{})
		d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")
		require.NoError(t, d.Set("project_id", testVPCImportProjectID))
		require.NoError(t, d.Set("region", testVPCImportRegion))

		setProjectRegionIdentity(d)

		identity, err := d.Identity()
		require.NoError(t, err)
		assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", identity.Get("id"))
		assert.Equal(t, testVPCImportProjectID, identity.Get("project_id"))
		assert.Equal(t, testVPCImportRegion, identity.Get("region"))
	})

	t.Run("does nothing for a resource without identity", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
		d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")

		assert.NotPanics(t, func() { setProjectRegionIdentity(d) })
	})
}

const (
	testVPCParentID = "b311ce58-2658-46b5-b733-7a0f418703f2"
	testVPCRouteID  = testVPCParentID + "-route-10.0.0.0/8-192.168.199.10"
	testVPCRuleID   = "e3f2a1d4-5b6c-4d7e-8f90-a1b2c3d4e5f6"
)

func vpcIdentityTestData(t *testing.T, res *schema.Resource, identity map[string]string) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataWithIdentityRaw(t, res.Schema, res.Identity.SchemaFunc(), identity)
}

func TestVPCRoutesImportByIdentity(t *testing.T) {
	t.Run("subnet route", func(t *testing.T) {
		res := resourceVPCNetworkSubnetRouteV2()
		d := vpcIdentityTestData(t, res, map[string]string{
			"project_id":       testVPCImportProjectID,
			"region":           testVPCImportRegion,
			"subnet_id":        testVPCParentID,
			"destination_cidr": "10.0.0.0/8",
			"next_hop":         "192.168.199.10",
		})

		_, err := res.Importer.StateContext(context.Background(), d, &Config{})

		require.NoError(t, err)
		assert.Equal(t, testVPCRouteID, d.Id())
		assert.Equal(t, testVPCParentID, d.Get("subnet_id"))
		assert.Equal(t, "10.0.0.0/8", d.Get("destination_cidr"))
		assert.Equal(t, "192.168.199.10", d.Get("next_hop"))
		assert.Equal(t, testVPCImportProjectID, d.Get("project_id"))
		assert.Equal(t, testVPCImportRegion, d.Get("region"))
	})

	t.Run("router route", func(t *testing.T) {
		res := resourceVPCRouterRouteV2()
		d := vpcIdentityTestData(t, res, map[string]string{
			"project_id":       testVPCImportProjectID,
			"region":           testVPCImportRegion,
			"router_id":        testVPCParentID,
			"destination_cidr": "10.0.0.0/8",
			"next_hop":         "192.168.199.10",
		})

		_, err := res.Importer.StateContext(context.Background(), d, &Config{})

		require.NoError(t, err)
		assert.Equal(t, testVPCRouteID, d.Id())
		assert.Equal(t, testVPCParentID, d.Get("router_id"))
		assert.Equal(t, "10.0.0.0/8", d.Get("destination_cidr"))
		assert.Equal(t, "192.168.199.10", d.Get("next_hop"))
	})

	t.Run("string ID still works", func(t *testing.T) {
		res := resourceVPCRouterRouteV2()
		d := schema.TestResourceDataRaw(t, res.Schema, map[string]any{})
		d.SetId(testVPCRouteID)
		config := &Config{ProjectID: testVPCImportProjectID, Region: testVPCImportRegion}

		_, err := res.Importer.StateContext(context.Background(), d, config)

		require.NoError(t, err)
		assert.Equal(t, testVPCParentID, d.Get("router_id"))
		assert.Equal(t, testVPCImportProjectID, d.Get("project_id"))
	})
}

func TestVPCPortForwardingImportByIdentity(t *testing.T) {
	res := resourceVPCPortForwardingV2()
	d := vpcIdentityTestData(t, res, map[string]string{
		"id":            testVPCRuleID,
		"floatingip_id": testVPCParentID,
		"project_id":    testVPCImportProjectID,
		"region":        testVPCImportRegion,
	})

	_, err := res.Importer.StateContext(context.Background(), d, &Config{})

	require.NoError(t, err)
	assert.Equal(t, testVPCRuleID, d.Id())
	assert.Equal(t, testVPCParentID, d.Get("floatingip_id"))
	assert.Equal(t, testVPCImportProjectID, d.Get("project_id"))
	assert.Equal(t, testVPCImportRegion, d.Get("region"))
}

func TestSetIdentity(t *testing.T) {
	t.Run("route stores its parts", func(t *testing.T) {
		res := resourceVPCNetworkSubnetRouteV2()
		d := vpcIdentityTestData(t, res, map[string]string{})
		d.SetId(testVPCRouteID)
		require.NoError(t, d.Set("project_id", testVPCImportProjectID))
		require.NoError(t, d.Set("region", testVPCImportRegion))
		require.NoError(t, d.Set("subnet_id", testVPCParentID))
		require.NoError(t, d.Set("destination_cidr", "10.0.0.0/8"))
		require.NoError(t, d.Set("next_hop", "192.168.199.10"))

		setIdentity(d, "subnet_id", "destination_cidr", "next_hop")

		identity, err := d.Identity()
		require.NoError(t, err)
		assert.Equal(t, testVPCParentID, identity.Get("subnet_id"))
		assert.Equal(t, "10.0.0.0/8", identity.Get("destination_cidr"))
		assert.Equal(t, "192.168.199.10", identity.Get("next_hop"))
		assert.Equal(t, testVPCImportProjectID, identity.Get("project_id"))
	})

	t.Run("port forwarding stores the rule ID and the floating IP", func(t *testing.T) {
		res := resourceVPCPortForwardingV2()
		d := vpcIdentityTestData(t, res, map[string]string{})
		d.SetId(testVPCRuleID)
		require.NoError(t, d.Set("floatingip_id", testVPCParentID))
		require.NoError(t, d.Set("project_id", testVPCImportProjectID))
		require.NoError(t, d.Set("region", testVPCImportRegion))

		setIdentity(d, "id", "floatingip_id")

		identity, err := d.Identity()
		require.NoError(t, err)
		assert.Equal(t, testVPCRuleID, identity.Get("id"))
		assert.Equal(t, testVPCParentID, identity.Get("floatingip_id"))
	})
}
