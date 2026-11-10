package selectel

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

const (
	testAccBlockStorageCreateToken     = "selectel_tf_create_token" //nolint:gosec // Metadata key, not a credential.
	testAccComputeVolumeV3ResourceName = "selectel_compute_volume_v3.test"
)

func TestAccSelectelComputeVolumeV3_Basic(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-basic")
	initialConfig := testAccComputeVolumeV3Config(name, "initial description", "initial", 1)
	updatedConfig := testAccComputeVolumeV3Config(name+"-updated", "updated description", "updated", 1)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheckWithAuth(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "timeouts.create", "30m"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "timeouts.update", "31m"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "timeouts.delete", "32m"),
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "name", name),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", "1"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "metadata.purpose", "initial"),
					resource.TestCheckNoResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"metadata."+testAccBlockStorageCreateToken,
					),
				),
			},
			{
				// The API rejects a metadata update of a volume with a QoS type
				// for a short interval after the volume QoS was last set.
				PreConfig: func() { time.Sleep(10 * time.Second) },
				Config:    updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "metadata.purpose", "updated"),
				),
			},
			testAccComputeVolumeV3ImportStep(),
		},
	})
}

func TestAccSelectelComputeVolumeV3_Sources(t *testing.T) {
	testAccBlockStorageSetup(t)
	t.Run("volume", func(t *testing.T) {
		sourceName := acctest.RandomWithPrefix("tf-acc-selectel-volume-source")
		cloneName := acctest.RandomWithPrefix("tf-acc-selectel-volume-clone")
		config := testAccComputeVolumeV3FromVolumeConfig(sourceName, cloneName)

		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccSelectelPreCheckWithAuth(t)
			},
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
			Steps: []resource.TestStep{{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(
						t,
						"selectel_compute_volume_v3.source",
						nil,
					),
					testAccCheckComputeVolumeV3Exists(
						t,
						"selectel_compute_volume_v3.test",
						nil,
					),
					resource.TestCheckResourceAttrPair(
						"selectel_compute_volume_v3.test",
						"source_volume_id",
						"selectel_compute_volume_v3.source",
						"id",
					),
				),
			}, testAccComputeVolumeV3ImportStep()},
		})
	})

	t.Run("snapshot", func(t *testing.T) {
		name := acctest.RandomWithPrefix("tf-acc-selectel-volume-snapshot")
		fixture := testAccCreateComputeSnapshotV3Fixture(t, name, volume.CreateOpts{
			Size:             1,
			AvailabilityZone: testAccBlockStorageZone,
			VolumeType:       testAccBlockStorageType,
		}, 1)
		source := fixture.Snapshots[0]
		config := testAccComputeVolumeV3FromFixtureConfig(name, "snapshot_id", source.ID, source.Size)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
						resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "snapshot_id", source.ID),
						resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", strconv.Itoa(source.Size)),
					),
				},
				testAccComputeVolumeV3ImportStep(),
			},
		})
	})
}

func TestAccSelectelComputeVolumeV3_Resize(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-resize")
	initialConfig := testAccComputeVolumeV3Config(name, "resize test", "resize", 1)
	resizedConfig := strings.Replace(
		testAccComputeVolumeV3Config(name, "resize test", "resize", 2),
		"\n  metadata = {",
		"\n  enable_online_resize = true\n\n  metadata = {",
		1,
	)
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheckWithAuth(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(
						t,
						testAccComputeVolumeV3ResourceName,
						&initialID,
					),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", "1"),
				),
			},
			{
				Config: resizedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3ID(testAccComputeVolumeV3ResourceName, &initialID, false),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", "2"),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"enable_online_resize",
						"true",
					),
				),
			},
		},
	})
}
