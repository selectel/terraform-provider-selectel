package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/snapshot"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

const testAccComputeSnapshotV3Timeout = 600

type testAccComputeSnapshotV3Fixture struct {
	VolumeID  string
	Snapshots []snapshot.View
}

func TestAccSelectelComputeSnapshotV3(t *testing.T) {
	testAccBlockStorageSetup(t)
	fixture := testAccCreateComputeSnapshotV3Fixture(t,
		acctest.RandomWithPrefix("tf-acc-selectel-snapshot"),
		volume.CreateOpts{Size: 1, AvailabilityZone: testAccBlockStorageZone, VolumeType: testAccBlockStorageType}, 1,
	)
	snapshot := fixture.Snapshots[0]
	const resourceName = "data.selectel_compute_snapshot_v3.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
data "selectel_compute_snapshot_v3" "test" {
  project_id = %q
  region     = %q
  name       = %q
  status     = "available"
  volume_id  = %q
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), snapshot.Name, fixture.VolumeID),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "id", snapshot.ID),
				resource.TestCheckResourceAttr(resourceName, "volume_id", fixture.VolumeID),
				resource.TestCheckResourceAttr(resourceName, "description", snapshot.Description),
				resource.TestCheckResourceAttr(resourceName, "size", "1"),
				resource.TestCheckResourceAttr(resourceName, "metadata.purpose", "snapshot-acceptance"),
			),
		}},
	})
}
