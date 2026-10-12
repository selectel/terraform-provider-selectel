package selectel

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/require"
)

func testAccComputeVolumeV3ImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      testAccComputeVolumeV3ResourceName,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			// This local permission is not reported by the API.
			"enable_online_resize",
			// These creation sources are not returned by the Block Storage SDK.
			"image_id", "backup_id",
			// Import does not take ownership of user metadata or restore timeouts.
			"metadata", "timeouts",
		},
	}
}

func TestAccSelectelComputeVolumeV3_Drift(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-drift")
	driftedName := name + "-outside-terraform"
	config := testAccComputeVolumeV3Config(name, "drift test", "drift", 1)
	var volumeID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheckWithAuth(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: testAccCheckComputeVolumeV3Exists(
					t,
					testAccComputeVolumeV3ResourceName,
					&volumeID,
				),
			},
			{
				PreConfig: func() {
					require.NoError(
						t,
						testAccComputeVolumeV3SetRemoteName(
							t,
							volumeID,
							os.Getenv("INFRA_PROJECT_ID"),
							os.Getenv("INFRA_REGION"),
							driftedName,
						),
					)
				},
				RefreshState: true,
				// The external rename must produce a plan restoring the configured name.
				ExpectNonEmptyPlan: true,
				Check: resource.TestCheckResourceAttr(
					testAccComputeVolumeV3ResourceName,
					"name",
					driftedName,
				),
			},
			{
				// The API rejects a metadata update of a volume with a QoS type
				// for a short interval after the volume QoS was last set.
				PreConfig: func() { time.Sleep(10 * time.Second) },
				Config:    config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"name",
						name,
					),
				),
			},
		},
	})
}

func TestAccSelectelComputeVolumeV3_Disappears(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-disappears")
	config := testAccComputeVolumeV3Config(name, "disappears test", "disappears", 1)
	var deletedID string
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheckWithAuth(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				// The Check deletes the volume outside Terraform; the next plan must recreate it.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, &deletedID),
					func(state *terraform.State) error {
						client, err := testAccComputeVolumeV3Client(t, state.RootModule().Resources[testAccComputeVolumeV3ResourceName])
						if err != nil {
							return err
						}

						return testAccDeleteBlockStorageVolume(client, deletedID)
					},
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3ID(testAccComputeVolumeV3ResourceName, &deletedID, true),
				),
			},
		},
	})
}

func testAccComputeVolumeV3SetRemoteName(
	t *testing.T,
	volumeID string,
	projectID string,
	region string,
	name string,
) error {
	t.Helper()

	if volumeID == "" {
		return fmt.Errorf("Block Storage volume ID was not captured before drift")
	}
	client, err := testAccBlockStorageClient(t, projectID, region)
	if err != nil {
		return err
	}

	_, _, err = volume.Update(
		t.Context(),
		client,
		volumeID,
		volume.UpdateOpts{Name: &name},
	)
	if err != nil {
		return fmt.Errorf(
			"failed to introduce external name drift for Block Storage volume %s: %w",
			volumeID,
			err,
		)
	}

	return nil
}

func TestAccSelectelComputeVolumeV3_RegionalTypeName(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-regional-type")
	config := testAccComputeVolumeV3ConfigWithScope(
		os.Getenv("INFRA_PROJECT_ID"),
		testAccBlockStorageRegion,
		testAccBlockStorageZone,
		"universal2.ru-6",
		name,
		1,
		"regional type test",
		"regional-type",
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheckWithAuth(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
				),
			},
		},
	})
}
