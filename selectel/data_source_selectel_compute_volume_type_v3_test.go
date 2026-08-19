package selectel

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testVolumeTypeID   = "volume-type-id"
	testVolumeTypeName = "universal.ru-2a"
)

func TestUnitSelectelComputeVolumeTypeV3ConfigurationValidation(t *testing.T) {
	dataSourceSchema := dataSourceComputeVolumeTypeV3().Schema
	testCases := []struct {
		name         string
		config       map[string]any
		expectsError bool
	}{
		{
			name: "missing selector",
			config: map[string]any{
				"project_id": testBlockStorageProjectID,
				"region":     testBlockStorageRegion,
			},
			expectsError: true,
		},
		{
			name: "conflicting selectors",
			config: map[string]any{
				"project_id":     testBlockStorageProjectID,
				"region":         testBlockStorageRegion,
				"volume_type_id": testVolumeTypeID,
				"name":           testVolumeTypeName,
			},
			expectsError: true,
		},
		{
			name: "ID",
			config: map[string]any{
				"project_id":     testBlockStorageProjectID,
				"region":         testBlockStorageRegion,
				"volume_type_id": testVolumeTypeID,
			},
		},
		{
			name: "default",
			config: map[string]any{
				"project_id":     testBlockStorageProjectID,
				"region":         testBlockStorageRegion,
				"volume_type_id": "default",
			},
		},
		{
			name: "name",
			config: map[string]any{
				"project_id": testBlockStorageProjectID,
				"region":     testBlockStorageRegion,
				"name":       testVolumeTypeName,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := schema.InternalMap(dataSourceSchema).Validate(
				terraform.NewResourceConfigRaw(testCase.config),
			)

			assert.Equal(t, testCase.expectsError, diagnostics.HasError(), diagnostics)
		})
	}
}

func TestUnitSelectelComputeVolumeTypeV3ReadByIDAndDefault(t *testing.T) {
	testCases := []struct {
		name               string
		lookupID           string
		resolvedID         string
		supportsCustomIOPS bool
	}{
		{name: "ID", lookupID: testVolumeTypeID, resolvedID: testVolumeTypeID, supportsCustomIOPS: true},
		{name: "default", lookupID: "default", resolvedID: "default-volume-type-id"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodGet, request.Method)

				switch request.URL.Path {
				case testVolumeTypesPath() + "/" + testCase.lookupID:
					writeTestVolumeType(
						t,
						response,
						testCase.resolvedID,
						testVolumeTypeName,
						"Universal SSD",
						true,
						map[string]string{
							"RESKEY:availability_zones": "ru-2a",
							"multiattach":               "<is> False",
						},
					)
				case testQoSLimitsPath():
					qosTypeID := "other-volume-type-id"
					if testCase.supportsCustomIOPS {
						qosTypeID = testCase.resolvedID
					}
					writeTestQoSLimits(t, response, map[string]testQoSLimits{
						"a-name-that-does-not-match-the-selected-type": {
							VolumeTypeID:    qosTypeID,
							AllowUserQoS:    true,
							FullQoSDiskType: true,
						},
					})
				default:
					http.NotFound(response, request)
				}
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeVolumeTypeV3(
				t.Context(),
				t,
				server,
				map[string]any{"volume_type_id": testCase.lookupID},
			)

			require.False(t, diagnostics.HasError(), diagnostics)
			assert.Equal(t, testCase.resolvedID, resourceData.Id())
			assert.Equal(t, testVolumeTypeName, resourceData.Get("name"))
			assert.Equal(t, "Universal SSD", resourceData.Get("description"))
			assert.Equal(t, true, resourceData.Get("is_public"))
			assert.Equal(t, testCase.supportsCustomIOPS, resourceData.Get("supports_custom_iops"))
			assert.Equal(t, map[string]any{
				"RESKEY:availability_zones": "ru-2a",
				"multiattach":               "<is> False",
			}, resourceData.Get("extra_specs"))
			assert.NotContains(t, resourceData.Get("extra_specs"), "volume_backend_name")
		})
	}
}

func TestUnitSelectelComputeVolumeTypeV3ReadsEveryPageByExactName(t *testing.T) {
	var requestCount int
	var server *blockStorageTestServer

	server = newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		assert.Equal(t, http.MethodGet, request.Method)
		if request.URL.Path == testQoSLimitsPath() {
			writeTestQoSLimits(t, response, map[string]testQoSLimits{
				"renamed-flexible-type": {
					VolumeTypeID:    testVolumeTypeID,
					AllowUserQoS:    true,
					FullQoSDiskType: true,
				},
			})

			return
		}

		assert.Equal(t, testVolumeTypesPath(), request.URL.Path)

		switch request.URL.Query().Get("marker") {
		case "":
			writeTestVolumeTypePage(t, response, []testVolumeType{
				{ID: "other-id", Name: "Universal.ru-2a"},
			}, server.URL+testVolumeTypesPath()+"?marker=other-id")
		case "other-id":
			writeTestVolumeTypePage(t, response, []testVolumeType{
				{
					ID:          testVolumeTypeID,
					Name:        testVolumeTypeName,
					Description: "Universal SSD",
					IsPublic:    true,
					ExtraSpecs:  map[string]string{"RESKEY:availability_zones": "ru-2a"},
				},
			}, "")
		default:
			http.NotFound(response, request)
		}
	})
	defer server.Close()

	resourceData, diagnostics := readTestComputeVolumeTypeV3(
		t.Context(),
		t,
		server,
		map[string]any{"name": testVolumeTypeName},
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Equal(t, 3, requestCount)
	assert.Equal(t, testVolumeTypeID, resourceData.Id())
	assert.Equal(t, testVolumeTypeName, resourceData.Get("name"))
	assert.Equal(t, "Universal SSD", resourceData.Get("description"))
	assert.Equal(t, map[string]any{
		"RESKEY:availability_zones": "ru-2a",
	}, resourceData.Get("extra_specs"))
	assert.Equal(t, true, resourceData.Get("supports_custom_iops"))
}

func TestUnitSelectelComputeVolumeTypeV3ReportsQoSCapabilityError(t *testing.T) {
	server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case testVolumeTypesPath() + "/" + testVolumeTypeID:
			writeTestVolumeType(t, response, testVolumeTypeID, testVolumeTypeName, "Universal SSD", true, nil)
		case testQoSLimitsPath():
			writeTestVolumeTypeFault(t, response, http.StatusInternalServerError, "qos endpoint failed")
		default:
			http.NotFound(response, request)
		}
	})
	defer server.Close()

	resourceData, diagnostics := readTestComputeVolumeTypeV3(
		t.Context(),
		t,
		server,
		map[string]any{"volume_type_id": testVolumeTypeID},
	)

	require.True(t, diagnostics.HasError())
	assert.Empty(t, resourceData.Id())
	assert.Contains(t, diagnostics[0].Summary, "server_error")
	assert.Contains(t, diagnostics[0].Summary, "QoS capabilities")
}

func TestUnitSelectelComputeVolumeTypeV3SupportsCustomIOPS(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		limits   []volumetype.QoSLimitsView
		expected bool
	}{
		{
			name: "both flags",
			limits: []volumetype.QoSLimitsView{{
				VolumeTypeID: testVolumeTypeID, AllowUserQoS: true, FullQoSDiskType: true,
			}},
			expected: true,
		},
		{
			name: "missing user flag",
			limits: []volumetype.QoSLimitsView{{
				VolumeTypeID: testVolumeTypeID, FullQoSDiskType: true,
			}},
		},
		{
			name: "missing disk type flag",
			limits: []volumetype.QoSLimitsView{{
				VolumeTypeID: testVolumeTypeID, AllowUserQoS: true,
			}},
		},
		{
			name: "different UUID",
			limits: []volumetype.QoSLimitsView{{
				VolumeTypeID: "another-id", AllowUserQoS: true, FullQoSDiskType: true,
			}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, supportsCustomIOPS(testVolumeTypeID, testCase.limits))
		})
	}
}

func TestUnitSelectelComputeVolumeTypeV3NameCardinality(t *testing.T) {
	testCases := []struct {
		name            string
		volumeTypes     []testVolumeType
		expectedMessage []string
	}{
		{
			name:            "zero including regional name",
			expectedMessage: []string{"was not found", "unknown", "regional", "availability zone"},
		},
		{
			name: "multiple",
			volumeTypes: []testVolumeType{
				{ID: "first-id", Name: testVolumeTypeName},
				{ID: "second-id", Name: testVolumeTypeName},
			},
			expectedMessage: []string{"found 2", "volume_type_id", "unambiguous"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, _ *http.Request) {
				writeTestVolumeTypePage(t, response, testCase.volumeTypes, "")
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeVolumeTypeV3(
				t.Context(),
				t,
				server,
				map[string]any{"name": testVolumeTypeName},
			)

			require.True(t, diagnostics.HasError())
			assert.Empty(t, resourceData.Id())
			for _, message := range testCase.expectedMessage {
				assert.Contains(t, diagnostics[0].Summary, message)
			}
		})
	}
}

func TestUnitSelectelComputeVolumeTypeV3ReportsIncompleteList(t *testing.T) {
	var server *blockStorageTestServer

	server = newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("marker") == "" {
			writeTestVolumeTypePage(t, response, []testVolumeType{
				{ID: "first-id", Name: "other"},
			}, server.URL+testVolumeTypesPath()+"?marker=first-id")

			return
		}

		writeTestVolumeTypeFault(t, response, http.StatusInternalServerError, "server failed")
	})
	defer server.Close()

	resourceData, diagnostics := readTestComputeVolumeTypeV3(
		t.Context(),
		t,
		server,
		map[string]any{"name": testVolumeTypeName},
	)

	require.True(t, diagnostics.HasError())
	assert.Empty(t, resourceData.Id())
	assert.Contains(t, diagnostics[0].Summary, "incomplete_list")
	assert.NotContains(t, diagnostics[0].Summary, "was not found")
}

func readTestComputeVolumeTypeV3(
	ctx context.Context,
	t *testing.T,
	server *blockStorageTestServer,
	values map[string]any,
) (*schema.ResourceData, diag.Diagnostics) {
	t.Helper()

	resourceData := testComputeVolumeTypeV3ResourceData(t, values)
	diagnostics := dataSourceComputeVolumeTypeV3Read(ctx, resourceData, testBlockStorageConfig(server.URL))

	return resourceData, diagnostics
}

func testComputeVolumeTypeV3ResourceData(
	t *testing.T,
	values map[string]any,
) *schema.ResourceData {
	t.Helper()

	config := map[string]any{
		"project_id": testBlockStorageProjectID,
		"region":     testBlockStorageRegion,
	}
	maps.Copy(config, values)

	return schema.TestResourceDataRaw(t, dataSourceComputeVolumeTypeV3().Schema, config)
}

func testVolumeTypesPath() string {
	return "/volumev3/" + testBlockStorageRegion + "/" + testBlockStorageProjectID + "/types"
}

func testQoSLimitsPath() string {
	return "/volumev3/" + testBlockStorageRegion + "/" + testBlockStorageProjectID + "/qos-specs/qos_limits"
}

type testVolumeType struct {
	ID          string
	Name        string
	Description string
	IsPublic    bool
	ExtraSpecs  map[string]string
}

type testQoSLimits struct {
	VolumeTypeID    string
	AllowUserQoS    bool
	FullQoSDiskType bool
}

func writeTestQoSLimits(t *testing.T, response http.ResponseWriter, entries map[string]testQoSLimits) {
	t.Helper()

	body := map[string]any{
		"cfg_timeout":         60,
		"region_volume_types": []string{"universal2", "fast2"},
	}
	for name, entry := range entries {
		body[name] = map[string]any{
			"volume_type_id":     entry.VolumeTypeID,
			"qos_specs":          map[string]int{"total_iops_sec_min": 1000},
			"allow_user_qos":     entry.AllowUserQoS,
			"full_qos_disk_type": entry.FullQoSDiskType,
		}
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(body))
}

func writeTestVolumeType(
	t *testing.T,
	response http.ResponseWriter,
	id string,
	name string,
	description string,
	isPublic bool,
	extraSpecs map[string]string,
) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"volume_type": map[string]any{
			"id":          id,
			"name":        name,
			"description": description,
			"is_public":   isPublic,
			"extra_specs": extraSpecs,
		},
	}))
}

const (
	testAccComputeVolumeTypeV3ByIDName    = "data.selectel_compute_volume_type_v3.by_id"
	testAccComputeVolumeTypeV3DefaultName = "data.selectel_compute_volume_type_v3.default"
	testAccComputeVolumeTypeV3ByExactName = "data.selectel_compute_volume_type_v3.by_name"
)

type testAccComputeVolumeTypeV3Fixture struct {
	ID                 string
	Name               string
	Description        string
	IsPublic           bool
	ExtraSpecs         map[string]string
	SupportsCustomIOPS bool
}

func TestAccSelectelComputeVolumeTypeV3(t *testing.T) {
	testAccComputeVolumeTypeV3PreCheck(t)
	fixture := testAccDiscoverComputeVolumeTypeV3Fixture(t)
	defaultFixture := testAccReadDefaultComputeVolumeTypeV3Fixture(t)
	config := testAccComputeVolumeTypeV3Config(fixture)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccComputeVolumeTypeV3PreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeTypeV3Fixture(testAccComputeVolumeTypeV3ByIDName, fixture),
					resource.TestCheckResourceAttr(testAccComputeVolumeTypeV3ByIDName, "volume_type_id", fixture.ID),
					testAccCheckComputeVolumeTypeV3Fixture(testAccComputeVolumeTypeV3DefaultName, defaultFixture),
					resource.TestCheckResourceAttr(testAccComputeVolumeTypeV3DefaultName, "volume_type_id", volumetype.DefaultTypeID),
					testAccCheckComputeVolumeTypeV3Fixture(testAccComputeVolumeTypeV3ByExactName, fixture),
					resource.TestCheckResourceAttr(testAccComputeVolumeTypeV3ByExactName, "name", fixture.Name),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testAccComputeVolumeTypeV3PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for Block Storage volume type acceptance tests")
	}

	testAccSelectelPreCheckWithAuth(t)

	for _, environmentVariable := range []string{"INFRA_PROJECT_ID", "INFRA_REGION"} {
		if os.Getenv(environmentVariable) == "" {
			t.Fatalf("%s must be set for Block Storage volume type acceptance tests", environmentVariable)
		}
	}
}

func testAccDiscoverComputeVolumeTypeV3Fixture(t *testing.T) testAccComputeVolumeTypeV3Fixture {
	t.Helper()

	client := testAccComputeVolumeTypeV3Client(t)
	volumeTypes, err := volumetype.List(t.Context(), client, volumetype.ListOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, volumeTypes, "the Block Storage endpoint returned no visible volume types")
	capabilities := testAccComputeVolumeTypeV3Capabilities(t, client)

	nameCounts := make(map[string]int, len(volumeTypes))
	for i := range volumeTypes {
		nameCounts[volumeTypes[i].Name]++
	}

	sort.Slice(volumeTypes, func(i, j int) bool {
		if volumeTypes[i].Name == volumeTypes[j].Name {
			return volumeTypes[i].ID < volumeTypes[j].ID
		}

		return volumeTypes[i].Name < volumeTypes[j].Name
	})

	for i := range volumeTypes {
		if nameCounts[volumeTypes[i].Name] == 1 && capabilities[volumeTypes[i].ID] {
			return testAccComputeVolumeTypeV3FixtureFromView(&volumeTypes[i], true)
		}
	}

	defaultVolumeType, _, err := volumetype.Get(t.Context(), client, volumetype.DefaultTypeID)
	require.NoError(t, err)
	if nameCounts[defaultVolumeType.Name] == 1 {
		return testAccComputeVolumeTypeV3FixtureFromView(
			defaultVolumeType,
			capabilities[defaultVolumeType.ID],
		)
	}

	for i := range volumeTypes {
		if nameCounts[volumeTypes[i].Name] == 1 {
			return testAccComputeVolumeTypeV3FixtureFromView(
				&volumeTypes[i],
				capabilities[volumeTypes[i].ID],
			)
		}
	}

	t.Fatal("the Block Storage endpoint returned no volume type with a unique exact name")

	return testAccComputeVolumeTypeV3Fixture{}
}

func testAccReadDefaultComputeVolumeTypeV3Fixture(t *testing.T) testAccComputeVolumeTypeV3Fixture {
	t.Helper()

	client := testAccComputeVolumeTypeV3Client(t)
	defaultVolumeType, _, err := volumetype.Get(t.Context(), client, volumetype.DefaultTypeID)
	require.NoError(t, err)

	capabilities := testAccComputeVolumeTypeV3Capabilities(t, client)

	return testAccComputeVolumeTypeV3FixtureFromView(
		defaultVolumeType,
		capabilities[defaultVolumeType.ID],
	)
}

func testAccComputeVolumeTypeV3Client(t *testing.T) *blockstorage.Client {
	t.Helper()

	resourceData := schema.TestResourceDataRaw(t, dataSourceComputeVolumeTypeV3().Schema, map[string]any{
		"project_id":     os.Getenv("INFRA_PROJECT_ID"),
		"region":         os.Getenv("INFRA_REGION"),
		"volume_type_id": volumetype.DefaultTypeID,
	})
	config := &Config{
		AuthURL:        os.Getenv("OS_AUTH_URL"),
		AuthRegion:     os.Getenv("OS_REGION_NAME"),
		DomainName:     os.Getenv("OS_DOMAIN_NAME"),
		UserDomainName: os.Getenv("OS_USER_DOMAIN_NAME"),
		Username:       os.Getenv("OS_USERNAME"),
		Password:       os.Getenv("OS_PASSWORD"),
		UserAgent:      "terraform-provider-selectel/acceptance-tests",
	}

	client, diagnostics := getBlockStorageClient(resourceData, config)
	require.False(t, diagnostics.HasError(), diagnostics)

	return client
}

func testAccComputeVolumeTypeV3Capabilities(
	t *testing.T,
	client *blockstorage.Client,
) map[string]bool {
	t.Helper()

	limits, _, err := volumetype.ListQoSLimits(t.Context(), client)
	require.NoError(t, err)

	capabilities := make(map[string]bool, len(limits))
	for i := range limits {
		if limits[i].AllowUserQoS && limits[i].FullQoSDiskType {
			capabilities[limits[i].VolumeTypeID] = true
		}
	}

	return capabilities
}

func testAccComputeVolumeTypeV3FixtureFromView(
	volumeType *volumetype.View,
	supportsCustomIOPS bool,
) testAccComputeVolumeTypeV3Fixture {
	return testAccComputeVolumeTypeV3Fixture{
		ID:                 volumeType.ID,
		Name:               volumeType.Name,
		Description:        volumeType.Description,
		IsPublic:           volumeType.IsPublic,
		ExtraSpecs:         volumeType.ExtraSpecs,
		SupportsCustomIOPS: supportsCustomIOPS,
	}
}

func testAccComputeVolumeTypeV3Config(fixture testAccComputeVolumeTypeV3Fixture) string {
	return fmt.Sprintf(`
data "selectel_compute_volume_type_v3" "by_id" {
  project_id     = %q
  region         = %q
  volume_type_id = %q
}

data "selectel_compute_volume_type_v3" "default" {
  project_id     = %q
  region         = %q
  volume_type_id = "default"
}

data "selectel_compute_volume_type_v3" "by_name" {
  project_id = %q
  region     = %q
  name       = %q
}
`,
		os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), fixture.ID,
		os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"),
		os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), fixture.Name,
	)
}

func testAccCheckComputeVolumeTypeV3Fixture(
	resourceName string,
	fixture testAccComputeVolumeTypeV3Fixture,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume type data source %s was not found in state", resourceName)
		}
		if resourceState.Primary.ID != fixture.ID {
			return fmt.Errorf(
				"expected Block Storage volume type UUID %q, got %q",
				fixture.ID,
				resourceState.Primary.ID,
			)
		}

		expectedAttributes := map[string]string{
			"name":                 fixture.Name,
			"description":          fixture.Description,
			"is_public":            strconv.FormatBool(fixture.IsPublic),
			"extra_specs.%":        strconv.Itoa(len(fixture.ExtraSpecs)),
			"supports_custom_iops": strconv.FormatBool(fixture.SupportsCustomIOPS),
		}
		for key, value := range fixture.ExtraSpecs {
			expectedAttributes["extra_specs."+key] = value
		}

		for attribute, expected := range expectedAttributes {
			if actual := resourceState.Primary.Attributes[attribute]; actual != expected {
				return fmt.Errorf(
					"expected %s to be %q for Block Storage volume type %q, got %q",
					attribute,
					expected,
					fixture.ID,
					actual,
				)
			}
		}

		return nil
	}
}

func writeTestVolumeTypePage(
	t *testing.T,
	response http.ResponseWriter,
	volumeTypes []testVolumeType,
	nextURL string,
) {
	t.Helper()

	items := make([]map[string]any, 0, len(volumeTypes))
	for _, volumeType := range volumeTypes {
		items = append(items, map[string]any{
			"id":          volumeType.ID,
			"name":        volumeType.Name,
			"description": volumeType.Description,
			"is_public":   volumeType.IsPublic,
			"extra_specs": volumeType.ExtraSpecs,
		})
	}

	body := map[string]any{"volume_types": items}
	if nextURL != "" {
		body["volume_type_links"] = []map[string]string{{"rel": "next", "href": nextURL}}
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(body))
}

func writeTestVolumeTypeFault(t *testing.T, response http.ResponseWriter, status int, message string) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"computeFault": map[string]any{
			"code":    status,
			"message": message,
		},
	}))
}
