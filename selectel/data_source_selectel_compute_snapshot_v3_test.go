package selectel

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud"
	gophersnapshots "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSelectelComputeSnapshotV3Schema(t *testing.T) {
	dataSource := dataSourceComputeSnapshotV3()
	dataSourceSchema := dataSource.Schema

	for _, attribute := range []string{"project_id", "region"} {
		assert.True(t, dataSourceSchema[attribute].Required, attribute)
		assert.False(t, dataSourceSchema[attribute].Optional, attribute)
		assert.False(t, dataSourceSchema[attribute].Computed, attribute)
	}
	for _, attribute := range []string{"name", "status", "volume_id"} {
		assert.True(t, dataSourceSchema[attribute].Optional, attribute)
		assert.True(t, dataSourceSchema[attribute].Computed, attribute)
	}
	for _, attribute := range []string{"description", "size", "metadata"} {
		assert.True(t, dataSourceSchema[attribute].Computed, attribute)
		assert.False(t, dataSourceSchema[attribute].Optional, attribute)
	}
	assert.True(t, dataSourceSchema["most_recent"].Optional)
	assert.False(t, dataSourceSchema["most_recent"].Computed)
	assert.Nil(t, dataSource.CreateContext)
	assert.Nil(t, dataSource.UpdateContext)
	assert.Nil(t, dataSource.DeleteContext)
	assert.Nil(t, dataSource.Importer)
	assert.NotNil(t, Provider("test").DataSourcesMap["selectel_compute_snapshot_v3"])

	valid := map[string]any{
		"project_id": testBlockStorageProjectID,
		"region":     testBlockStorageRegion,
	}
	for _, testCase := range []struct {
		name   string
		values map[string]any
	}{
		{name: "missing project", values: map[string]any{"region": testBlockStorageRegion}},
		{name: "missing region", values: map[string]any{"project_id": testBlockStorageProjectID}},
		{name: "blank project", values: map[string]any{"project_id": " ", "region": testBlockStorageRegion}},
		{name: "blank region", values: map[string]any{"project_id": testBlockStorageProjectID, "region": "\t"}},
		{name: "blank name", values: map[string]any{"project_id": testBlockStorageProjectID, "region": testBlockStorageRegion, "name": " "}},
		{name: "blank status", values: map[string]any{"project_id": testBlockStorageProjectID, "region": testBlockStorageRegion, "status": " "}},
		{name: "blank volume ID", values: map[string]any{"project_id": testBlockStorageProjectID, "region": testBlockStorageRegion, "volume_id": " "}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := schema.InternalMap(dataSourceSchema).Validate(
				terraform.NewResourceConfigRaw(testCase.values),
			)
			assert.True(t, diagnostics.HasError(), diagnostics)
		})
	}

	diagnostics := schema.InternalMap(dataSourceSchema).Validate(terraform.NewResourceConfigRaw(valid))
	assert.False(t, diagnostics.HasError(), diagnostics)
}

func TestUnitSelectelComputeSnapshotV3Filters(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		values map[string]any
		check  func(*testing.T, *http.Request)
	}{
		{
			name: "omits empty selectors",
			check: func(t *testing.T, request *http.Request) {
				assert.Empty(t, request.URL.RawQuery)
			},
		},
		{
			name: "combines selectors",
			values: map[string]any{
				"name":        "daily",
				"status":      "available",
				"volume_id":   "volume-id",
				"most_recent": true,
			},
			check: func(t *testing.T, request *http.Request) {
				query := request.URL.Query()
				assert.Len(t, query, 3)
				assert.Equal(t, "daily", query.Get("name"))
				assert.Equal(t, "available", query.Get("status"))
				assert.Equal(t, "volume-id", query.Get("volume_id"))
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			selected := testComputeSnapshotV3{
				ID: "snapshot-id", Name: "daily", Status: "available", VolumeID: "volume-id",
			}
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodGet, request.Method)
				assert.Equal(t, testComputeSnapshotV3CollectionPath(), request.URL.Path)
				testCase.check(t, request)
				writeTestComputeSnapshotV3Page(t, response, []testComputeSnapshotV3{selected}, "")
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeSnapshotV3(t, server, testCase.values)

			require.False(t, diagnostics.HasError(), diagnostics)
			assertComputeSnapshotV3State(t, resourceData, selected)
		})
	}
}

func TestUnitSelectelComputeSnapshotV3RefreshesState(t *testing.T) {
	requestCount := 0
	selected := testComputeSnapshotV3{
		ID:          "snapshot-id",
		CreatedAt:   "2026-08-20T10:00:00.000000",
		Name:        "daily",
		Description: "initial",
		VolumeID:    "volume-id",
		Status:      "available",
		Size:        10,
		Metadata:    map[string]string{"environment": "test"},
	}
	server := newBlockStorageTestServer(t, func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		writeTestComputeSnapshotV3Page(t, response, []testComputeSnapshotV3{selected}, "")
	})
	defer server.Close()

	resourceData := testComputeSnapshotV3ResourceData(t, map[string]any{"name": "daily"})
	config := testBlockStorageConfig(server.URL)
	for range 2 {
		diagnostics := dataSourceComputeSnapshotV3Read(t.Context(), resourceData, config)
		require.False(t, diagnostics.HasError(), diagnostics)
		assertComputeSnapshotV3State(t, resourceData, selected)
	}

	selected.Description = "updated"
	selected.Status = "in-use"
	selected.Size = 20
	selected.Metadata = map[string]string{"environment": "production"}
	diagnostics := dataSourceComputeSnapshotV3Read(t.Context(), resourceData, config)
	require.False(t, diagnostics.HasError(), diagnostics)
	assertComputeSnapshotV3State(t, resourceData, selected)
	assert.Equal(t, 3, requestCount)
}

func TestUnitSelectelComputeSnapshotV3Cardinality(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		snapshots       []testComputeSnapshotV3
		values          map[string]any
		expectedIDs     []string
		expectedMessage string
	}{
		{
			name:            "zero results",
			expectedMessage: "no Block Storage snapshots matched",
		},
		{
			name: "ambiguous without most recent",
			snapshots: []testComputeSnapshotV3{
				{ID: "first", CreatedAt: "2026-08-20T10:00:00.000000"},
				{ID: "second", CreatedAt: "2026-08-21T10:00:00.000000"},
			},
			expectedMessage: "found 2 Block Storage snapshots",
		},
		{
			name: "unique newest",
			snapshots: []testComputeSnapshotV3{
				{ID: "older", CreatedAt: "2026-08-20T10:00:00.000000"},
				{ID: "newest", CreatedAt: "2026-08-22T10:00:00.000000"},
				{ID: "middle", CreatedAt: "2026-08-21T10:00:00.000000"},
			},
			values:      map[string]any{"most_recent": true},
			expectedIDs: []string{"newest"},
		},
		{
			name: "equal newest timestamps",
			snapshots: []testComputeSnapshotV3{
				{ID: "older", CreatedAt: "2026-08-20T10:00:00.000000"},
				{ID: "newest-a", CreatedAt: "2026-08-22T10:00:00.000000"},
				{ID: "newest-b", CreatedAt: "2026-08-22T10:00:00.000000"},
			},
			values:      map[string]any{"most_recent": true},
			expectedIDs: []string{"newest-a", "newest-b"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, func(response http.ResponseWriter, _ *http.Request) {
				writeTestComputeSnapshotV3Page(t, response, testCase.snapshots, "")
			})
			defer server.Close()

			resourceData, diagnostics := readTestComputeSnapshotV3(t, server, testCase.values)

			if testCase.expectedMessage != "" {
				require.True(t, diagnostics.HasError())
				assert.Empty(t, resourceData.Id())
				assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)

				return
			}

			require.False(t, diagnostics.HasError(), diagnostics)
			assert.Contains(t, testCase.expectedIDs, resourceData.Id())
			for _, candidate := range testCase.snapshots {
				if candidate.ID == resourceData.Id() {
					assertComputeSnapshotV3State(t, resourceData, candidate)
				}
			}
		})
	}
}

func TestUnitSelectelComputeSnapshotV3ReadErrors(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			http.Error(response, "authentication failed", http.StatusUnauthorized)
		}))
		defer server.Close()

		resourceData := testComputeSnapshotV3ResourceData(t, nil)
		diagnostics := dataSourceComputeSnapshotV3Read(
			t.Context(), resourceData, testBlockStorageConfig(server.URL),
		)

		require.True(t, diagnostics.HasError())
		assert.Empty(t, resourceData.Id())
		assert.Contains(t, diagnostics[0].Summary, "failed to get project-scoped")
	})

	for _, testCase := range []struct {
		name            string
		handler         http.HandlerFunc
		expectedMessage string
	}{
		{
			name: "HTTP",
			handler: func(response http.ResponseWriter, _ *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(http.StatusForbidden)
				_, _ = response.Write([]byte(`{"forbidden":{"code":403,"message":"denied"}}`))
			},
			expectedMessage: "forbidden",
		},
		{
			name: "decode",
			handler: func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusOK)
				_, _ = response.Write([]byte(`{`))
			},
			expectedMessage: "unexpected",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := newBlockStorageTestServer(t, testCase.handler)
			defer server.Close()

			resourceData, diagnostics := readTestComputeSnapshotV3(t, server, nil)

			require.True(t, diagnostics.HasError())
			assert.Empty(t, resourceData.Id())
			assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)
		})
	}

	t.Run("incomplete pagination", func(t *testing.T) {
		var server *blockStorageTestServer
		server = newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Query().Get("marker") == "" {
				writeTestComputeSnapshotV3Page(
					t,
					response,
					[]testComputeSnapshotV3{{ID: "partial"}},
					server.URL+testComputeSnapshotV3CollectionPath()+"?marker=next",
				)

				return
			}

			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(`{"error":{"message":"later page failed"}}`))
		})
		defer server.Close()

		resourceData, diagnostics := readTestComputeSnapshotV3(t, server, nil)

		require.True(t, diagnostics.HasError())
		assert.Empty(t, resourceData.Id())
		assert.Contains(t, diagnostics[0].Summary, "incomplete_list")
		assert.NotContains(t, diagnostics[0].Summary, "no Block Storage snapshots matched")
	})

	t.Run("state set", func(t *testing.T) {
		selected := testComputeSnapshotV3{ID: "snapshot-id", Metadata: map[string]string{"key": "value"}}
		server := newBlockStorageTestServer(t, func(response http.ResponseWriter, _ *http.Request) {
			writeTestComputeSnapshotV3Page(t, response, []testComputeSnapshotV3{selected}, "")
		})
		defer server.Close()

		client, clientDiagnostics := getBlockStorageClient(
			testComputeSnapshotV3ResourceData(t, nil),
			testBlockStorageConfig(server.URL),
		)
		require.False(t, clientDiagnostics.HasError(), clientDiagnostics)

		brokenSchema := dataSourceComputeSnapshotV3().Schema
		brokenSchema["metadata"] = &schema.Schema{Type: schema.TypeInt, Computed: true}
		resourceData := schema.TestResourceDataRaw(t, brokenSchema, map[string]any{
			"project_id": testBlockStorageProjectID,
			"region":     testBlockStorageRegion,
		})
		diagnostics := readComputeSnapshotV3(t.Context(), resourceData, client)

		require.True(t, diagnostics.HasError())
		assert.Empty(t, resourceData.Id())
		assert.Contains(t, diagnostics[0].Summary, "failed to set Block Storage snapshot metadata")
	})
}

const (
	testAccComputeSnapshotV3ExactName      = "data.selectel_compute_snapshot_v3.exact"
	testAccComputeSnapshotV3MostRecentName = "data.selectel_compute_snapshot_v3.most_recent"
	testAccComputeSnapshotV3Timeout        = 600
)

type testAccComputeSnapshotV3Fixture struct {
	Client    *gophercloud.ServiceClient
	Prefix    string
	VolumeID  string
	Snapshots []gophersnapshots.Snapshot
	Cleanup   func() error
}

func TestAccSelectelComputeSnapshotV3(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for Block Storage snapshot acceptance tests")
	}
	testAccComputeVolumeV3PreCheck(t)

	fixture := testAccCreateComputeSnapshotV3Fixture(
		t,
		acctest.RandomWithPrefix("tf-acc-selectel-snapshot"),
		volumes.CreateOpts{Size: 1},
		2,
	)
	exact := fixture.Snapshots[0]
	exactChecks := make([]resource.TestCheckFunc, 0, 7+len(exact.Metadata))
	exactChecks = append(exactChecks,
		testAccCheckComputeSnapshotV3ID(testAccComputeSnapshotV3ExactName, map[string]struct{}{exact.ID: {}}),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "name", exact.Name),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "description", exact.Description),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "volume_id", exact.VolumeID),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "status", exact.Status),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "size", strconv.Itoa(exact.Size)),
		resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "metadata.%", strconv.Itoa(len(exact.Metadata))),
	)
	for key, value := range exact.Metadata {
		exactChecks = append(
			exactChecks,
			resource.TestCheckResourceAttr(testAccComputeSnapshotV3ExactName, "metadata."+key, value),
		)
	}

	mostRecentConfig := testAccComputeSnapshotV3MostRecentConfig(fixture)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccComputeVolumeV3PreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshotV3ExactConfig(fixture, exact),
				Check:  resource.ComposeTestCheckFunc(exactChecks...),
			},
			{
				Config:      testAccComputeSnapshotV3AmbiguousConfig(fixture),
				ExpectError: regexp.MustCompile(`found 2 Block Storage snapshots`),
			},
			{
				Config: mostRecentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeSnapshotV3ID(
						testAccComputeSnapshotV3MostRecentName,
						testAccComputeSnapshotV3NewestIDs(fixture.Snapshots),
					),
					resource.TestCheckResourceAttr(
						testAccComputeSnapshotV3MostRecentName,
						"volume_id",
						fixture.VolumeID,
					),
				),
			},
			{
				Config:   mostRecentConfig,
				PlanOnly: true,
			},
		},
	})

	require.NoError(t, fixture.Cleanup())
	testAccCheckComputeSnapshotV3FixturesAbsent(t, fixture.Client, fixture.Prefix)
}

func testAccCreateComputeSnapshotV3Fixture(
	t *testing.T,
	prefix string,
	volumeOpts volumes.CreateOpts,
	snapshotCount int,
) testAccComputeSnapshotV3Fixture {
	t.Helper()

	client, err := testAccComputeSnapshotV3ClientForScope(os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"))
	require.NoError(t, err)
	cleanups := make([]func() error, 0, snapshotCount+1)
	registerCleanup := func(cleanup func() error) {
		done := false
		run := func() error {
			if done {
				return nil
			}
			if err := cleanup(); err != nil {
				return err
			}
			done = true

			return nil
		}
		cleanups = append(cleanups, run)
		t.Cleanup(func() { assert.NoError(t, run()) })
	}

	volumeOpts.Name = prefix + "-volume"
	volumeOpts.Metadata = map[string]string{"purpose": "snapshot-acceptance"}
	createdVolume, err := volumes.Create(client, volumeOpts).Extract()
	require.NoError(t, err)
	require.NotEmpty(t, createdVolume.ID)
	registerCleanup(func() error {
		return testAccDeleteComputeSnapshotV3Volume(client, createdVolume.ID)
	})
	require.NoError(
		t,
		volumes.WaitForStatus(client, createdVolume.ID, "available", testAccComputeSnapshotV3Timeout),
	)

	fixture := testAccComputeSnapshotV3Fixture{
		Client:   client,
		Prefix:   prefix,
		VolumeID: createdVolume.ID,
	}
	for index := 1; index <= snapshotCount; index++ {
		createdSnapshot, err := gophersnapshots.Create(client, gophersnapshots.CreateOpts{
			VolumeID:    createdVolume.ID,
			Name:        fmt.Sprintf("%s-%d", prefix, index),
			Description: fmt.Sprintf("Terraform acceptance snapshot %d", index),
			Metadata:    map[string]string{"purpose": "snapshot-acceptance", "index": strconv.Itoa(index)},
		}).Extract()
		require.NoError(t, err)
		require.NotEmpty(t, createdSnapshot.ID)
		snapshotID := createdSnapshot.ID
		registerCleanup(func() error {
			return testAccDeleteComputeSnapshotV3Snapshot(client, snapshotID)
		})
		require.NoError(
			t,
			gophersnapshots.WaitForStatus(client, snapshotID, "available", testAccComputeSnapshotV3Timeout),
		)

		current, err := gophersnapshots.Get(client, snapshotID).Extract()
		require.NoError(t, err)
		require.False(t, current.CreatedAt.IsZero())
		fixture.Snapshots = append(fixture.Snapshots, *current)
	}

	fixture.Cleanup = func() error {
		cleanupErrors := make([]error, 0, len(cleanups))
		for index := len(cleanups) - 1; index >= 0; index-- {
			cleanupErrors = append(cleanupErrors, cleanups[index]())
		}

		return errors.Join(cleanupErrors...)
	}

	return fixture
}

func testAccComputeSnapshotV3ClientForScope(projectID, region string) (*gophercloud.ServiceClient, error) {
	config := &Config{
		AuthURL:        os.Getenv("OS_AUTH_URL"),
		AuthRegion:     os.Getenv("OS_REGION_NAME"),
		DomainName:     os.Getenv("OS_DOMAIN_NAME"),
		UserDomainName: os.Getenv("OS_USER_DOMAIN_NAME"),
		Username:       os.Getenv("OS_USERNAME"),
		Password:       os.Getenv("OS_PASSWORD"),
		UserAgent:      "terraform-provider-selectel/acceptance-tests",
	}
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, err
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(BlockStorage, region)
	if err != nil {
		return nil, err
	}
	if endpoint.URL == "" {
		return nil, fmt.Errorf("Block Storage endpoint is empty for region %s", region)
	}

	providerClient := &gophercloud.ProviderClient{TokenID: selvpcClient.GetXAuthToken()}
	providerClient.UserAgent.Prepend(config.UserAgent)

	return &gophercloud.ServiceClient{
		ProviderClient: providerClient,
		Endpoint:       gophercloud.NormalizeURL(endpoint.URL),
		Type:           BlockStorage,
	}, nil
}

func testAccDeleteComputeSnapshotV3Snapshot(
	client *gophercloud.ServiceClient,
	snapshotID string,
) error {
	if err := gophersnapshots.Delete(client, snapshotID).ExtractErr(); err != nil {
		if testAccComputeSnapshotV3NotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete Block Storage snapshot %s: %w", snapshotID, err)
	}

	err := testAccComputeSnapshotV3WaitNotFound(func() error {
		_, err := gophersnapshots.Get(client, snapshotID).Extract()

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to wait for Block Storage snapshot %s deletion: %w", snapshotID, err)
	}

	return nil
}

func testAccDeleteComputeSnapshotV3Volume(
	client *gophercloud.ServiceClient,
	volumeID string,
) error {
	if err := volumes.Delete(client, volumeID, nil).ExtractErr(); err != nil {
		if testAccComputeSnapshotV3NotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete Block Storage volume %s: %w", volumeID, err)
	}

	err := testAccComputeSnapshotV3WaitNotFound(func() error {
		_, err := volumes.Get(client, volumeID).Extract()

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to wait for Block Storage volume %s deletion: %w", volumeID, err)
	}

	return nil
}

func testAccComputeSnapshotV3NotFound(err error) bool {
	var notFound gophercloud.ErrDefault404

	return errors.As(err, &notFound)
}

func testAccComputeSnapshotV3WaitNotFound(get func() error) error {
	return gophercloud.WaitFor(testAccComputeSnapshotV3Timeout, func() (bool, error) {
		err := get()
		if err == nil {
			return false, nil
		}
		if testAccComputeSnapshotV3NotFound(err) {
			return true, nil
		}

		return false, err
	})
}

func testAccComputeSnapshotV3ExactConfig(
	fixture testAccComputeSnapshotV3Fixture,
	snapshot gophersnapshots.Snapshot,
) string {
	return fmt.Sprintf(`
data "selectel_compute_snapshot_v3" "exact" {
  project_id = %q
  region     = %q
  name       = %q
  status     = %q
  volume_id  = %q
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), snapshot.Name, snapshot.Status, fixture.VolumeID)
}

func testAccComputeSnapshotV3AmbiguousConfig(fixture testAccComputeSnapshotV3Fixture) string {
	return fmt.Sprintf(`
data "selectel_compute_snapshot_v3" "ambiguous" {
  project_id = %q
  region     = %q
  volume_id  = %q
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), fixture.VolumeID)
}

func testAccComputeSnapshotV3MostRecentConfig(fixture testAccComputeSnapshotV3Fixture) string {
	return fmt.Sprintf(`
data "selectel_compute_snapshot_v3" "most_recent" {
  project_id  = %q
  region      = %q
  volume_id   = %q
  most_recent = true
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), fixture.VolumeID)
}

func testAccComputeSnapshotV3NewestIDs(snapshots []gophersnapshots.Snapshot) map[string]struct{} {
	newest := snapshots[0].CreatedAt
	result := map[string]struct{}{snapshots[0].ID: {}}
	for _, candidate := range snapshots[1:] {
		switch {
		case candidate.CreatedAt.After(newest):
			newest = candidate.CreatedAt
			result = map[string]struct{}{candidate.ID: {}}
		case candidate.CreatedAt.Equal(newest):
			result[candidate.ID] = struct{}{}
		}
	}

	return result
}

func testAccCheckComputeSnapshotV3ID(
	resourceName string,
	expectedIDs map[string]struct{},
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Block Storage snapshot data source %s was not found in state", resourceName)
		}
		if _, ok := expectedIDs[resourceState.Primary.ID]; !ok {
			return fmt.Errorf(
				"unexpected Block Storage snapshot ID %q, expected one of %v",
				resourceState.Primary.ID,
				expectedIDs,
			)
		}

		return nil
	}
}

func testAccCheckComputeSnapshotV3FixturesAbsent(
	t *testing.T,
	client *gophercloud.ServiceClient,
	prefix string,
) {
	t.Helper()

	snapshotPages, err := gophersnapshots.List(client, nil).AllPages()
	require.NoError(t, err)
	remainingSnapshots, err := gophersnapshots.ExtractSnapshots(snapshotPages)
	require.NoError(t, err)
	for _, candidate := range remainingSnapshots {
		assert.False(t, strings.HasPrefix(candidate.Name, prefix), "snapshot fixture %s remains", candidate.ID)
	}

	volumePages, err := volumes.List(client, nil).AllPages()
	require.NoError(t, err)
	remainingVolumes, err := volumes.ExtractVolumes(volumePages)
	require.NoError(t, err)
	for _, candidate := range remainingVolumes {
		assert.False(t, strings.HasPrefix(candidate.Name, prefix), "volume fixture %s remains", candidate.ID)
	}
}

func readTestComputeSnapshotV3(
	t *testing.T,
	server *blockStorageTestServer,
	values map[string]any,
) (*schema.ResourceData, diag.Diagnostics) {
	t.Helper()

	resourceData := testComputeSnapshotV3ResourceData(t, values)
	diagnostics := dataSourceComputeSnapshotV3Read(
		t.Context(), resourceData, testBlockStorageConfig(server.URL),
	)

	return resourceData, diagnostics
}

func testComputeSnapshotV3ResourceData(t *testing.T, values map[string]any) *schema.ResourceData {
	t.Helper()

	config := map[string]any{
		"project_id": testBlockStorageProjectID,
		"region":     testBlockStorageRegion,
	}
	maps.Copy(config, values)

	return schema.TestResourceDataRaw(t, dataSourceComputeSnapshotV3().Schema, config)
}

type testComputeSnapshotV3 struct {
	ID          string
	CreatedAt   string
	Name        string
	Description string
	VolumeID    string
	Status      string
	Size        int
	Metadata    map[string]string
}

func testComputeSnapshotV3CollectionPath() string {
	return "/volumev3/" + testBlockStorageRegion + "/" + testBlockStorageProjectID + "/snapshots"
}

func writeTestComputeSnapshotV3Page(
	t *testing.T,
	response http.ResponseWriter,
	snapshots []testComputeSnapshotV3,
	nextURL string,
) {
	t.Helper()

	items := make([]map[string]any, 0, len(snapshots))
	for _, item := range snapshots {
		items = append(items, map[string]any{
			"id":          item.ID,
			"created_at":  item.CreatedAt,
			"name":        item.Name,
			"description": item.Description,
			"volume_id":   item.VolumeID,
			"status":      item.Status,
			"size":        item.Size,
			"metadata":    item.Metadata,
		})
	}

	body := map[string]any{"snapshots": items}
	if nextURL != "" {
		body["snapshots_links"] = []map[string]string{{"rel": "next", "href": nextURL}}
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(body))
}

func assertComputeSnapshotV3State(
	t *testing.T,
	resourceData *schema.ResourceData,
	selected testComputeSnapshotV3,
) {
	t.Helper()

	assert.Equal(t, selected.ID, resourceData.Id())
	assert.Equal(t, selected.Name, resourceData.Get("name"))
	assert.Equal(t, selected.Description, resourceData.Get("description"))
	assert.Equal(t, selected.VolumeID, resourceData.Get("volume_id"))
	assert.Equal(t, selected.Status, resourceData.Get("status"))
	assert.Equal(t, selected.Size, resourceData.Get("size"))

	expectedMetadata := make(map[string]any, len(selected.Metadata))
	for key, value := range selected.Metadata {
		expectedMetadata[key] = value
	}
	assert.Equal(t, expectedMetadata, resourceData.Get("metadata"))
}
