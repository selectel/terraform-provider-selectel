package selectel

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSelectelComputeVolumeV3_Replacement(t *testing.T) {
	testAccBlockStorageSetup(t)
	initialProjectID := os.Getenv("INFRA_PROJECT_ID")
	replacementProjectID := testAccExtraProject(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-replacement")
	initialConfig := testAccComputeVolumeV3Config(name, "", "", 1)
	replacementConfig := testAccComputeVolumeV3ConfigWithScope(
		replacementProjectID, testAccBlockStorageRegion, testAccBlockStorageZone,
		testAccBlockStorageType, name, 1, "", "",
	)
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check:  testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, &initialID),
			},
			{
				Config: replacementConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3ID(testAccComputeVolumeV3ResourceName, &initialID, true),
					testAccCheckComputeVolumeV3DeletedInScope(t, &initialID, initialProjectID, testAccBlockStorageRegion),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "project_id", replacementProjectID),
				),
			},
		},
	})
}
