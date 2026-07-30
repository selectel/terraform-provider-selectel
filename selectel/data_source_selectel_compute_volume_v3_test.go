package selectel

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testComputeVolumeLookupV3ID = "lookup-volume-id"

func TestUnitSelectelComputeVolumeLookupV3ConfigurationValidation(t *testing.T) {
	dataSourceSchema := dataSourceComputeVolumeV3().Schema
	for _, testCase := range []struct {
		name         string
		config       map[string]any
		expectsError bool
	}{
		{
			name: "ID conflicts with search criteria",
			config: map[string]any{
				"project_id": testBlockStorageProjectID,
				"region":     testBlockStorageRegion,
				"volume_id":  testComputeVolumeLookupV3ID,
				"name":       "volume-name",
				"status":     "available",
				"metadata":   map[string]any{"environment": "test"},
			},
			expectsError: true,
		},
		{
			name: "reserved metadata criterion",
			config: map[string]any{
				"project_id": testBlockStorageProjectID,
				"region":     testBlockStorageRegion,
				"metadata": map[string]any{
					computeVolumeV3ReservedMetadataKey: "token",
				},
			},
			expectsError: true,
		},
		{
			name: "criteria",
			config: map[string]any{
				"project_id": testBlockStorageProjectID,
				"region":     testBlockStorageRegion,
				"name":       "volume-name",
				"status":     "available",
				"metadata":   map[string]any{"environment": "test"},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := schema.InternalMap(dataSourceSchema).Validate(
				terraform.NewResourceConfigRaw(testCase.config),
			)

			assert.Equal(t, testCase.expectsError, diagnostics.HasError(), diagnostics)
		})
	}
}

const (
	testAccComputeVolumeLookupV3ByIDName   = "data.selectel_compute_volume_v3.by_id"
	testAccComputeVolumeLookupV3ByNameName = "data.selectel_compute_volume_v3.by_name"
)

func TestAccSelectelComputeVolumeLookupV3(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-lookup")
	config := testAccComputeVolumeLookupV3Config(name)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccComputeVolumeV3PreCheck(
				t,
				testAccComputeVolumeV3AvailabilityZoneEnv,
				testAccComputeVolumeV3VolumeTypeEnv,
			)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckComputeVolumeV3Destroy(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3UniqueName(t, testAccComputeVolumeV3ResourceName),
					testAccCheckComputeVolumeLookupV3IDsMatch(testAccComputeVolumeLookupV3ByIDName),
					testAccCheckComputeVolumeLookupV3IDsMatch(testAccComputeVolumeLookupV3ByNameName),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"volume_id",
						testAccComputeVolumeV3ResourceName,
						"id",
					),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeLookupV3ByNameName,
						"name",
						name,
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"project_id",
						testAccComputeVolumeV3ResourceName,
						"project_id",
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"region",
						testAccComputeVolumeV3ResourceName,
						"region",
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"description",
						testAccComputeVolumeV3ResourceName,
						"description",
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"size",
						testAccComputeVolumeV3ResourceName,
						"size",
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"availability_zone",
						testAccComputeVolumeV3ResourceName,
						"availability_zone",
					),
					resource.TestCheckResourceAttrPair(
						testAccComputeVolumeLookupV3ByIDName,
						"volume_type",
						testAccComputeVolumeV3ResourceName,
						"volume_type",
					),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "status", "available"),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "bootable", "false"),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "metadata.purpose", "lookup-acceptance"),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "metadata.lookup_test", "published-attributes"),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "source_vol_id", ""),
					resource.TestCheckResourceAttr(testAccComputeVolumeLookupV3ByIDName, "attachment.#", "0"),
					testAccCheckComputeVolumeLookupV3ReservedMetadataHidden(t, testAccComputeVolumeLookupV3ByIDName),
					testAccCheckComputeVolumeLookupV3ReservedMetadataHidden(t, testAccComputeVolumeLookupV3ByNameName),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testAccComputeVolumeLookupV3Config(name string) string {
	return fmt.Sprintf(`
resource "selectel_compute_volume_v3" "test" {
  project_id        = %q
  region            = %q
  availability_zone = %q
  volume_type       = %q
  name              = %q
  description       = "TASK-018 acceptance lookup"
  size              = 1

  metadata = {
    purpose     = "lookup-acceptance"
    lookup_test = "published-attributes"
  }
}

data "selectel_compute_volume_v3" "by_id" {
  project_id = selectel_compute_volume_v3.test.project_id
  region     = selectel_compute_volume_v3.test.region
  volume_id  = selectel_compute_volume_v3.test.id
}

data "selectel_compute_volume_v3" "by_name" {
  project_id = selectel_compute_volume_v3.test.project_id
  region     = selectel_compute_volume_v3.test.region
  name       = selectel_compute_volume_v3.test.name
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv),
		os.Getenv(testAccComputeVolumeV3VolumeTypeEnv),
		name,
	)
}

func testAccCheckComputeVolumeLookupV3IDsMatch(lookupName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		volumeState, ok := state.RootModule().Resources[testAccComputeVolumeV3ResourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume resource %s was not found in state", testAccComputeVolumeV3ResourceName)
		}

		lookupState, ok := state.RootModule().Resources[lookupName]
		if !ok {
			return fmt.Errorf("Block Storage volume data source %s was not found in state", lookupName)
		}
		if lookupState.Primary.ID != volumeState.Primary.ID {
			return fmt.Errorf(
				"Block Storage volume data source ID %s does not match resource ID %s",
				lookupState.Primary.ID,
				volumeState.Primary.ID,
			)
		}

		return nil
	}
}

func testAccCheckComputeVolumeLookupV3ReservedMetadataHidden(
	t *testing.T,
	lookupName string,
) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		volumeState, ok := state.RootModule().Resources[testAccComputeVolumeV3ResourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume resource %s was not found in state", testAccComputeVolumeV3ResourceName)
		}

		lookupState, ok := state.RootModule().Resources[lookupName]
		if !ok {
			return fmt.Errorf("Block Storage volume data source %s was not found in state", lookupName)
		}

		client, err := testAccComputeVolumeV3Client(t, volumeState)
		if err != nil {
			return err
		}
		remoteVolume, _, err := volume.Get(t.Context(), client, volumeState.Primary.ID)
		if err != nil {
			return fmt.Errorf(
				"failed to read Block Storage volume %s while checking reserved metadata: %w",
				volumeState.Primary.ID,
				err,
			)
		}
		if remoteVolume.Metadata[computeVolumeV3ReservedMetadataKey] == "" {
			return fmt.Errorf("Block Storage volume %s does not contain the reserved Create metadata", volumeState.Primary.ID)
		}
		if _, exists := lookupState.Primary.Attributes["metadata."+computeVolumeV3ReservedMetadataKey]; exists {
			return fmt.Errorf("Block Storage volume data source published reserved Create metadata")
		}

		return nil
	}
}

func TestUnitSelectelComputeVolumeLookupV3ReadByIDAndRefresh(t *testing.T) {
	requestCount := 0
	selected := testComputeVolumeLookupV3{
		ID:               testComputeVolumeLookupV3ID,
		Status:           "available",
		Size:             20,
		Name:             "volume-name",
		Description:      "volume-description",
		AvailabilityZone: "ru-2a",
		VolumeType:       "fast.ru-2a",
		Bootable:         "true",
		SnapshotID:       "snapshot-id",
		SourceVolID:      "source-volume-id",
		Metadata: map[string]string{
			"environment":                      "test",
			"server-added":                     "value",
			computeVolumeV3ReservedMetadataKey: "token",
		},
		Attachments: []testComputeVolumeLookupV3Attachment{{
			VolumeID: "lookup-volume-id",
			ServerID: "server-id",
			Device:   "/dev/vdb",
		}},
	}
	server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(
			t,
			testComputeVolumeV3CollectionPath()+"/"+testComputeVolumeLookupV3ID,
			request.URL.Path,
		)
		writeTestComputeVolumeLookupV3(t, response, selected)
	})
	defer server.Close()

	resourceData := testComputeVolumeLookupV3ResourceData(
		t,
		map[string]any{"volume_id": testComputeVolumeLookupV3ID},
	)
	config := testBlockStorageConfig(server.URL)

	for range 2 {
		diagnostics := dataSourceComputeVolumeV3Read(t.Context(), resourceData, config)
		require.False(t, diagnostics.HasError(), diagnostics)
		assertComputeVolumeLookupV3State(t, resourceData, selected)
	}

	selected.Name = "renamed-volume"
	selected.Size = 30
	selected.Metadata = map[string]string{"environment": "updated"}
	selected.Attachments = []testComputeVolumeLookupV3Attachment{{
		VolumeID: testComputeVolumeLookupV3ID,
		ServerID: "updated-server-id",
		Device:   "/dev/vdc",
	}}
	diagnostics := dataSourceComputeVolumeV3Read(t.Context(), resourceData, config)
	require.False(t, diagnostics.HasError(), diagnostics)
	assertComputeVolumeLookupV3State(t, resourceData, selected)
	assert.Equal(t, 3, requestCount)
}

func TestUnitSelectelComputeVolumeLookupV3SearchesEveryPageByExactCriteria(t *testing.T) {
	var requestCount int
	var server *blockStorageTestServer
	selected := testComputeVolumeLookupV3{
		ID:               testComputeVolumeLookupV3ID,
		Status:           "available",
		Size:             10,
		Name:             "target-volume",
		Description:      "selected",
		AvailabilityZone: "ru-2b",
		VolumeType:       "universal.ru-2b",
		Bootable:         "false",
		Metadata: map[string]string{
			"environment":                      "production",
			"empty-value":                      "",
			"server-added":                     "value",
			computeVolumeV3ReservedMetadataKey: "token",
		},
	}

	server = newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, testComputeVolumeV3CollectionPath()+"/detail", request.URL.Path)

		switch request.URL.Query().Get("marker") {
		case "":
			writeTestComputeVolumeLookupV3Page(
				t,
				response,
				[]testComputeVolumeLookupV3{
					{
						ID:       "wrong-status",
						Name:     "target-volume",
						Status:   "in-use",
						Metadata: map[string]string{"environment": "production"},
					},
					{
						ID:       "wrong-metadata",
						Name:     "target-volume",
						Status:   "available",
						Metadata: map[string]string{"environment": "testing"},
					},
					{
						ID:       "missing-empty-valued-key",
						Name:     "target-volume",
						Status:   "available",
						Metadata: map[string]string{"environment": "production"},
					},
				},
				server.URL+testComputeVolumeV3CollectionPath()+"/detail?marker=next",
			)
		case "next":
			writeTestComputeVolumeLookupV3Page(
				t,
				response,
				[]testComputeVolumeLookupV3{selected},
				"",
			)
		default:
			http.NotFound(response, request)
		}
	})
	defer server.Close()

	resourceData, diagnostics := readTestComputeVolumeLookupV3(
		t.Context(),
		t,
		server,
		map[string]any{
			"name":   "target-volume",
			"status": "available",
			"metadata": map[string]any{
				"environment": "production",
				"empty-value": "",
			},
		},
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Equal(t, 2, requestCount)
	assertComputeVolumeLookupV3State(t, resourceData, selected)
}

func TestUnitSelectelComputeVolumeLookupV3SearchCardinality(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		volumes         []testComputeVolumeLookupV3
		expectedMessage string
	}{
		{
			name:            "no matches",
			volumes:         []testComputeVolumeLookupV3{{ID: "other-id", Name: "other-name"}},
			expectedMessage: "no Block Storage volumes matched",
		},
		{
			name: "multiple matches",
			volumes: []testComputeVolumeLookupV3{
				{ID: "first-id", Name: "target-volume"},
				{ID: "second-id", Name: "target-volume"},
			},
			expectedMessage: "found 2 Block Storage volumes",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodGet, request.Method)
				writeTestComputeVolumeLookupV3Page(t, response, testCase.volumes, "")
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeVolumeLookupV3(
				t.Context(),
				t,
				server,
				map[string]any{"name": "target-volume"},
			)

			require.True(t, diagnostics.HasError())
			assert.Empty(t, resourceData.Id())
			assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)
		})
	}
}

func TestUnitSelectelComputeVolumeLookupV3IncompleteList(t *testing.T) {
	var server *blockStorageTestServer
	server = newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("marker") == "" {
			writeTestComputeVolumeLookupV3Page(
				t,
				response,
				nil,
				server.URL+testComputeVolumeV3CollectionPath()+"/detail?marker=next",
			)

			return
		}

		writeTestComputeVolumeV3Fault(t, response, http.StatusInternalServerError, "later page failed")
	})
	defer server.Close()

	resourceData, diagnostics := readTestComputeVolumeLookupV3(
		t.Context(),
		t,
		server,
		map[string]any{"name": "target-volume"},
	)

	require.True(t, diagnostics.HasError())
	assert.Empty(t, resourceData.Id())
	assert.Contains(t, diagnostics[0].Summary, "incomplete_list")
	assert.NotContains(t, diagnostics[0].Summary, "no Block Storage volumes matched")
}

func TestUnitSelectelComputeVolumeLookupV3DirectLookupErrors(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		status          int
		expectedMessage string
		excludedMessage string
	}{
		{
			name:            "not found",
			status:          http.StatusNotFound,
			expectedMessage: "does not exist or is not accessible",
		},
		{
			name:            "forbidden",
			status:          http.StatusForbidden,
			expectedMessage: "forbidden",
			excludedMessage: "does not exist",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, _ *http.Request) {
				writeTestComputeVolumeV3Fault(t, response, testCase.status, testCase.name)
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeVolumeLookupV3(
				t.Context(),
				t,
				server,
				map[string]any{"volume_id": testComputeVolumeLookupV3ID},
			)

			require.True(t, diagnostics.HasError())
			assert.Empty(t, resourceData.Id())
			assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)
			if testCase.excludedMessage != "" {
				assert.NotContains(t, diagnostics[0].Summary, testCase.excludedMessage)
			}
		})
	}
}

func readTestComputeVolumeLookupV3(
	ctx context.Context,
	t *testing.T,
	server *blockStorageTestServer,
	values map[string]any,
) (*schema.ResourceData, diag.Diagnostics) {
	t.Helper()

	resourceData := testComputeVolumeLookupV3ResourceData(t, values)
	diagnostics := dataSourceComputeVolumeV3Read(
		ctx,
		resourceData,
		testBlockStorageConfig(server.URL),
	)

	return resourceData, diagnostics
}

func testComputeVolumeLookupV3ResourceData(
	t *testing.T,
	values map[string]any,
) *schema.ResourceData {
	t.Helper()

	config := map[string]any{
		"project_id": testBlockStorageProjectID,
		"region":     testBlockStorageRegion,
	}
	maps.Copy(config, values)

	return schema.TestResourceDataRaw(t, dataSourceComputeVolumeV3().Schema, config)
}

type testComputeVolumeLookupV3Attachment struct {
	VolumeID string
	ServerID string
	Device   string
}

type testComputeVolumeLookupV3 struct {
	ID               string
	Status           string
	Size             int
	Name             string
	Description      string
	AvailabilityZone string
	VolumeType       string
	Bootable         string
	SnapshotID       string
	SourceVolID      string
	Metadata         map[string]string
	Attachments      []testComputeVolumeLookupV3Attachment
}

func writeTestComputeVolumeLookupV3(
	t *testing.T,
	response http.ResponseWriter,
	selected testComputeVolumeLookupV3,
) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"volume": testComputeVolumeLookupV3Body(selected),
	}))
}

func writeTestComputeVolumeLookupV3Page(
	t *testing.T,
	response http.ResponseWriter,
	volumes []testComputeVolumeLookupV3,
	nextURL string,
) {
	t.Helper()

	items := make([]map[string]any, 0, len(volumes))
	for _, item := range volumes {
		items = append(items, testComputeVolumeLookupV3Body(item))
	}

	body := map[string]any{"volumes": items}
	if nextURL != "" {
		body["volumes_links"] = []map[string]string{{"rel": "next", "href": nextURL}}
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(body))
}

func testComputeVolumeLookupV3Body(selected testComputeVolumeLookupV3) map[string]any {
	attachments := make([]map[string]string, 0, len(selected.Attachments))
	for _, attachment := range selected.Attachments {
		attachments = append(attachments, map[string]string{
			"id":        attachment.VolumeID,
			"server_id": attachment.ServerID,
			"device":    attachment.Device,
		})
	}

	return map[string]any{
		"id":                selected.ID,
		"status":            selected.Status,
		"size":              selected.Size,
		"name":              selected.Name,
		"description":       selected.Description,
		"availability_zone": selected.AvailabilityZone,
		"volume_type":       selected.VolumeType,
		"bootable":          selected.Bootable,
		"snapshot_id":       selected.SnapshotID,
		"source_volid":      selected.SourceVolID,
		"metadata":          selected.Metadata,
		"attachments":       attachments,
	}
}

func assertComputeVolumeLookupV3State(
	t *testing.T,
	resourceData *schema.ResourceData,
	selected testComputeVolumeLookupV3,
) {
	t.Helper()

	assert.Equal(t, selected.ID, resourceData.Id())
	assert.Equal(t, selected.Name, resourceData.Get("name"))
	assert.Equal(t, selected.Description, resourceData.Get("description"))
	assert.Equal(t, selected.Size, resourceData.Get("size"))
	assert.Equal(t, selected.Status, resourceData.Get("status"))
	assert.Equal(t, selected.AvailabilityZone, resourceData.Get("availability_zone"))
	assert.Equal(t, selected.VolumeType, resourceData.Get("volume_type"))
	assert.Equal(t, selected.Bootable, resourceData.Get("bootable"))
	assert.Equal(t, selected.SnapshotID, resourceData.Get("snapshot_id"))
	assert.Equal(t, selected.SourceVolID, resourceData.Get("source_vol_id"))

	expectedMetadata := make(map[string]any, len(selected.Metadata))
	for key, value := range selected.Metadata {
		if key != computeVolumeV3ReservedMetadataKey {
			expectedMetadata[key] = value
		}
	}
	assert.Equal(t, expectedMetadata, resourceData.Get("metadata"))
	assert.NotContains(t, resourceData.Get("metadata"), computeVolumeV3ReservedMetadataKey)

	attachments := resourceData.Get("attachment").(*schema.Set).List()
	require.Len(t, attachments, len(selected.Attachments))
	for i, attachment := range selected.Attachments {
		assert.Equal(t, map[string]any{
			"id":          attachment.VolumeID,
			"instance_id": attachment.ServerID,
			"device":      attachment.Device,
		}, attachments[i])
	}
}
