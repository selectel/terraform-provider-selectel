package selectel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccBlockStorageSweepProjectEnv = "SELECTEL_ACC_BLOCKSTORAGE_SWEEP_PROJECT_ID"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("selectel_compute_volume_v3", &resource.Sweeper{
		Name: "selectel_compute_volume_v3",
		F:    testAccSweepBlockStorage,
	})
}

func testAccSweepBlockStorage(region string) error {
	projectID := strings.TrimSpace(os.Getenv(testAccBlockStorageSweepProjectEnv))
	if projectID == "" || strings.TrimSpace(region) == "" {
		return fmt.Errorf("set %s to a dedicated, idle acceptance project and pass its region with -sweep", testAccBlockStorageSweepProjectEnv)
	}
	client, err := testAccComputeSnapshotV3ClientForScope(projectID, region)
	if err != nil {
		return err
	}

	return testAccSweepBlockStorageFixtures(client)
}

func testAccSweepBlockStorageFixtures(client *gophercloud.ServiceClient) error {
	volumePages, err := volumes.List(client, nil).AllPages()
	if err != nil {
		return err
	}
	allVolumes, err := volumes.ExtractVolumes(volumePages)
	if err != nil {
		return err
	}
	ownedVolumes := make(map[string]volumes.Volume)
	for _, candidate := range allVolumes {
		if !testAccBlockStorageFixtureName(candidate.Name) ||
			(candidate.Metadata[computeVolumeV3ReservedMetadataKey] == "" &&
				candidate.Metadata["purpose"] != "snapshot-acceptance") {
			continue
		}
		if len(candidate.Attachments) != 0 || candidate.Status == "in-use" {
			return fmt.Errorf("refusing to sweep attached test volume %s; detach it before cleanup", candidate.ID)
		}
		ownedVolumes[candidate.ID] = candidate
	}

	ownedSnapshots, err := testAccBlockStorageSnapshotIDs(client, ownedVolumes)
	if err != nil {
		return err
	}
	// ponytail: fixtures have one clone layer; order by dependencies if tests add clone chains.
	for volumeID, candidate := range ownedVolumes {
		if candidate.SnapshotID == "" && candidate.SourceVolID == "" {
			continue
		}
		if err := testAccDeleteComputeSnapshotV3Volume(client, volumeID); err != nil {
			return err
		}
		delete(ownedVolumes, volumeID)
	}
	for _, snapshotID := range ownedSnapshots {
		if err := testAccDeleteComputeSnapshotV3Snapshot(client, snapshotID); err != nil {
			return err
		}
	}
	// Snapshot deletion must finish before deleting source volumes.
	for volumeID := range ownedVolumes {
		if err := testAccDeleteComputeSnapshotV3Volume(client, volumeID); err != nil {
			return err
		}
	}

	return nil
}

func testAccBlockStorageSnapshotIDs(client *gophercloud.ServiceClient, ownedVolumes map[string]volumes.Volume) ([]string, error) {
	snapshotPages, err := snapshots.List(client, nil).AllPages()
	if err != nil {
		return nil, err
	}
	allSnapshots, err := snapshots.ExtractSnapshots(snapshotPages)
	if err != nil {
		return nil, err
	}
	ownedSnapshots := make([]string, 0, len(allSnapshots))
	for _, candidate := range allSnapshots {
		if !testAccBlockStorageFixtureName(candidate.Name) {
			continue
		}
		// The summary list may omit metadata and volume_id; confirm ownership with GET.
		current, err := snapshots.Get(client, candidate.ID).Extract()
		if testAccComputeSnapshotV3NotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if _, owned := ownedVolumes[current.VolumeID]; !owned || current.Metadata["purpose"] != "snapshot-acceptance" ||
			!testAccBlockStorageFixtureName(current.Name) {
			continue
		}
		ownedSnapshots = append(ownedSnapshots, current.ID)
	}

	return ownedSnapshots, nil
}

func testAccBlockStorageFixtureName(name string) bool {
	return strings.HasPrefix(name, "tf-acc-selectel-volume-") ||
		strings.HasPrefix(name, "tf-acc-selectel-snapshot-")
}

func TestUnitBlockStorageSweeperRequiresScope(t *testing.T) {
	t.Setenv(testAccBlockStorageSweepProjectEnv, "")
	require.ErrorContains(t, testAccSweepBlockStorage("ru-1"), testAccBlockStorageSweepProjectEnv)
	t.Setenv(testAccBlockStorageSweepProjectEnv, "test-project")
	require.ErrorContains(t, testAccSweepBlockStorage(""), "-sweep")
}

func TestUnitBlockStorageSweeper(t *testing.T) {
	for _, scenario := range []struct {
		name        string
		attachments []volumes.Attachment
	}{
		{name: "cleanup"},
		{name: "attached", attachments: []volumes.Attachment{{}}},
		{name: "snapshot deletion fails"},
		{name: "volume deletion fails"},
		{name: "list fails"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			owned := volumes.Volume{
				ID: "owned-volume", Name: "tf-acc-selectel-volume-test", Status: "available",
				Metadata: map[string]string{"purpose": "snapshot-acceptance"},
			}
			owned.Attachments = scenario.attachments
			ownedSnapshot := snapshots.Snapshot{
				ID: "owned-snapshot", Name: "tf-acc-selectel-snapshot-test", VolumeID: owned.ID,
				Metadata: map[string]string{"purpose": "snapshot-acceptance"},
			}
			var deleted []string
			var snapshotGone, snapshotAbsenceChecked, cloneAbsenceChecked bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + r.URL.Path {
				case "GET /volumes/detail":
					if scenario.name == "list fails" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"volumes": []volumes.Volume{
						owned,
						{
							ID: "clone-volume", Name: "tf-acc-selectel-volume-clone", SnapshotID: "owned-snapshot",
							Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: "token"},
						},
						{ID: "prefix-only", Name: "tf-acc-selectel-volume-foreign"},
						{ID: "marker-only", Name: "production", Metadata: owned.Metadata},
					}}))
				case "GET /snapshots":
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"snapshots": []snapshots.Snapshot{
						{ID: ownedSnapshot.ID, Name: ownedSnapshot.Name},
						{ID: "foreign-snapshot", Name: "tf-acc-selectel-snapshot-foreign"},
						{ID: "unmarked-snapshot", Name: "tf-acc-selectel-snapshot-unmarked"},
						{ID: "production-snapshot", Name: "production"},
					}}))
				case "GET /snapshots/owned-snapshot":
					if snapshotGone {
						snapshotAbsenceChecked = true
						w.WriteHeader(http.StatusNotFound)
						return
					}
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"snapshot": ownedSnapshot}))
				case "GET /snapshots/foreign-snapshot":
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"snapshot": snapshots.Snapshot{
						ID: "foreign-snapshot", Name: "tf-acc-selectel-snapshot-foreign", VolumeID: "prefix-only",
						Metadata: ownedSnapshot.Metadata,
					}}))
				case "GET /snapshots/unmarked-snapshot":
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"snapshot": snapshots.Snapshot{
						ID: "unmarked-snapshot", Name: "tf-acc-selectel-snapshot-unmarked", VolumeID: owned.ID,
					}}))
				case "DELETE /snapshots/owned-snapshot":
					assert.True(t, cloneAbsenceChecked)
					assert.Equal(t, []string{"clone-volume"}, deleted)
					if scenario.name == "snapshot deletion fails" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					deleted = append(deleted, ownedSnapshot.ID)
					snapshotGone = true
					w.WriteHeader(http.StatusAccepted)
				case "DELETE /volumes/owned-volume":
					assert.True(t, snapshotAbsenceChecked, "snapshot must be deleted before its source volume")
					if scenario.name == "volume deletion fails" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					deleted = append(deleted, owned.ID)
					w.WriteHeader(http.StatusAccepted)
				case "GET /volumes/owned-volume":
					w.WriteHeader(http.StatusNotFound)
				case "DELETE /volumes/clone-volume":
					deleted = append(deleted, "clone-volume")
					w.WriteHeader(http.StatusAccepted)
				case "GET /volumes/clone-volume":
					cloneAbsenceChecked = true
					w.WriteHeader(http.StatusNotFound)
				default:
					t.Errorf("unexpected sweeper request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			t.Cleanup(server.Close)
			client := &gophercloud.ServiceClient{
				ProviderClient: &gophercloud.ProviderClient{}, Endpoint: server.URL + "/",
			}
			err := testAccSweepBlockStorageFixtures(client)
			switch scenario.name {
			case "cleanup":
				require.NoError(t, err)
				assert.Equal(t, []string{"clone-volume", ownedSnapshot.ID, owned.ID}, deleted)
			case "snapshot deletion fails":
				require.Error(t, err)
				assert.Equal(t, []string{"clone-volume"}, deleted)
			case "volume deletion fails":
				require.Error(t, err)
				assert.Equal(t, []string{"clone-volume", ownedSnapshot.ID}, deleted)
			default:
				require.Error(t, err)
				assert.Empty(t, deleted)
			}
		})
	}
}
