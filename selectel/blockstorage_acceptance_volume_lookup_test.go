package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSelectelComputeVolumeLookupV3(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-lookup")
	config := testAccComputeVolumeV3Config(name, "lookup test", "lookup-acceptance", 1) + `
data "selectel_compute_volume_v3" "by_id" {
  project_id = selectel_compute_volume_v3.test.project_id
  region     = selectel_compute_volume_v3.test.region
  volume_id  = selectel_compute_volume_v3.test.id
}

data "selectel_compute_volume_v3" "by_name" {
  project_id = selectel_compute_volume_v3.test.project_id
  region     = selectel_compute_volume_v3.test.region
  name       = selectel_compute_volume_v3.test.name
  status     = "available"

  metadata = {
    purpose = selectel_compute_volume_v3.test.metadata["purpose"]
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeTestCheckFunc(
				testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
				resource.TestCheckResourceAttrPair("data.selectel_compute_volume_v3.by_id", "id", testAccComputeVolumeV3ResourceName, "id"),
				resource.TestCheckResourceAttrPair("data.selectel_compute_volume_v3.by_name", "id", testAccComputeVolumeV3ResourceName, "id"),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_v3.by_id", "name", name),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_v3.by_id", "size", "1"),
				resource.TestCheckResourceAttr("data.selectel_compute_volume_v3.by_id", "metadata.purpose", "lookup-acceptance"),
				resource.TestCheckNoResourceAttr("data.selectel_compute_volume_v3.by_id", "metadata."+testAccBlockStorageCreateToken),
			),
		}},
	})
}
