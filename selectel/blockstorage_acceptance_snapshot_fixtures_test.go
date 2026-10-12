package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	gophersnapshots "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/snapshot"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAccCreateComputeSnapshotV3Fixture(
	t *testing.T,
	prefix string,
	volumeOpts volume.CreateOpts,
	snapshotCount int,
) testAccComputeSnapshotV3Fixture {
	t.Helper()

	client, err := testAccBlockStorageClient(t, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"))
	require.NoError(t, err)

	volumeOpts.Name = prefix + "-volume"
	volumeOpts.Metadata = map[string]string{"purpose": "snapshot-acceptance"}
	createdVolume, _, err := volume.Create(t.Context(), client, volumeOpts)
	require.NoError(t, err)
	require.NotEmpty(t, createdVolume.ID)
	t.Cleanup(func() {
		assert.NoError(t, testAccDeleteBlockStorageVolume(client, createdVolume.ID))
	})
	require.NoError(
		t,
		gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
			current, _, err := volume.Get(t.Context(), client, createdVolume.ID)
			if err != nil {
				return false, err
			}

			return current.Status == "available", nil
		}),
	)

	fixture := testAccComputeSnapshotV3Fixture{
		VolumeID: createdVolume.ID,
	}
	// The SDK only supports snapshot reads; writes remain test fixture operations.
	var snapshotClient *gophercloud.ServiceClient
	if snapshotCount > 0 {
		snapshotClient, err = testAccOpenStackServiceClient(os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), BlockStorage)
		require.NoError(t, err)
	}
	for index := 1; index <= snapshotCount; index++ {
		createdSnapshot, err := gophersnapshots.Create(snapshotClient, gophersnapshots.CreateOpts{
			VolumeID:    createdVolume.ID,
			Name:        fmt.Sprintf("%s-%d", prefix, index),
			Description: fmt.Sprintf("Terraform acceptance snapshot %d", index),
			Metadata:    map[string]string{"purpose": "snapshot-acceptance", "index": strconv.Itoa(index)},
		}).Extract()
		require.NoError(t, err)
		require.NotEmpty(t, createdSnapshot.ID)
		snapshotID := createdSnapshot.ID
		t.Cleanup(func() {
			assert.NoError(t, testAccDeleteComputeSnapshotV3Snapshot(client, snapshotClient, snapshotID))
		})
		require.NoError(
			t,
			gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
				current, _, err := snapshot.Get(t.Context(), client, snapshotID)
				if err != nil {
					return false, err
				}

				return current.Status == "available", nil
			}),
		)

		current, _, err := snapshot.Get(t.Context(), client, snapshotID)
		require.NoError(t, err)
		fixture.Snapshots = append(fixture.Snapshots, *current)
	}

	return fixture
}

func testAccDeleteComputeSnapshotV3Snapshot(
	client *blockstorage.Client,
	snapshotClient *gophercloud.ServiceClient,
	snapshotID string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), testAccComputeSnapshotV3Timeout*time.Second)
	defer cancel()

	if err := gophersnapshots.Delete(snapshotClient, snapshotID).ExtractErr(); err != nil {
		if testAccBlockStorageNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete Block Storage snapshot %s: %w", snapshotID, err)
	}

	err := testAccBlockStorageWaitNotFound(func() error {
		_, _, err := snapshot.Get(ctx, client, snapshotID)

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to wait for Block Storage snapshot %s deletion: %w", snapshotID, err)
	}

	return nil
}

func testAccDeleteBlockStorageVolume(
	client *blockstorage.Client,
	volumeID string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), testAccComputeSnapshotV3Timeout*time.Second)
	defer cancel()

	if _, err := volume.Delete(ctx, client, volumeID); err != nil {
		if testAccBlockStorageNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete Block Storage volume %s: %w", volumeID, err)
	}

	err := testAccBlockStorageWaitNotFound(func() error {
		_, _, err := volume.Get(ctx, client, volumeID)

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to wait for Block Storage volume %s deletion: %w", volumeID, err)
	}

	return nil
}

func testAccBlockStorageNotFound(err error) bool {
	var notFound gophercloud.ErrDefault404

	return blockstorage.IsKind(err, blockstorage.KindNotFound) || errors.As(err, &notFound)
}

func testAccBlockStorageWaitNotFound(get func() error) error {
	return gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
		err := get()
		if err == nil {
			return false, nil
		}
		if testAccBlockStorageNotFound(err) {
			return true, nil
		}

		return false, err
	})
}
