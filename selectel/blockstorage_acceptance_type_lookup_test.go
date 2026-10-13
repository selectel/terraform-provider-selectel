package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
	"github.com/stretchr/testify/require"
)

func TestAccSelectelComputeVolumeTypeV3(t *testing.T) {
	testAccBlockStorageSetup(t)
	project, region := os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION")
	client, err := testAccBlockStorageClient(t, project, region)
	require.NoError(t, err)
	defaultType, _, err := volumetype.Get(t.Context(), client, volumetype.DefaultTypeID)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
data "selectel_compute_volume_type_v3" "default" {
  project_id     = %[1]q
  region         = %[2]q
  volume_type_id = "default"
}

data "selectel_compute_volume_type_v3" "by_id" {
  project_id     = %[1]q
  region         = %[2]q
  volume_type_id = %[3]q
}

data "selectel_compute_volume_type_v3" "by_name" {
  project_id = %[1]q
  region     = %[2]q
  name       = %[4]q
}
`, project, region, defaultType.ID, defaultType.Name),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.selectel_compute_volume_type_v3.default", "id", defaultType.ID),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_type_v3.default", "name", defaultType.Name),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_type_v3.by_id", "id", defaultType.ID),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_type_v3.by_name", "id", defaultType.ID),
			),
		}},
	})
}
