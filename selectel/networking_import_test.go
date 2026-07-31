package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNetworkingImportProjectID = "8ac3d7f5b1a4c2e9b0d6f4a1c7e2b5d8"
	testNetworkingImportRegion    = "ru-9"
)

// networkingImportTestSchema is the minimal schema every Neutron resource of
// this migration shares: project_id and region as required, ForceNew fields.
func networkingImportTestSchema() map[string]*schema.Schema {
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

func TestResourceNetworkingImportProjectRegion(t *testing.T) {
	t.Run("sets both fields from the provider configuration", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, networkingImportTestSchema(), map[string]interface{}{})
		config := &Config{
			ProjectID: testNetworkingImportProjectID,
			Region:    testNetworkingImportRegion,
		}

		err := resourceNetworkingImportProjectRegion(d, config)

		require.NoError(t, err)
		assert.Equal(t, testNetworkingImportProjectID, d.Get("project_id").(string))
		assert.Equal(t, testNetworkingImportRegion, d.Get("region").(string))
	})

	t.Run("fails when the provider has no project_id", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, networkingImportTestSchema(), map[string]interface{}{})
		config := &Config{Region: testNetworkingImportRegion}

		err := resourceNetworkingImportProjectRegion(d, config)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "INFRA_PROJECT_ID")
		assert.Empty(t, d.Get("project_id").(string))
	})

	t.Run("fails when the provider has no region", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, networkingImportTestSchema(), map[string]interface{}{})
		config := &Config{ProjectID: testNetworkingImportProjectID}

		err := resourceNetworkingImportProjectRegion(d, config)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "INFRA_REGION")
	})
}
