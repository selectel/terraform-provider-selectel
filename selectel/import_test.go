package selectel

import (
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
