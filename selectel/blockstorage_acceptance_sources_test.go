package selectel

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/backups"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These fixtures belong to the test: no existing user image or backup is changed.
func TestAccSelectelComputeVolumeV3_ImageSource(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-image")
	client, err := testAccOpenStackServiceClient(os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), "image")
	require.NoError(t, err)
	// The image catalog advertises the unversioned endpoint.
	if !strings.HasSuffix(strings.TrimRight(client.Endpoint, "/"), "/v2") {
		client.Endpoint = strings.TrimRight(client.Endpoint, "/") + "/v2/"
	}
	visibility := images.ImageVisibilityPrivate
	image, err := images.Create(client, images.CreateOpts{
		Name: name, DiskFormat: "raw", ContainerFormat: "bare", MinDisk: 1, Visibility: &visibility,
	}).Extract()
	require.NoError(t, err)
	require.NotEmpty(t, image.ID)
	t.Logf("Created acceptance image %s", image.ID)
	t.Cleanup(func() {
		assert.NoError(t, images.Delete(client, image.ID).ExtractErr())
		assert.NoError(t, gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
			current, err := images.Get(client, image.ID).Extract()
			if testAccBlockStorageNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}

			// Glance can expose a soft-deleted image with HTTP 200.
			return current.Status == "deleted", nil
		}))
	})
	// A small blank raw image exercises the image source path without an OS/VM fixture.
	require.NoError(t, imagedata.Upload(client, image.ID, bytes.NewReader(make([]byte, 1024*1024))).ExtractErr())
	require.NoError(t, gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
		current, err := images.Get(client, image.ID).Extract()
		if err != nil {
			return false, err
		}
		if current.Status == "killed" {
			return false, fmt.Errorf("acceptance image %s failed to upload", image.ID)
		}

		return current.Status == "active", nil
	}))
	testAccComputeVolumeV3Source(t, name, "image_id", image.ID)
}

func TestAccSelectelComputeVolumeV3_BackupSource(t *testing.T) {
	testAccBlockStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-backup")
	source := testAccCreateComputeSnapshotV3Fixture(t, name, volume.CreateOpts{
		Size: 1, AvailabilityZone: testAccBlockStorageZone,
		VolumeType: testAccBlockStorageType,
	}, 0)
	client, err := testAccOpenStackServiceClient(os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), BlockStorage)
	require.NoError(t, err)
	backup, err := backups.Create(client, backups.CreateOpts{VolumeID: source.VolumeID, Name: name}).Extract()
	require.NoError(t, err)
	require.NotEmpty(t, backup.ID)
	t.Logf("Created acceptance backup %s", backup.ID)
	t.Cleanup(func() {
		assert.NoError(t, backups.Delete(client, backup.ID).ExtractErr())
		assert.NoError(t, testAccBlockStorageWaitNotFound(func() error {
			_, err := backups.Get(client, backup.ID).Extract()
			return err
		}))
	})
	require.NoError(t, gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
		current, err := backups.Get(client, backup.ID).Extract()
		if err != nil {
			return false, err
		}
		if current.Status == "error" {
			return false, fmt.Errorf("acceptance backup %s failed", backup.ID)
		}

		return current.Status == "available", nil
	}))
	testAccComputeVolumeV3Source(t, name, "backup_id", backup.ID)
}

func testAccComputeVolumeV3Source(t *testing.T, name, field, sourceID string) {
	t.Helper()
	config := testAccComputeVolumeV3FromFixtureConfig(name, field, sourceID, 1)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSelectelPreCheckWithAuth(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeTestCheckFunc(
				testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
				resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, field, sourceID),
				resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", "1"),
			)},
			testAccComputeVolumeV3ImportStep(),
		},
	})
}
