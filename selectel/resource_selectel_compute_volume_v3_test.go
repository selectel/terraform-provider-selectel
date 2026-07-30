package selectel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testComputeVolumeV3NewName        = "new-name"
	testComputeVolumeV3NewDescription = "new-description"
)

const (
	testAccComputeVolumeV3ResourceName = "selectel_compute_volume_v3.test"

	testAccComputeVolumeV3AvailabilityZoneEnv       = "SELECTEL_BLOCKSTORAGE_AVAILABILITY_ZONE"
	testAccComputeVolumeV3VolumeTypeEnv             = "SELECTEL_BLOCKSTORAGE_VOLUME_TYPE"
	testAccComputeVolumeV3SecondProjectIDEnv        = "SELECTEL_BLOCKSTORAGE_SECOND_PROJECT_ID"
	testAccComputeVolumeV3SecondRegionEnv           = "SELECTEL_BLOCKSTORAGE_SECOND_REGION"
	testAccComputeVolumeV3SecondRegionZoneEnv       = "SELECTEL_BLOCKSTORAGE_SECOND_REGION_AVAILABILITY_ZONE"
	testAccComputeVolumeV3SecondRegionVolumeTypeEnv = "SELECTEL_BLOCKSTORAGE_SECOND_REGION_VOLUME_TYPE"
	testAccComputeVolumeV3SnapshotIDEnv             = "SELECTEL_BLOCKSTORAGE_SNAPSHOT_ID"
	testAccComputeVolumeV3SnapshotSizeEnv           = "SELECTEL_BLOCKSTORAGE_SNAPSHOT_SIZE"
	testAccComputeVolumeV3RegionalRegionEnv         = "SELECTEL_BLOCKSTORAGE_REGIONAL_REGION"
	testAccComputeVolumeV3RegionalZoneEnv           = "SELECTEL_BLOCKSTORAGE_REGIONAL_ZONE"
	testAccComputeVolumeV3RegionalTypeEnv           = "SELECTEL_BLOCKSTORAGE_REGIONAL_TYPE"
)

func TestAccSelectelComputeVolumeV3_Basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-basic")
	initialConfig := testAccComputeVolumeV3Config(name, "initial description", "initial", 1)
	updatedConfig := testAccComputeVolumeV3Config(name+"-updated", "updated description", "updated", 1)

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
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3UniqueName(t, testAccComputeVolumeV3ResourceName),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "name", name),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "size", "1"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "metadata.purpose", "initial"),
					resource.TestCheckNoResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"metadata."+computeVolumeV3ReservedMetadataKey,
					),
				),
			},
			{
				Config:   initialConfig,
				PlanOnly: true,
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3UniqueName(t, testAccComputeVolumeV3ResourceName),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(testAccComputeVolumeV3ResourceName, "metadata.purpose", "updated"),
				),
			},
			{
				Config:   updatedConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestAccSelectelComputeVolumeV3_Sources(t *testing.T) {
	t.Run("volume", func(t *testing.T) {
		sourceName := acctest.RandomWithPrefix("tf-acc-selectel-volume-source")
		cloneName := acctest.RandomWithPrefix("tf-acc-selectel-volume-clone")

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
			Steps: []resource.TestStep{{
				Config: testAccComputeVolumeV3FromVolumeConfig(sourceName, cloneName),
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
						"source_vol_id",
						"selectel_compute_volume_v3.source",
						"id",
					),
					testAccCheckComputeVolumeV3UniqueName(
						t,
						"selectel_compute_volume_v3.test",
					),
				),
			}},
		})
	})

	t.Run("snapshot", func(t *testing.T) {
		name := acctest.RandomWithPrefix("tf-acc-selectel-volume-snapshot")
		sourceSize, _ := strconv.Atoi(os.Getenv(testAccComputeVolumeV3SnapshotSizeEnv))

		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccComputeVolumeV3PreCheck(
					t,
					testAccComputeVolumeV3AvailabilityZoneEnv,
					testAccComputeVolumeV3VolumeTypeEnv,
					testAccComputeVolumeV3SnapshotIDEnv,
					testAccComputeVolumeV3SnapshotSizeEnv,
				)
				testAccComputeVolumeV3PositiveSizePreCheck(t, testAccComputeVolumeV3SnapshotSizeEnv)
			},
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckComputeVolumeV3Destroy(t),
			Steps: []resource.TestStep{{
				Config: testAccComputeVolumeV3FromFixtureConfig(
					name,
					"snapshot_id",
					os.Getenv(testAccComputeVolumeV3SnapshotIDEnv),
					sourceSize,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					testAccCheckComputeVolumeV3UniqueName(t, testAccComputeVolumeV3ResourceName),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"snapshot_id",
						os.Getenv(testAccComputeVolumeV3SnapshotIDEnv),
					),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"size",
						strconv.Itoa(sourceSize),
					),
				),
			}},
		})
	})
}

func TestAccSelectelComputeVolumeV3_Resize(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-resize")
	initialConfig := testAccComputeVolumeV3Config(name, "resize test", "resize", 1)
	resizedConfig := testAccComputeVolumeV3Config(name, "resize test", "resize", 2)
	var initialID string

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
				),
			},
			{
				Config:   resizedConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestAccSelectelComputeVolumeV3_Replacement(t *testing.T) {
	initialProjectID := os.Getenv("INFRA_PROJECT_ID")
	initialRegion := os.Getenv("INFRA_REGION")
	initialAvailabilityZone := os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv)
	initialVolumeType := os.Getenv(testAccComputeVolumeV3VolumeTypeEnv)

	for _, testCase := range []struct {
		name                        string
		replacementProjectID        string
		replacementRegion           string
		replacementAvailabilityZone string
		replacementVolumeType       string
		replacementAttribute        string
		replacementAttributeValue   string
		fixtureEnvNames             []string
	}{
		{
			name:                        "project_id",
			replacementProjectID:        os.Getenv(testAccComputeVolumeV3SecondProjectIDEnv),
			replacementRegion:           initialRegion,
			replacementAvailabilityZone: initialAvailabilityZone,
			replacementVolumeType:       initialVolumeType,
			replacementAttribute:        "project_id",
			replacementAttributeValue:   os.Getenv(testAccComputeVolumeV3SecondProjectIDEnv),
			fixtureEnvNames: []string{
				testAccComputeVolumeV3AvailabilityZoneEnv,
				testAccComputeVolumeV3VolumeTypeEnv,
				testAccComputeVolumeV3SecondProjectIDEnv,
			},
		},
		{
			name:                        "region",
			replacementProjectID:        initialProjectID,
			replacementRegion:           os.Getenv(testAccComputeVolumeV3SecondRegionEnv),
			replacementAvailabilityZone: os.Getenv(testAccComputeVolumeV3SecondRegionZoneEnv),
			replacementVolumeType:       os.Getenv(testAccComputeVolumeV3SecondRegionVolumeTypeEnv),
			replacementAttribute:        "region",
			replacementAttributeValue:   os.Getenv(testAccComputeVolumeV3SecondRegionEnv),
			fixtureEnvNames: []string{
				testAccComputeVolumeV3AvailabilityZoneEnv,
				testAccComputeVolumeV3VolumeTypeEnv,
				testAccComputeVolumeV3SecondRegionEnv,
				testAccComputeVolumeV3SecondRegionZoneEnv,
				testAccComputeVolumeV3SecondRegionVolumeTypeEnv,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			name := acctest.RandomWithPrefix("tf-acc-selectel-volume-replacement-" + testCase.name)
			initialConfig := testAccComputeVolumeV3ConfigWithScope(
				initialProjectID,
				initialRegion,
				initialAvailabilityZone,
				initialVolumeType,
				name,
				1,
				"",
				"",
			)
			replacementConfig := testAccComputeVolumeV3ConfigWithScope(
				testCase.replacementProjectID,
				testCase.replacementRegion,
				testCase.replacementAvailabilityZone,
				testCase.replacementVolumeType,
				name,
				1,
				"",
				"",
			)
			var initialID string

			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					testAccComputeVolumeV3PreCheck(t, testCase.fixtureEnvNames...)
					if testCase.replacementAttribute == "project_id" &&
						testCase.replacementProjectID == initialProjectID {
						t.Fatalf(
							"%s must differ from INFRA_PROJECT_ID for the project replacement scenario",
							testAccComputeVolumeV3SecondProjectIDEnv,
						)
					}
					if testCase.replacementAttribute == "region" &&
						testCase.replacementRegion == initialRegion {
						t.Fatalf(
							"%s must differ from INFRA_REGION for the region replacement scenario",
							testAccComputeVolumeV3SecondRegionEnv,
						)
					}
				},
				ProviderFactories: testAccProviders,
				CheckDestroy:      testAccCheckComputeVolumeV3Destroy(t),
				Steps: []resource.TestStep{
					{
						Config: initialConfig,
						Check: resource.ComposeTestCheckFunc(
							testAccCheckComputeVolumeV3Exists(
								t,
								testAccComputeVolumeV3ResourceName,
								&initialID,
							),
							testAccCheckComputeVolumeV3UniqueName(
								t,
								testAccComputeVolumeV3ResourceName,
							),
						),
					},
					{
						Config: replacementConfig,
						Check: resource.ComposeTestCheckFunc(
							testAccCheckComputeVolumeV3Exists(
								t,
								testAccComputeVolumeV3ResourceName,
								nil,
							),
							testAccCheckComputeVolumeV3ID(
								testAccComputeVolumeV3ResourceName,
								&initialID,
								true,
							),
							testAccCheckComputeVolumeV3DeletedInScope(
								t,
								&initialID,
								initialProjectID,
								initialRegion,
							),
							testAccCheckComputeVolumeV3UniqueName(
								t,
								testAccComputeVolumeV3ResourceName,
							),
							resource.TestCheckResourceAttr(
								testAccComputeVolumeV3ResourceName,
								testCase.replacementAttribute,
								testCase.replacementAttributeValue,
							),
						),
					},
				},
			})
		})
	}
}

func TestAccSelectelComputeVolumeV3_Import(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-import")

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
				Config: testAccComputeVolumeV3Config(name, "import test", "import", 1),
			},
			{
				ResourceName:      testAccComputeVolumeV3ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"enable_online_resize",
					"image_id",
					"snapshot_id",
					"source_vol_id",
					"backup_id",
				},
			},
		},
	})
}

func TestAccSelectelComputeVolumeV3_Drift(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-drift")
	driftedName := name + "-outside-terraform"
	config := testAccComputeVolumeV3Config(name, "drift test", "drift", 1)
	var volumeID string

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
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.TestCheckResourceAttr(
					testAccComputeVolumeV3ResourceName,
					"name",
					driftedName,
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeVolumeV3Exists(t, testAccComputeVolumeV3ResourceName, nil),
					resource.TestCheckResourceAttr(
						testAccComputeVolumeV3ResourceName,
						"name",
						name,
					),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
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
	client, err := testAccComputeVolumeV3ClientForScope(t, projectID, region)
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
	name := acctest.RandomWithPrefix("tf-acc-selectel-volume-regional-type")
	config := testAccComputeVolumeV3ConfigWithScope(
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv(testAccComputeVolumeV3RegionalRegionEnv),
		os.Getenv(testAccComputeVolumeV3RegionalZoneEnv),
		os.Getenv(testAccComputeVolumeV3RegionalTypeEnv),
		name,
		1,
		"regional type test",
		"regional-type",
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccComputeVolumeV3PreCheck(
				t,
				testAccComputeVolumeV3RegionalRegionEnv,
				testAccComputeVolumeV3RegionalZoneEnv,
				testAccComputeVolumeV3RegionalTypeEnv,
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
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testAccComputeVolumeV3PreCheck(t *testing.T, fixtureEnvNames ...string) {
	t.Helper()

	testAccSelectelPreCheckWithAuth(t)
	for _, envName := range []string{"INFRA_PROJECT_ID", "INFRA_REGION"} {
		if os.Getenv(envName) == "" {
			t.Fatalf("%s must be set for Block Storage acceptance tests", envName)
		}
	}
	for _, envName := range fixtureEnvNames {
		if os.Getenv(envName) == "" {
			t.Fatalf("%s must be set for Block Storage acceptance tests", envName)
		}
	}
}

func testAccComputeVolumeV3PositiveSizePreCheck(t *testing.T, envName string) {
	t.Helper()

	size, err := strconv.Atoi(os.Getenv(envName))
	if err != nil || size <= 0 {
		t.Fatalf("%s must be set to a positive integer for Block Storage acceptance tests", envName)
	}
}

func testAccComputeVolumeV3Config(
	name string,
	description string,
	metadataValue string,
	size int,
) string {
	return testAccComputeVolumeV3ConfigWithScope(
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv),
		os.Getenv(testAccComputeVolumeV3VolumeTypeEnv),
		name,
		size,
		description,
		metadataValue,
	)
}

func testAccComputeVolumeV3ConfigWithScope(
	projectID string,
	region string,
	availabilityZone string,
	volumeType string,
	name string,
	size int,
	description string,
	metadataValue string,
) string {
	return fmt.Sprintf(`
resource "selectel_compute_volume_v3" "test" {
  project_id        = %q
  region            = %q
  availability_zone = %q
  volume_type       = %q
  name              = %q
  description       = %q
  size              = %d

  metadata = {
    purpose = %q
  }
}
`,
		projectID,
		region,
		availabilityZone,
		volumeType,
		name,
		description,
		size,
		metadataValue,
	)
}

func testAccComputeVolumeV3FromVolumeConfig(sourceName string, cloneName string) string {
	return fmt.Sprintf(`
resource "selectel_compute_volume_v3" "source" {
  project_id         = %q
  region             = %q
  availability_zone  = %q
  volume_type        = %q
  name               = %q
  size               = 1
}

resource "selectel_compute_volume_v3" "test" {
  project_id         = %q
  region             = %q
  availability_zone  = %q
  name               = %q
  size               = 1
  source_vol_id      = selectel_compute_volume_v3.source.id
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv),
		os.Getenv(testAccComputeVolumeV3VolumeTypeEnv),
		sourceName,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv),
		cloneName,
	)
}

func testAccComputeVolumeV3FromFixtureConfig(
	name string,
	sourceField string,
	sourceID string,
	size int,
) string {
	return fmt.Sprintf(`
resource "selectel_compute_volume_v3" "test" {
  project_id         = %q
  region             = %q
  availability_zone  = %q
  volume_type        = %q
  name               = %q
  size               = %d
  %s                 = %q
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		os.Getenv(testAccComputeVolumeV3AvailabilityZoneEnv),
		os.Getenv(testAccComputeVolumeV3VolumeTypeEnv),
		name,
		size,
		sourceField,
		sourceID,
	)
}

func testAccCheckComputeVolumeV3Destroy(t *testing.T) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		for _, resourceState := range state.RootModule().Resources {
			if resourceState.Type != "selectel_compute_volume_v3" {
				continue
			}

			client, err := testAccComputeVolumeV3Client(t, resourceState)
			if err != nil {
				return err
			}

			if err := testAccVerifyComputeVolumeV3Deleted(
				t.Context(),
				client,
				resourceState.Primary.ID,
			); err != nil {
				return fmt.Errorf(
					"failed to verify deletion of Block Storage volume %s: %w",
					resourceState.Primary.ID,
					err,
				)
			}
		}

		return nil
	}
}

func testAccCheckComputeVolumeV3Exists(
	t *testing.T,
	resourceName string,
	idTarget *string,
) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume resource %s was not found in state", resourceName)
		}
		if resourceState.Primary.ID == "" {
			return fmt.Errorf("Block Storage volume resource %s has no ID", resourceName)
		}

		client, err := testAccComputeVolumeV3Client(t, resourceState)
		if err != nil {
			return err
		}
		found, _, err := volume.Get(t.Context(), client, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read Block Storage volume %s: %w", resourceState.Primary.ID, err)
		}
		if found.ID != resourceState.Primary.ID {
			return fmt.Errorf(
				"Block Storage API returned volume %s for state ID %s",
				found.ID,
				resourceState.Primary.ID,
			)
		}
		if idTarget != nil {
			*idTarget = found.ID
		}

		return nil
	}
}

func testAccCheckComputeVolumeV3UniqueName(
	t *testing.T,
	resourceName string,
) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume resource %s was not found in state", resourceName)
		}

		client, err := testAccComputeVolumeV3Client(t, resourceState)
		if err != nil {
			return err
		}
		volumes, err := volume.ListDetail(t.Context(), client, volume.ListOpts{})
		if err != nil {
			return fmt.Errorf("failed to list Block Storage volumes: %w", err)
		}

		name := resourceState.Primary.Attributes["name"]
		matchingIDs := make([]string, 0, 1)
		for _, candidate := range volumes {
			if candidate.Name == name {
				matchingIDs = append(matchingIDs, candidate.ID)
			}
		}
		if len(matchingIDs) != 1 || matchingIDs[0] != resourceState.Primary.ID {
			return fmt.Errorf(
				"expected exactly one Block Storage volume named %q with ID %s, found IDs %v",
				name,
				resourceState.Primary.ID,
				matchingIDs,
			)
		}

		return nil
	}
}

func testAccCheckComputeVolumeV3ID(
	resourceName string,
	previousID *string,
	expectReplacement bool,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Block Storage volume resource %s was not found in state", resourceName)
		}
		if *previousID == "" {
			return fmt.Errorf("previous Block Storage volume ID was not captured")
		}

		replaced := resourceState.Primary.ID != *previousID
		if replaced != expectReplacement {
			return fmt.Errorf(
				"Block Storage volume replacement=%t, expected %t; previous ID %s, current ID %s",
				replaced,
				expectReplacement,
				*previousID,
				resourceState.Primary.ID,
			)
		}

		return nil
	}
}

func testAccCheckComputeVolumeV3DeletedInScope(
	t *testing.T,
	volumeID *string,
	projectID string,
	region string,
) resource.TestCheckFunc {
	t.Helper()

	return func(_ *terraform.State) error {
		if *volumeID == "" {
			return fmt.Errorf("previous Block Storage volume ID was not captured")
		}

		client, err := testAccComputeVolumeV3ClientForScope(t, projectID, region)
		if err != nil {
			return err
		}
		if err := testAccVerifyComputeVolumeV3Deleted(t.Context(), client, *volumeID); err != nil {
			return fmt.Errorf(
				"failed to verify deletion of replaced Block Storage volume %s in project %s and region %s: %w",
				*volumeID,
				projectID,
				region,
				err,
			)
		}

		return nil
	}
}

func testAccVerifyComputeVolumeV3Deleted(
	ctx context.Context,
	client *blockstorage.Client,
	volumeID string,
) error {
	_, _, err := volume.Get(ctx, client, volumeID)
	if err == nil {
		return fmt.Errorf("Block Storage volume %s still exists", volumeID)
	}
	if !blockstorage.IsKind(err, blockstorage.KindNotFound) {
		return fmt.Errorf("failed to read Block Storage volume %s: %w", volumeID, err)
	}

	volumes, err := volume.ListDetail(ctx, client, volume.ListOpts{})
	if err != nil {
		return fmt.Errorf("failed to confirm absence of Block Storage volume %s: %w", volumeID, err)
	}
	for _, candidate := range volumes {
		if candidate.ID == volumeID {
			return fmt.Errorf(
				"Block Storage volume %s was hidden by GET but remains in the project list",
				volumeID,
			)
		}
	}

	return nil
}

func testAccComputeVolumeV3Client(
	t *testing.T,
	resourceState *terraform.ResourceState,
) (*blockstorage.Client, error) {
	t.Helper()

	return testAccComputeVolumeV3ClientForScope(
		t,
		resourceState.Primary.Attributes["project_id"],
		resourceState.Primary.Attributes["region"],
	)
}

func testAccComputeVolumeV3ClientForScope(
	t *testing.T,
	projectID string,
	region string,
) (*blockstorage.Client, error) {
	t.Helper()

	resourceData := schema.TestResourceDataRaw(
		t,
		resourceComputeVolumeV3Schema(),
		map[string]any{
			"project_id": projectID,
			"region":     region,
			"size":       1,
		},
	)
	client, diagnostics := getBlockStorageClient(resourceData, testAccProvider.Meta())
	if diagnostics.HasError() {
		return nil, fmt.Errorf(
			"failed to create Block Storage acceptance-test client: %s",
			diagnostics[0].Summary,
		)
	}

	return client, nil
}

func TestUnitSelectelComputeVolumeV3ConfigurationValidation(t *testing.T) {
	resourceSchema := resourceComputeVolumeV3Schema()
	testCases := []struct {
		name         string
		config       map[string]any
		expectsError bool
	}{
		{
			name: "valid minimum",
			config: map[string]any{
				"project_id": "project-id",
				"region":     "ru-1",
				"size":       1,
			},
		},
		{
			name: "zero size",
			config: map[string]any{
				"project_id": "project-id",
				"region":     "ru-1",
				"size":       0,
			},
			expectsError: true,
		},
		{
			name: "conflicting sources",
			config: map[string]any{
				"project_id":  "project-id",
				"region":      "ru-1",
				"size":        1,
				"snapshot_id": "snapshot-id",
				"image_id":    "image-id",
			},
			expectsError: true,
		},
		{
			name: "arbitrary source ID is validated by Cinder",
			config: map[string]any{
				"project_id":  "project-id",
				"region":      "ru-1",
				"size":        1,
				"snapshot_id": "not-a-uuid",
			},
		},
		{
			name: "reserved metadata key",
			config: map[string]any{
				"project_id": "project-id",
				"region":     "ru-1",
				"size":       1,
				"metadata": map[string]any{
					computeVolumeV3ReservedMetadataKey: "user-value",
				},
			},
			expectsError: true,
		},
		{
			name: "user metadata",
			config: map[string]any{
				"project_id": "project-id",
				"region":     "ru-1",
				"size":       1,
				"metadata": map[string]any{
					"environment": "test",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := schema.InternalMap(resourceSchema).Validate(
				terraform.NewResourceConfigRaw(testCase.config),
			)

			assert.Equal(t, testCase.expectsError, diagnostics.HasError(), diagnostics)
		})
	}
}

func TestUnitSelectelComputeVolumeV3ReplacementPlan(t *testing.T) {
	testCases := []struct {
		name          string
		field         string
		stateValue    any
		configValue   any
		diffKeyPrefix string
	}{
		{
			name:        "project",
			field:       "project_id",
			stateValue:  "old-project-id",
			configValue: "new-project-id",
		},
		{
			name:        "region",
			field:       "region",
			stateValue:  "ru-1",
			configValue: "ru-2",
		},
		{
			name:        "availability zone",
			field:       "availability_zone",
			stateValue:  "ru-1a",
			configValue: "ru-1b",
		},
		{
			name:        "volume type",
			field:       "volume_type",
			stateValue:  "fast.ru-1a",
			configValue: "universal.ru-1",
		},
		{
			name:        "snapshot source",
			field:       "snapshot_id",
			stateValue:  "old-snapshot-id",
			configValue: "new-snapshot-id",
		},
		{
			name:        "volume source",
			field:       "source_vol_id",
			stateValue:  "old-volume-id",
			configValue: "new-volume-id",
		},
		{
			name:        "image source",
			field:       "image_id",
			stateValue:  "old-image-id",
			configValue: "new-image-id",
		},
		{
			name:        "backup source",
			field:       "backup_id",
			stateValue:  "old-backup-id",
			configValue: "new-backup-id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diff := testComputeVolumeV3PlanDiff(
				t,
				map[string]any{testCase.field: testCase.stateValue},
				map[string]any{testCase.field: testCase.configValue},
			)

			diffKeyPrefix := testCase.diffKeyPrefix
			if diffKeyPrefix == "" {
				diffKeyPrefix = testCase.field
			}
			assertPlanRequiresReplacement(t, diff, diffKeyPrefix)
		})
	}
}

func TestUnitSelectelComputeVolumeV3VolumeTypePlan(t *testing.T) {
	testCases := []struct {
		name                    string
		region                  string
		stateVolumeType         string
		configVolumeType        string
		stateAvailabilityZone   string
		configAvailabilityZone  string
		expectsVolumeTypeChange bool
	}{
		{
			name:                    "regional config equals zonal API state",
			region:                  "ru-1",
			stateVolumeType:         "fast.ru-1a",
			configVolumeType:        "fast.ru-1",
			stateAvailabilityZone:   "ru-1a",
			configAvailabilityZone:  "ru-1a",
			expectsVolumeTypeChange: false,
		},
		{
			name:                    "three letter region config equals zonal API state",
			region:                  "gis-1",
			stateVolumeType:         "fast.gis-1a",
			configVolumeType:        "fast.gis-1",
			stateAvailabilityZone:   "gis-1a",
			configAvailabilityZone:  "gis-1a",
			expectsVolumeTypeChange: false,
		},
		{
			name:                    "two digit region config equals zonal API state",
			region:                  "ru-11",
			stateVolumeType:         "fast.ru-11a",
			configVolumeType:        "fast.ru-11",
			stateAvailabilityZone:   "ru-11a",
			configAvailabilityZone:  "ru-11a",
			expectsVolumeTypeChange: false,
		},
		{
			name:                    "different type prefix",
			region:                  "ru-1",
			stateVolumeType:         "fast.ru-1a",
			configVolumeType:        "universal.ru-1",
			stateAvailabilityZone:   "ru-1a",
			configAvailabilityZone:  "ru-1a",
			expectsVolumeTypeChange: true,
		},
		{
			name:                    "different availability zone",
			region:                  "ru-1",
			stateVolumeType:         "fast.ru-1a",
			configVolumeType:        "fast.ru-1",
			stateAvailabilityZone:   "ru-1a",
			configAvailabilityZone:  "ru-1b",
			expectsVolumeTypeChange: true,
		},
		{
			name:                    "foreign availability zone",
			region:                  "ru-1",
			stateVolumeType:         "fast.ru-2a",
			configVolumeType:        "fast.ru-1",
			stateAvailabilityZone:   "ru-2a",
			configAvailabilityZone:  "ru-2a",
			expectsVolumeTypeChange: true,
		},
		{
			name:                    "private cloud account domain",
			region:                  "ru-3",
			stateVolumeType:         "fast.ru-3-56312",
			configVolumeType:        "fast.ru-3",
			stateAvailabilityZone:   "ru-3-56312",
			configAvailabilityZone:  "ru-3-56312",
			expectsVolumeTypeChange: true,
		},
		{
			name:                    "availability zone suffix longer than one symbol",
			region:                  "ru-1",
			stateVolumeType:         "fast.ru-1aa",
			configVolumeType:        "fast.ru-1",
			stateAvailabilityZone:   "ru-1aa",
			configAvailabilityZone:  "ru-1aa",
			expectsVolumeTypeChange: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diff := testComputeVolumeV3PlanDiff(
				t,
				map[string]any{
					"region":            testCase.region,
					"availability_zone": testCase.stateAvailabilityZone,
					"volume_type":       testCase.stateVolumeType,
				},
				map[string]any{
					"region":            testCase.region,
					"availability_zone": testCase.configAvailabilityZone,
					"volume_type":       testCase.configVolumeType,
				},
			)

			volumeTypeDiff, hasVolumeTypeDiff := diffAttribute(diff, "volume_type")
			assert.Equal(t, testCase.expectsVolumeTypeChange, hasVolumeTypeDiff)
			if testCase.expectsVolumeTypeChange {
				assert.True(t, volumeTypeDiff.RequiresNew)
			}
			if !testCase.expectsVolumeTypeChange {
				assert.Nil(t, diff)
			}
		})
	}
}

func TestUnitSelectelComputeVolumeV3FlexibleQoSPlan(t *testing.T) {
	stateMetadata := map[string]any{
		"total_iops_sec":  "30000",
		"total_bytes_sec": "1073741824",
	}

	t.Run("server managed bandwidth is stable", func(t *testing.T) {
		diff := testComputeVolumeV3PlanDiff(
			t,
			map[string]any{"metadata": stateMetadata},
			map[string]any{"metadata": map[string]any{
				"total_iops_sec": "30000",
			}},
		)

		assert.Nil(t, diff)
	})

	t.Run("IOPS changes in place without deleting server managed bandwidth", func(t *testing.T) {
		diff := testComputeVolumeV3PlanDiff(
			t,
			map[string]any{"metadata": stateMetadata},
			map[string]any{"metadata": map[string]any{
				"total_iops_sec": "40000",
			}},
		)

		require.NotNil(t, diff)
		require.NotEmpty(t, diff.Attributes)
		for key, attributeDiff := range diff.Attributes {
			assert.NotContains(t, key, "total_bytes_sec")
			assert.False(t, attributeDiff.RequiresNew)
		}
	})

	t.Run("server managed bandwidth cannot be overridden by configuration", func(t *testing.T) {
		diff := testComputeVolumeV3PlanDiff(
			t,
			map[string]any{"metadata": stateMetadata},
			map[string]any{"metadata": map[string]any{
				"total_iops_sec":  "30000",
				"total_bytes_sec": "1",
			}},
		)

		assert.Nil(t, diff)
	})
}

func TestUnitSelectelComputeVolumeV3ImportRequiresScope(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		config          *Config
		expectedMessage string
	}{
		{
			name:            "missing project",
			config:          &Config{Region: "ru-1"},
			expectedMessage: "INFRA_PROJECT_ID",
		},
		{
			name:            "missing region",
			config:          &Config{ProjectID: "project-id"},
			expectedMessage: "INFRA_REGION",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceComputeVolumeV3Schema(), nil)
			resourceData.SetId("volume-id")

			imported, err := resourceComputeVolumeV3ImportState(t.Context(), resourceData, testCase.config)

			require.ErrorContains(t, err, testCase.expectedMessage)
			assert.Nil(t, imported)
		})
	}
}

func TestUnitSelectelComputeVolumeV3ImportSetsScope(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceComputeVolumeV3Schema(), nil)
	resourceData.SetId("volume-id")

	imported, err := resourceComputeVolumeV3ImportState(t.Context(), resourceData, &Config{
		ProjectID: "project-id",
		Region:    "ru-1",
	})

	require.NoError(t, err)
	require.Len(t, imported, 1)
	assert.Same(t, resourceData, imported[0])
	assert.Equal(t, "volume-id", imported[0].Id())
	assert.Equal(t, "project-id", imported[0].Get("project_id"))
	assert.Equal(t, "ru-1", imported[0].Get("region"))
}

func TestUnitSelectelComputeVolumeV3DeleteRetryAfterPollingFailure(t *testing.T) {
	var getCount int
	var deleteCount int
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
			getCount++
			switch getCount {
			case 1:
				writeTestComputeVolumeV3Remote(t, response, testComputeVolumeV3Remote{
					ID:     "volume-id",
					Status: "available",
				})
			case 2:
				writeTestComputeVolumeV3Fault(
					t,
					response,
					http.StatusInternalServerError,
					"poll failed",
				)
			case 3:
				writeTestComputeVolumeV3Remote(t, response, testComputeVolumeV3Remote{
					ID:     "volume-id",
					Status: "deleting",
				})
			default:
				writeTestComputeVolumeV3Fault(
					t,
					response,
					http.StatusNotFound,
					"volume not found",
				)
			}
		case request.Method == http.MethodDelete &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
			deleteCount++
			response.WriteHeader(http.StatusAccepted)
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
			writeTestComputeVolumeV3Page(t, response, nil, "")
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3DeleteResourceData(t)
	client := testComputeVolumeV3Client(server.URL)

	diagnostics := deleteComputeVolumeV3(t.Context(), resourceData, client)
	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "ID remains")
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, 1, deleteCount)

	diagnostics = deleteComputeVolumeV3(t.Context(), resourceData, client)
	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Empty(t, resourceData.Id())
	assert.Equal(t, 1, deleteCount)
}

func TestUnitSelectelComputeVolumeV3DeleteRejectsAttachedVolume(t *testing.T) {
	var requestCount int
	var deleteCount int
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id" {
			response.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
				"volume": map[string]any{
					"id":     "volume-id",
					"status": "in-use",
					"attachments": []map[string]string{{
						"id":            "volume-id",
						"attachment_id": "attachment-id",
						"server_id":     "server-id",
						"device":        "/dev/vdb",
					}},
				},
			}))

			return
		}
		if request.Method == http.MethodDelete {
			deleteCount++
		}
		http.NotFound(response, request)
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3DeleteResourceData(t)
	diagnostics := deleteComputeVolumeV3(
		t.Context(),
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "detach")
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, 1, requestCount)
	assert.Zero(t, deleteCount)
}

func TestUnitSelectelComputeVolumeV3DeleteRequestFailuresKeepState(t *testing.T) {
	var deleteCount int
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			writeTestComputeVolumeV3Remote(t, response, testComputeVolumeV3Remote{
				ID:     "volume-id",
				Status: "available",
			})
		case http.MethodDelete:
			deleteCount++
			writeTestComputeVolumeV3Fault(t, response, http.StatusForbidden, "forbidden")
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3DeleteResourceData(t)
	diagnostics := deleteComputeVolumeV3(
		t.Context(),
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "forbidden")
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, 1, deleteCount)
}

type testComputeVolumeV3DeleteNotFoundRecorder struct {
	t                 *testing.T
	serverURL         string
	notFoundBeforeAPI bool
	listVolume        bool
	pageError         bool
	deleteCount       int
}

func (r *testComputeVolumeV3DeleteNotFoundRecorder) ServeHTTP(
	response http.ResponseWriter,
	request *http.Request,
) {
	switch {
	case request.Method == http.MethodGet &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
		if r.notFoundBeforeAPI {
			writeTestComputeVolumeV3Fault(
				r.t,
				response,
				http.StatusNotFound,
				"volume not found",
			)
		} else {
			writeTestComputeVolumeV3Remote(r.t, response, testComputeVolumeV3Remote{
				ID:     "volume-id",
				Status: "available",
			})
		}
	case request.Method == http.MethodDelete &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
		r.deleteCount++
		writeTestComputeVolumeV3Fault(
			r.t,
			response,
			http.StatusNotFound,
			"volume not found",
		)
	case request.Method == http.MethodGet &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail" &&
		request.URL.Query().Get("marker") == "next":
		writeTestComputeVolumeV3Fault(
			r.t,
			response,
			http.StatusInternalServerError,
			"later page failed",
		)
	case request.Method == http.MethodGet &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
		volumes := []testComputeVolumeV3(nil)
		if r.listVolume {
			volumes = []testComputeVolumeV3{{ID: "volume-id"}}
		}
		nextURL := ""
		if r.pageError {
			nextURL = r.serverURL + testComputeVolumeV3CollectionPath() +
				"/detail?marker=next"
		}
		writeTestComputeVolumeV3Page(r.t, response, volumes, nextURL)
	default:
		http.NotFound(response, request)
	}
}

func TestUnitSelectelComputeVolumeV3DeleteConfirmsNotFound(t *testing.T) {
	for _, testCase := range []struct {
		name              string
		notFoundBeforeAPI bool
		listVolume        bool
		pageError         bool
		expectsErr        bool
	}{
		{name: "repeat after confirmed absence", notFoundBeforeAPI: true},
		{name: "ID remains in list", listVolume: true, expectsErr: true},
		{name: "later page fails", pageError: true, expectsErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := &testComputeVolumeV3DeleteNotFoundRecorder{
				t:                 t,
				notFoundBeforeAPI: testCase.notFoundBeforeAPI,
				listVolume:        testCase.listVolume,
				pageError:         testCase.pageError,
			}
			server := httptest.NewServer(recorder)
			defer server.Close()
			recorder.serverURL = server.URL

			resourceData := testComputeVolumeV3DeleteResourceData(t)
			diagnostics := deleteComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			assert.Equal(t, testCase.expectsErr, diagnostics.HasError(), diagnostics)
			if testCase.expectsErr {
				assert.Equal(t, "volume-id", resourceData.Id())
			} else {
				assert.Empty(t, resourceData.Id())
			}
			if testCase.notFoundBeforeAPI {
				assert.Zero(t, recorder.deleteCount)
			} else {
				assert.Equal(t, 1, recorder.deleteCount)
			}
		})
	}
}

func TestUnitSelectelComputeVolumeV3CreateRequestVariants(t *testing.T) {
	testCases := []struct {
		name            string
		sourceField     string
		sourceWireField string
		expectedVersion string
	}{
		{name: "empty volume"},
		{name: "snapshot", sourceField: "snapshot_id", sourceWireField: "snapshot_id"},
		{name: "source volume", sourceField: "source_vol_id", sourceWireField: "source_volid"},
		{name: "image", sourceField: "image_id", sourceWireField: "imageRef"},
		{
			name:            "backup",
			sourceField:     "backup_id",
			sourceWireField: "backup_id",
			expectedVersion: "volume 3.47",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var postCount int
			var getCount int
			var createToken string

			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodPost && request.URL.Path == testComputeVolumeV3CollectionPath():
					postCount++

					var body map[string]any
					require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
					volumeBody := body["volume"].(map[string]any)
					metadata := volumeBody["metadata"].(map[string]any)

					assert.Equal(t, float64(10), volumeBody["size"])
					assert.Equal(t, "volume-name", volumeBody["name"])
					assert.Equal(t, "volume-description", volumeBody["description"])
					assert.Equal(t, "ru-2a", volumeBody["availability_zone"])
					assert.Equal(t, "fast.ru-2", volumeBody["volume_type"])
					assert.Equal(t, "user-value", metadata["user-key"])
					assert.NotEmpty(t, metadata[computeVolumeV3ReservedMetadataKey])
					assert.Equal(t, testCase.expectedVersion, request.Header.Get("OpenStack-API-Version"))
					createToken = metadata[computeVolumeV3ReservedMetadataKey].(string)

					for _, wireField := range []string{"snapshot_id", "source_volid", "imageRef", "backup_id"} {
						if wireField == testCase.sourceWireField {
							assert.Equal(t, "source-id", volumeBody[wireField])
						} else {
							assert.NotContains(t, volumeBody, wireField)
						}
					}

					assert.NotContains(t, body, "OS-SCH-HNT:scheduler_hints")

					writeTestComputeVolumeV3(t, response, http.StatusAccepted, "volume-id", "creating", metadata)
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
					getCount++
					writeTestComputeVolumeV3Remote(t, response, testComputeVolumeV3Remote{
						ID:               "volume-id",
						Status:           "available",
						Size:             10,
						Name:             "server-volume-name",
						Description:      "server-volume-description",
						AvailabilityZone: "ru-2a",
						VolumeType:       "fast.ru-2a",
						Metadata: map[string]string{
							"user-key":                         "server-user-value",
							"server-added-key":                 "server-value",
							computeVolumeV3ReservedMetadataKey: createToken,
						},
					})
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			values := testComputeVolumeV3CreateValues()
			if testCase.sourceField != "" {
				values[testCase.sourceField] = "source-id"
			}
			resourceData := testComputeVolumeV3ResourceData(t, values)

			diagnostics := createComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.False(t, diagnostics.HasError(), diagnostics)
			assert.Equal(t, "volume-id", resourceData.Id())
			assert.Equal(t, 1, postCount)
			assert.Equal(t, 2, getCount)
			assert.Equal(t, "server-volume-name", resourceData.Get("name"))
			assert.Equal(t, "server-volume-description", resourceData.Get("description"))
			assert.Equal(t, "fast.ru-2a", resourceData.Get("volume_type"))
			assert.Equal(t, map[string]any{
				"user-key":         "server-user-value",
				"server-added-key": "server-value",
			}, resourceData.Get("metadata"))
			assert.NotContains(t, resourceData.Get("metadata"), computeVolumeV3ReservedMetadataKey)
		})
	}
}

func TestUnitSelectelComputeVolumeV3CreateDefinitiveFailures(t *testing.T) {
	testCases := []struct {
		name   string
		status int
	}{
		{name: "invalid request", status: http.StatusBadRequest},
		{name: "forbidden", status: http.StatusForbidden},
		{name: "rejected microversion", status: http.StatusNotAcceptable},
		{name: "over quota", status: http.StatusRequestEntityTooLarge},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var postCount int
			var listCount int

			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch request.Method {
				case http.MethodPost:
					postCount++
					response.Header().Set("X-Openstack-Request-Id", "req-definitive")
					if testCase.status == http.StatusRequestEntityTooLarge {
						writeTestComputeVolumeV3OverQuota(t, response)
					} else {
						writeTestComputeVolumeV3Fault(t, response, testCase.status, testCase.name)
					}
				case http.MethodGet:
					listCount++
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			diagnostics := createComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.True(t, diagnostics.HasError())
			assert.Contains(t, diagnostics[0].Summary, "rejected")
			assert.Empty(t, resourceData.Id())
			assert.Equal(t, 1, postCount)
			assert.Zero(t, listCount)
		})
	}
}

func TestUnitSelectelComputeVolumeV3CreateRecoversAmbiguousHTTPFailures(t *testing.T) {
	testCases := []struct {
		name   string
		status int
	}{
		{name: "conflict", status: http.StatusConflict},
		{name: "non quota 413", status: http.StatusRequestEntityTooLarge},
		{name: "unknown client response", status: http.StatusTeapot},
		{name: "server error", status: http.StatusInternalServerError},
		{name: "unstructured invalid request", status: http.StatusBadRequest},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var createToken string
			var postCount int
			var listCount int

			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodPost:
					postCount++
					createToken = readTestComputeVolumeV3CreateToken(t, request)
					response.Header().Set("X-Openstack-Request-Id", "req-ambiguous")
					if testCase.name == "unstructured invalid request" {
						response.WriteHeader(testCase.status)
						_, err := response.Write([]byte("invalid request without a Cinder fault"))
						require.NoError(t, err)

						return
					}
					writeTestComputeVolumeV3Fault(t, response, testCase.status, testCase.name)
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
					listCount++
					writeTestComputeVolumeV3Page(t, response, []testComputeVolumeV3{{
						ID:       "recovered-id",
						Status:   "creating",
						Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: createToken},
					}}, "")
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/recovered-id":
					writeTestComputeVolumeV3(t, response, http.StatusOK, "recovered-id", "available", nil)
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			diagnostics := createComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.False(t, diagnostics.HasError(), diagnostics)
			assert.Equal(t, "recovered-id", resourceData.Id())
			assert.Equal(t, 1, postCount)
			assert.Equal(t, 1, listCount)
		})
	}
}

func TestUnitSelectelComputeVolumeV3CreateRecoversTransportFailure(t *testing.T) {
	var createToken string

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
			writeTestComputeVolumeV3Page(t, response, []testComputeVolumeV3{{
				ID:       "recovered-id",
				Status:   "creating",
				Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: createToken},
			}}, "")
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/recovered-id":
			writeTestComputeVolumeV3(t, response, http.StatusOK, "recovered-id", "available", nil)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	httpClient := &testComputeVolumeV3CreateErrorClient{
		delegate:  server.Client(),
		createErr: errors.New("lost create response"),
		token:     &createToken,
	}
	client, err := blockstorage.NewClient(blockstorage.Config{
		Endpoint:   server.URL + testComputeVolumeV3EndpointPath(),
		Token:      testBlockStorageToken,
		HTTPClient: httpClient,
	})
	require.NoError(t, err)
	resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())

	diagnostics := createComputeVolumeV3(t.Context(), resourceData, client)

	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Equal(t, "recovered-id", resourceData.Id())
	assert.Equal(t, 1, httpClient.postCount)
}

func TestUnitSelectelComputeVolumeV3CreateRecoveryFailuresKeepPendingID(t *testing.T) {
	testCases := []struct {
		name             string
		writeList        func(*testing.T, http.ResponseWriter, string, string)
		expectedMessages []string
	}{
		{
			name: "zero matches after non quota 413",
			writeList: func(t *testing.T, response http.ResponseWriter, _, _ string) {
				writeTestComputeVolumeV3Page(t, response, nil, "")
			},
			expectedMessages: []string{"ambiguous", "pending:", "will not repeat POST"},
		},
		{
			name: "multiple matches",
			writeList: func(t *testing.T, response http.ResponseWriter, token, _ string) {
				writeTestComputeVolumeV3Page(t, response, []testComputeVolumeV3{
					{ID: "first-id", Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: token}},
					{ID: "second-id", Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: token}},
				}, "")
			},
			expectedMessages: []string{"found 2", "first-id", "second-id", "will not repeat POST"},
		},
		{
			name: "incomplete list",
			writeList: func(t *testing.T, response http.ResponseWriter, _, serverURL string) {
				writeTestComputeVolumeV3Page(
					t,
					response,
					nil,
					serverURL+testComputeVolumeV3CollectionPath()+"/detail?marker=next",
				)
			},
			expectedMessages: []string{"complete Block Storage volume list", "incomplete_list", "will not repeat POST"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var createToken string
			var postCount int
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodPost:
					postCount++
					createToken = readTestComputeVolumeV3CreateToken(t, request)
					response.Header().Set("X-Openstack-Request-Id", "req-non-quota-413")
					writeTestComputeVolumeV3Fault(
						t,
						response,
						http.StatusRequestEntityTooLarge,
						"request metadata is too large",
					)
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail" &&
					request.URL.Query().Get("marker") == "":
					testCase.writeList(t, response, createToken, server.URL)
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
					writeTestComputeVolumeV3Fault(t, response, http.StatusInternalServerError, "page failed")
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			ctx := t.Context()
			if testCase.name == "zero matches after non quota 413" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Millisecond)
				defer cancel()
			}

			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			diagnostics := createComputeVolumeV3(
				ctx,
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.True(t, diagnostics.HasError())
			assert.True(t, strings.HasPrefix(resourceData.Id(), computeVolumeV3PendingIDPrefix))
			assert.Equal(t, 1, postCount)
			for _, expectedMessage := range testCase.expectedMessages {
				assert.Contains(t, diagnostics[0].Summary, expectedMessage)
			}
		})
	}
}

func TestUnitSelectelComputeVolumeV3CreateWaitTimeoutKeepsVolumeID(t *testing.T) {
	var postCount int

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			postCount++
			writeTestComputeVolumeV3(t, response, http.StatusAccepted, "volume-id", "creating", nil)
		case http.MethodGet:
			writeTestComputeVolumeV3(t, response, http.StatusOK, "volume-id", "creating", nil)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()

	resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
	diagnostics := createComputeVolumeV3(
		ctx,
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "failed waiting")
	assert.Contains(t, diagnostics[0].Summary, "volume-id")
	assert.Contains(t, diagnostics[0].Summary, "last observed status: creating")
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, 1, postCount)
}

func TestUnitSelectelComputeVolumeV3ReadRefreshesState(t *testing.T) {
	var getCount int
	var postCount int

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
			getCount++
			response.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
				"volume": map[string]any{
					"id":                "volume-id",
					"status":            "in-use",
					"size":              20,
					"name":              "remote-name",
					"description":       "remote-description",
					"availability_zone": "ru-2b",
					"volume_type":       "fast.ru-2b",
					"snapshot_id":       "snapshot-id",
					"source_volid":      "source-volume-id",
					"metadata": map[string]string{
						"user-key":                         "remote-value",
						"server-added-key":                 "server-value",
						computeVolumeV3ReservedMetadataKey: "create-token",
					},
					"attachments": []map[string]string{{
						"id":            "volume-id",
						"attachment_id": "attachment-id",
						"server_id":     "server-id",
						"device":        "/dev/vdb",
					}},
				},
			}))
		case request.Method == http.MethodPost:
			postCount++
			http.NotFound(response, request)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
	resourceData.SetId("volume-id")

	diagnostics := readComputeVolumeV3(
		t.Context(),
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, 20, resourceData.Get("size"))
	assert.Equal(t, "remote-name", resourceData.Get("name"))
	assert.Equal(t, "remote-description", resourceData.Get("description"))
	assert.Equal(t, "ru-2b", resourceData.Get("availability_zone"))
	assert.Equal(t, "fast.ru-2b", resourceData.Get("volume_type"))
	assert.Equal(t, "snapshot-id", resourceData.Get("snapshot_id"))
	assert.Equal(t, "source-volume-id", resourceData.Get("source_vol_id"))
	assert.Equal(t, map[string]any{
		"user-key":         "remote-value",
		"server-added-key": "server-value",
	}, resourceData.Get("metadata"))

	attachments := resourceData.Get("attachment").(*schema.Set).List()
	require.Len(t, attachments, 1)
	assert.Equal(t, map[string]any{
		"id":          "volume-id",
		"instance_id": "server-id",
		"device":      "/dev/vdb",
	}, attachments[0])
	assert.Equal(t, 1, getCount)
	assert.Zero(t, postCount)
}

func TestUnitSelectelComputeVolumeV3ReadForbiddenKeepsState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		writeTestComputeVolumeV3Fault(t, response, http.StatusForbidden, "forbidden")
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
	resourceData.SetId("volume-id")

	diagnostics := readComputeVolumeV3(
		t.Context(),
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "forbidden")
	assert.Equal(t, "volume-id", resourceData.Id())
	assert.Equal(t, "volume-name", resourceData.Get("name"))
}

func TestUnitSelectelComputeVolumeV3ReadConfirmsNotFound(t *testing.T) {
	testCases := []struct {
		name            string
		listContainsID  bool
		expectedMessage string
	}{
		{name: "confirmed absence"},
		{
			name:            "masked authorization failure",
			listContainsID:  true,
			expectedMessage: "masked authorization failure",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var directGetCount int
			var listCount int
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
					directGetCount++
					writeTestComputeVolumeV3Fault(t, response, http.StatusNotFound, "volume not found")
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
					listCount++
					var volumes []testComputeVolumeV3
					if testCase.listContainsID {
						volumes = []testComputeVolumeV3{{ID: "volume-id"}}
					}
					nextURL := ""
					if !testCase.listContainsID && request.URL.Query().Get("marker") == "" {
						volumes = []testComputeVolumeV3{{ID: "another-volume-id"}}
						nextURL = server.URL + testComputeVolumeV3CollectionPath() +
							"/detail?marker=another-volume-id"
					}
					writeTestComputeVolumeV3Page(t, response, volumes, nextURL)
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			resourceData.SetId("volume-id")

			diagnostics := readComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			assert.Equal(t, 1, directGetCount)
			if testCase.expectedMessage == "" {
				assert.Equal(t, 2, listCount)
				require.False(t, diagnostics.HasError(), diagnostics)
				assert.Empty(t, resourceData.Id())
			} else {
				assert.Equal(t, 1, listCount)
				require.True(t, diagnostics.HasError())
				assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)
				assert.Equal(t, "volume-id", resourceData.Id())
			}
		})
	}
}

func TestUnitSelectelComputeVolumeV3ReadUnconfirmedNotFoundKeepsState(t *testing.T) {
	testCases := []struct {
		name            string
		listStatus      int
		incomplete      bool
		expectedMessage string
	}{
		{
			name:            "forbidden list",
			listStatus:      http.StatusForbidden,
			expectedMessage: "forbidden",
		},
		{
			name:            "incomplete list",
			listStatus:      http.StatusInternalServerError,
			incomplete:      true,
			expectedMessage: "incomplete_list",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
					writeTestComputeVolumeV3Fault(t, response, http.StatusNotFound, "volume not found")
				case request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail" &&
					testCase.incomplete &&
					request.URL.Query().Get("marker") == "":
					writeTestComputeVolumeV3Page(
						t,
						response,
						nil,
						server.URL+testComputeVolumeV3CollectionPath()+"/detail?marker=next",
					)
				case request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
					writeTestComputeVolumeV3Fault(t, response, testCase.listStatus, testCase.name)
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			resourceData.SetId("volume-id")

			diagnostics := readComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.True(t, diagnostics.HasError())
			assert.Contains(t, diagnostics[0].Summary, testCase.expectedMessage)
			assert.Contains(t, diagnostics[0].Summary, "confirm")
			assert.Equal(t, "volume-id", resourceData.Id())
			assert.Equal(t, "volume-name", resourceData.Get("name"))
		})
	}
}

func TestUnitSelectelComputeVolumeV3ReadRecoversPendingID(t *testing.T) {
	const createToken = "create-token"

	var getCount int
	var listCount int
	var postCount int

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
			listCount++
			writeTestComputeVolumeV3Page(t, response, []testComputeVolumeV3{{
				ID:       "recovered-id",
				Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: createToken},
			}}, "")
		case request.Method == http.MethodGet &&
			request.URL.Path == testComputeVolumeV3CollectionPath()+"/recovered-id":
			getCount++
			response.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
				"volume": map[string]any{
					"id":                "recovered-id",
					"status":            "available",
					"size":              12,
					"name":              "recovered-volume",
					"description":       "recovered-description",
					"availability_zone": "ru-2a",
					"volume_type":       "fast.ru-2a",
					"metadata": map[string]string{
						"user-key":                         "user-value",
						computeVolumeV3ReservedMetadataKey: createToken,
					},
					"attachments": []any{},
				},
			}))
		case request.Method == http.MethodPost:
			postCount++
			http.NotFound(response, request)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
	resourceData.SetId(computeVolumeV3PendingIDPrefix + createToken)

	diagnostics := readComputeVolumeV3(
		t.Context(),
		resourceData,
		testComputeVolumeV3Client(server.URL),
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	assert.Equal(t, "recovered-id", resourceData.Id())
	assert.Equal(t, 12, resourceData.Get("size"))
	assert.Equal(t, "recovered-volume", resourceData.Get("name"))
	assert.Equal(t, map[string]any{"user-key": "user-value"}, resourceData.Get("metadata"))
	assert.Equal(t, 1, listCount)
	assert.Equal(t, 2, getCount)
	assert.Zero(t, postCount)
}

func TestUnitSelectelComputeVolumeV3ReadPendingRecoveryFailuresKeepPendingID(t *testing.T) {
	const createToken = "create-token"

	testCases := []struct {
		name             string
		writeList        func(*testing.T, http.ResponseWriter, string)
		expectedMessages []string
	}{
		{
			name: "zero matches",
			writeList: func(t *testing.T, response http.ResponseWriter, _ string) {
				writeTestComputeVolumeV3Page(t, response, nil, "")
			},
			expectedMessages: []string{"found no volume", "will not repeat POST"},
		},
		{
			name: "multiple matches",
			writeList: func(t *testing.T, response http.ResponseWriter, _ string) {
				writeTestComputeVolumeV3Page(t, response, []testComputeVolumeV3{
					{ID: "first-id", Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: createToken}},
					{ID: "second-id", Metadata: map[string]string{computeVolumeV3ReservedMetadataKey: createToken}},
				}, "")
			},
			expectedMessages: []string{"found 2 volumes", "first-id", "second-id", "will not choose"},
		},
		{
			name: "incomplete list",
			writeList: func(t *testing.T, response http.ResponseWriter, serverURL string) {
				writeTestComputeVolumeV3Page(
					t,
					response,
					nil,
					serverURL+testComputeVolumeV3CollectionPath()+"/detail?marker=next",
				)
			},
			expectedMessages: []string{"server_error", "incomplete_list", "pending:create-token"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var postCount int
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch {
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail" &&
					request.URL.Query().Get("marker") == "":
					testCase.writeList(t, response, server.URL)
				case request.Method == http.MethodGet &&
					request.URL.Path == testComputeVolumeV3CollectionPath()+"/detail":
					writeTestComputeVolumeV3Fault(
						t,
						response,
						http.StatusInternalServerError,
						"page failed",
					)
				case request.Method == http.MethodPost:
					postCount++
					http.NotFound(response, request)
				default:
					http.NotFound(response, request)
				}
			}))
			defer server.Close()

			pendingID := computeVolumeV3PendingIDPrefix + createToken
			resourceData := testComputeVolumeV3ResourceData(t, testComputeVolumeV3CreateValues())
			resourceData.SetId(pendingID)

			diagnostics := readComputeVolumeV3(
				t.Context(),
				resourceData,
				testComputeVolumeV3Client(server.URL),
			)

			require.True(t, diagnostics.HasError())
			assert.Equal(t, pendingID, resourceData.Id())
			assert.Zero(t, postCount)
			for _, expectedMessage := range testCase.expectedMessages {
				assert.Contains(t, diagnostics[0].Summary, expectedMessage)
			}
		})
	}
}

func TestUnitSelectelComputeVolumeV3UpdateCombinesDescriptiveChanges(t *testing.T) {
	remote := testComputeVolumeV3Remote{
		ID:               "volume-id",
		Status:           "available",
		Size:             10,
		Name:             "old-name",
		Description:      "old-description",
		AvailabilityZone: "ru-2a",
		VolumeType:       "fast.ru-2a",
		Metadata:         map[string]string{"old-key": "old-value"},
	}
	recorder := newTestComputeVolumeV3UpdateRecorder(t, remote)
	server := httptest.NewServer(recorder)
	defer server.Close()

	oldValues := testComputeVolumeV3UpdateValues()
	newValues := testComputeVolumeV3UpdateValues()
	newValues["name"] = testComputeVolumeV3NewName
	newValues["description"] = testComputeVolumeV3NewDescription
	newValues["metadata"] = map[string]any{"new-key": "new-value"}

	state, diagnostics := testApplyComputeVolumeV3Update(
		t,
		testComputeVolumeV3Client(server.URL),
		oldValues,
		newValues,
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	require.NotNil(t, state)
	assert.Equal(t, testComputeVolumeV3NewName, state.Attributes["name"])
	assert.Equal(t, testComputeVolumeV3NewDescription, state.Attributes["description"])
	assert.Equal(t, "new-value", state.Attributes["metadata.new-key"])
	assert.Equal(t, []string{"descriptive update"}, recorder.mutationPhases)
	require.Len(t, recorder.updateBodies, 1)
	assert.Equal(t, map[string]any{
		"name":        testComputeVolumeV3NewName,
		"description": testComputeVolumeV3NewDescription,
		"metadata":    map[string]any{"new-key": "new-value"},
	}, recorder.updateBodies[0])
}

func TestUnitSelectelComputeVolumeV3OnlineResizeUsesMicroversion(t *testing.T) {
	remote := testComputeVolumeV3Remote{
		ID:               "volume-id",
		Status:           "in-use",
		Size:             10,
		Name:             "old-name",
		Description:      "old-description",
		AvailabilityZone: "ru-2a",
		VolumeType:       "fast.ru-2a",
		Metadata:         map[string]string{"old-key": "old-value"},
	}
	recorder := newTestComputeVolumeV3UpdateRecorder(t, remote)
	server := httptest.NewServer(recorder)
	defer server.Close()

	oldValues := testComputeVolumeV3UpdateValues()
	newValues := testComputeVolumeV3UpdateValues()
	oldValues["enable_online_resize"] = true
	newValues["enable_online_resize"] = true
	newValues["size"] = 20

	state, diagnostics := testApplyComputeVolumeV3Update(
		t,
		testComputeVolumeV3Client(server.URL),
		oldValues,
		newValues,
	)

	require.False(t, diagnostics.HasError(), diagnostics)
	require.NotNil(t, state)
	assert.Equal(t, "20", state.Attributes["size"])
	assert.Equal(t, []string{"size extend request"}, recorder.mutationPhases)
	require.Len(t, recorder.microversions, 1)
	assert.Equal(t, "volume 3.42", recorder.microversions[0])
	require.Len(t, recorder.actionBodies, 1)
	assert.Contains(t, recorder.actionBodies[0], "os-extend")
	assert.NotContains(t, recorder.actionBodies[0], "os-retype")
}

func TestUnitSelectelComputeVolumeV3UpdateRejectsInvalidResizeBeforeMutation(t *testing.T) {
	t.Run("shrink", func(t *testing.T) {
		recorder := newTestComputeVolumeV3UpdateRecorder(t, testComputeVolumeV3Remote{})
		server := httptest.NewServer(recorder)
		defer server.Close()

		oldValues := testComputeVolumeV3UpdateValues()
		newValues := testComputeVolumeV3UpdateValues()
		newValues["size"] = 5
		newValues["name"] = "must-not-be-applied"

		state, diagnostics := testApplyComputeVolumeV3Update(
			t,
			testComputeVolumeV3Client(server.URL),
			oldValues,
			newValues,
		)

		require.True(t, diagnostics.HasError())
		assert.Contains(t, diagnostics[0].Summary, "cannot shrink")
		require.NotNil(t, state)
		assert.Equal(t, "volume-id", state.ID)
		assert.Zero(t, recorder.requestCount)
		assert.Empty(t, recorder.mutationPhases)
	})

	t.Run("attached volume without opt-in", func(t *testing.T) {
		remote := testComputeVolumeV3Remote{
			ID:               "volume-id",
			Status:           "in-use",
			Size:             10,
			Name:             "old-name",
			Description:      "old-description",
			AvailabilityZone: "ru-2a",
			VolumeType:       "fast.ru-2a",
			Metadata:         map[string]string{"old-key": "old-value"},
		}
		recorder := newTestComputeVolumeV3UpdateRecorder(t, remote)
		server := httptest.NewServer(recorder)
		defer server.Close()

		oldValues := testComputeVolumeV3UpdateValues()
		newValues := testComputeVolumeV3UpdateValues()
		newValues["size"] = 20
		newValues["name"] = "must-not-be-applied"

		state, diagnostics := testApplyComputeVolumeV3Update(
			t,
			testComputeVolumeV3Client(server.URL),
			oldValues,
			newValues,
		)

		require.True(t, diagnostics.HasError())
		assert.Contains(t, diagnostics[0].Summary, "enable_online_resize is false")
		require.NotNil(t, state)
		assert.Equal(t, "volume-id", state.ID)
		assert.Equal(t, "old-name", state.Attributes["name"])
		assert.Equal(t, "10", state.Attributes["size"])
		assert.Equal(t, 1, recorder.requestCount)
		assert.Empty(t, recorder.mutationPhases)
	})
}

func TestUnitSelectelComputeVolumeV3UpdateReportsPartialFailures(t *testing.T) {
	testCases := []struct {
		failedPhase       string
		expectedCompleted string
		expectedRemote    testComputeVolumeV3Remote
	}{
		{
			failedPhase:       "descriptive update",
			expectedCompleted: "Completed phases: none",
			expectedRemote: testComputeVolumeV3Remote{
				Name: "old-name", Description: "old-description", Size: 10,
				Metadata: map[string]string{"old-key": "old-value"},
			},
		},
		{
			failedPhase:       "size wait",
			expectedCompleted: "Completed phases: descriptive update, size extend request",
			expectedRemote: testComputeVolumeV3Remote{
				Name: testComputeVolumeV3NewName, Description: testComputeVolumeV3NewDescription, Size: 20,
				Metadata: map[string]string{"new-key": "new-value"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.failedPhase, func(t *testing.T) {
			remote := testComputeVolumeV3Remote{
				ID:               "volume-id",
				Status:           "available",
				Size:             10,
				Name:             "old-name",
				Description:      "old-description",
				AvailabilityZone: "ru-2a",
				VolumeType:       "fast.ru-2a",
				Metadata:         map[string]string{"old-key": "old-value"},
			}
			recorder := newTestComputeVolumeV3UpdateRecorder(t, remote)
			recorder.failPhase = testCase.failedPhase
			server := httptest.NewServer(recorder)
			defer server.Close()

			oldValues := testComputeVolumeV3UpdateValues()
			newValues := testComputeVolumeV3UpdateValues()
			newValues["name"] = testComputeVolumeV3NewName
			newValues["description"] = testComputeVolumeV3NewDescription
			newValues["metadata"] = map[string]any{"new-key": "new-value"}
			newValues["size"] = 20

			state, diagnostics := testApplyComputeVolumeV3Update(
				t,
				testComputeVolumeV3Client(server.URL),
				oldValues,
				newValues,
			)

			require.True(t, diagnostics.HasError())
			assert.Contains(t, diagnostics[0].Summary, `phase "`+testCase.failedPhase+`"`)
			assert.Contains(t, diagnostics[0].Summary, testCase.expectedCompleted)
			assert.Contains(t, diagnostics[0].Summary, "Unfinished phases:")
			assert.Contains(t, diagnostics[0].Summary, "server_error")
			require.NotNil(t, state)
			assert.Equal(t, "volume-id", state.ID)
			assert.Equal(t, testCase.expectedRemote.Name, state.Attributes["name"])
			assert.Equal(t, testCase.expectedRemote.Description, state.Attributes["description"])
			assert.Equal(t, testCase.expectedRemote.Metadata["old-key"], state.Attributes["metadata.old-key"])
			assert.Equal(t, testCase.expectedRemote.Metadata["new-key"], state.Attributes["metadata.new-key"])
			assert.Equal(t, fmt.Sprint(testCase.expectedRemote.Size), state.Attributes["size"])
		})
	}
}

func TestUnitSelectelComputeVolumeV3UpdateRetryWaitsForAcceptedResize(t *testing.T) {
	for _, remoteSize := range []int{10, 20} {
		t.Run(fmt.Sprintf("remote size %d", remoteSize), func(t *testing.T) {
			remote := testComputeVolumeV3Remote{
				ID:               "volume-id",
				Status:           "extending",
				Size:             remoteSize,
				Name:             "old-name",
				Description:      "old-description",
				AvailabilityZone: "ru-2a",
				VolumeType:       "fast.ru-2a",
				Metadata:         map[string]string{"old-key": "old-value"},
			}
			recorder := newTestComputeVolumeV3UpdateRecorder(t, remote)
			recorder.finishExtending = true
			recorder.resizeTargetSize = 20
			server := httptest.NewServer(recorder)
			defer server.Close()

			oldValues := testComputeVolumeV3UpdateValues()
			newValues := testComputeVolumeV3UpdateValues()
			newValues["size"] = 20

			state, diagnostics := testApplyComputeVolumeV3Update(
				t,
				testComputeVolumeV3Client(server.URL),
				oldValues,
				newValues,
			)

			require.False(t, diagnostics.HasError(), diagnostics)
			require.NotNil(t, state)
			assert.Equal(t, "20", state.Attributes["size"])
			assert.Equal(t, 3, recorder.getCount)
			assert.Empty(t, recorder.mutationPhases)
			assert.Empty(t, recorder.actionBodies, "an accepted resize must not be submitted again")
		})
	}
}

func diffAttribute(diff *terraform.InstanceDiff, name string) (*terraform.ResourceAttrDiff, bool) {
	if diff == nil {
		return nil, false
	}

	attributeDiff, exists := diff.Attributes[name]

	return attributeDiff, exists
}

func testComputeVolumeV3PlanDiff(
	t *testing.T,
	stateOverrides map[string]any,
	configOverrides map[string]any,
) *terraform.InstanceDiff {
	t.Helper()

	stateValues := map[string]any{
		"project_id":        "project-id",
		"region":            "ru-1",
		"size":              10,
		"availability_zone": "ru-1a",
		"volume_type":       "fast.ru-1a",
		"metadata": map[string]any{
			"visible": "state",
		},
	}
	configValues := map[string]any{
		"project_id":        "project-id",
		"region":            "ru-1",
		"size":              10,
		"availability_zone": "ru-1a",
		"volume_type":       "fast.ru-1",
	}
	maps.Copy(stateValues, stateOverrides)
	maps.Copy(configValues, configOverrides)

	resourceSchema := resourceComputeVolumeV3Schema()
	stateData := schema.TestResourceDataRaw(t, resourceSchema, stateValues)
	stateData.SetId("volume-id")

	diff, err := resourceComputeVolumeV3().Diff(
		t.Context(),
		stateData.State(),
		terraform.NewResourceConfigRaw(configValues),
		nil,
	)
	require.NoError(t, err)

	return diff
}

func assertPlanRequiresReplacement(t *testing.T, diff *terraform.InstanceDiff, keyPrefix string) {
	t.Helper()

	require.NotNil(t, diff)
	matched := false
	for key, attributeDiff := range diff.Attributes {
		if key == keyPrefix || len(key) > len(keyPrefix) && key[:len(keyPrefix)] == keyPrefix {
			matched = true
			if attributeDiff.RequiresNew {
				return
			}
		}
	}

	require.True(t, matched, "expected diff for %q, got %#v", keyPrefix, diff.Attributes)
	t.Fatalf("expected replacement diff for %q, got %#v", keyPrefix, diff.Attributes)
}

func testComputeVolumeV3CreateValues() map[string]any {
	return map[string]any{
		"project_id":        testBlockStorageProjectID,
		"region":            testBlockStorageRegion,
		"size":              10,
		"name":              "volume-name",
		"description":       "volume-description",
		"availability_zone": "ru-2a",
		"volume_type":       "fast.ru-2",
		"metadata": map[string]any{
			"user-key": "user-value",
		},
	}
}

func testComputeVolumeV3ResourceData(
	t *testing.T,
	values map[string]any,
) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, resourceComputeVolumeV3Schema(), values)
}

func testComputeVolumeV3DeleteResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()

	resourceData := testComputeVolumeV3ResourceData(t, nil)
	resourceData.SetId("volume-id")

	return resourceData
}

func testComputeVolumeV3UpdateValues() map[string]any {
	return map[string]any{
		"project_id":           testBlockStorageProjectID,
		"region":               testBlockStorageRegion,
		"size":                 10,
		"enable_online_resize": false,
		"name":                 "old-name",
		"description":          "old-description",
		"availability_zone":    "ru-2a",
		"volume_type":          "fast.ru-2a",
		"metadata": map[string]any{
			"old-key": "old-value",
		},
	}
}

func testApplyComputeVolumeV3Update(
	t *testing.T,
	client *blockstorage.Client,
	oldValues map[string]any,
	newValues map[string]any,
) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()

	resourceData := testComputeVolumeV3ResourceData(t, oldValues)
	resourceData.SetId("volume-id")

	return testApplyComputeVolumeV3UpdateFromState(
		t,
		client,
		resourceData.State(),
		newValues,
	)
}

func testApplyComputeVolumeV3UpdateFromState(
	t *testing.T,
	client *blockstorage.Client,
	state *terraform.InstanceState,
	newValues map[string]any,
) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()

	testResource := &schema.Resource{
		Schema: resourceComputeVolumeV3Schema(),
		UpdateContext: func(
			ctx context.Context,
			d *schema.ResourceData,
			_ any,
		) diag.Diagnostics {
			return updateComputeVolumeV3(ctx, d, client)
		},
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(time.Second),
		},
	}

	diff, err := testResource.Diff(
		t.Context(),
		state,
		terraform.NewResourceConfigRaw(newValues),
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, diff)

	return testResource.Apply(t.Context(), state, diff, nil)
}

func testComputeVolumeV3Client(serverURL string) *blockstorage.Client {
	client, err := blockstorage.NewClient(blockstorage.Config{
		Endpoint: serverURL + testComputeVolumeV3EndpointPath(),
		Token:    testBlockStorageToken,
	})
	if err != nil {
		panic(err)
	}

	return client
}

func testComputeVolumeV3EndpointPath() string {
	return "/volumev3/" + testBlockStorageRegion + "/" + testBlockStorageProjectID
}

func testComputeVolumeV3CollectionPath() string {
	return testComputeVolumeV3EndpointPath() + "/volumes"
}

type testComputeVolumeV3 struct {
	ID       string
	Status   string
	Metadata map[string]string
}

type testComputeVolumeV3Remote struct {
	ID               string
	Status           string
	Size             int
	Name             string
	Description      string
	AvailabilityZone string
	VolumeType       string
	Metadata         map[string]string
}

type testComputeVolumeV3UpdateRecorder struct {
	t                *testing.T
	remote           testComputeVolumeV3Remote
	failPhase        string
	waitFailed       bool
	finishExtending  bool
	resizeTargetSize int
	requestCount     int
	getCount         int
	mutationPhases   []string
	microversions    []string
	updateBodies     []map[string]any
	actionBodies     []string
}

func newTestComputeVolumeV3UpdateRecorder(
	t *testing.T,
	remote testComputeVolumeV3Remote,
) *testComputeVolumeV3UpdateRecorder {
	t.Helper()

	return &testComputeVolumeV3UpdateRecorder{t: t, remote: remote}
}

func (r *testComputeVolumeV3UpdateRecorder) ServeHTTP(
	response http.ResponseWriter,
	request *http.Request,
) {
	r.requestCount++

	switch {
	case request.Method == http.MethodGet &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
		r.getCount++
		if r.finishExtending && r.remote.Status == "extending" && r.getCount > 1 {
			r.remote.Status = "available"
			r.remote.Size = r.resizeTargetSize
		}
		if r.failPhase == "size wait" &&
			!r.waitFailed &&
			r.remote.Size == 20 {
			r.waitFailed = true
			writeTestComputeVolumeV3Fault(
				r.t,
				response,
				http.StatusInternalServerError,
				"size wait failed",
			)

			return
		}

		writeTestComputeVolumeV3Remote(r.t, response, r.remote)
	case request.Method == http.MethodPut &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id":
		var requestBody struct {
			Volume map[string]any `json:"volume"`
		}
		require.NoError(r.t, json.NewDecoder(request.Body).Decode(&requestBody))
		require.NotEmpty(r.t, requestBody.Volume)

		phase := "descriptive update"
		r.mutationPhases = append(r.mutationPhases, phase)
		r.updateBodies = append(r.updateBodies, requestBody.Volume)
		if phase == r.failPhase {
			writeTestComputeVolumeV3Fault(
				r.t,
				response,
				http.StatusInternalServerError,
				phase+" failed",
			)

			return
		}

		if _, exists := requestBody.Volume["name"]; exists {
			r.remote.Name = requestBody.Volume["name"].(string)
		}
		if _, exists := requestBody.Volume["description"]; exists {
			r.remote.Description = requestBody.Volume["description"].(string)
		}
		if _, exists := requestBody.Volume["metadata"]; exists {
			rawMetadata := requestBody.Volume["metadata"].(map[string]any)
			r.remote.Metadata = make(map[string]string, len(rawMetadata))
			for key, value := range rawMetadata {
				r.remote.Metadata[key] = value.(string)
			}
		}
		writeTestComputeVolumeV3Remote(r.t, response, r.remote)
	case request.Method == http.MethodPost &&
		request.URL.Path == testComputeVolumeV3CollectionPath()+"/volume-id/action":
		var rawBody json.RawMessage
		require.NoError(r.t, json.NewDecoder(request.Body).Decode(&rawBody))
		body := string(rawBody)
		r.actionBodies = append(r.actionBodies, body)

		var requestBody struct {
			Extend struct {
				NewSize int `json:"new_size"`
			} `json:"os-extend"`
		}
		require.NoError(r.t, json.Unmarshal(rawBody, &requestBody))
		require.Positive(r.t, requestBody.Extend.NewSize)

		phase := "size extend request"
		r.mutationPhases = append(r.mutationPhases, phase)
		r.microversions = append(
			r.microversions,
			request.Header.Get("OpenStack-API-Version"),
		)
		if phase == r.failPhase {
			writeTestComputeVolumeV3Fault(
				r.t,
				response,
				http.StatusInternalServerError,
				phase+" failed",
			)

			return
		}

		r.remote.Size = requestBody.Extend.NewSize
		response.WriteHeader(http.StatusAccepted)
	default:
		http.NotFound(response, request)
	}
}

func writeTestComputeVolumeV3Remote(
	t *testing.T,
	response http.ResponseWriter,
	remote testComputeVolumeV3Remote,
) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"volume": map[string]any{
			"id":                remote.ID,
			"status":            remote.Status,
			"size":              remote.Size,
			"name":              remote.Name,
			"description":       remote.Description,
			"availability_zone": remote.AvailabilityZone,
			"volume_type":       remote.VolumeType,
			"metadata":          remote.Metadata,
			"attachments":       []any{},
		},
	}))
}

func writeTestComputeVolumeV3(
	t *testing.T,
	response http.ResponseWriter,
	statusCode int,
	id string,
	status string,
	metadata any,
) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"volume": map[string]any{
			"id":       id,
			"status":   status,
			"metadata": metadata,
		},
	}))
}

func writeTestComputeVolumeV3Page(
	t *testing.T,
	response http.ResponseWriter,
	volumes []testComputeVolumeV3,
	nextURL string,
) {
	t.Helper()

	items := make([]map[string]any, 0, len(volumes))
	for _, item := range volumes {
		items = append(items, map[string]any{
			"id":       item.ID,
			"status":   item.Status,
			"metadata": item.Metadata,
		})
	}

	body := map[string]any{"volumes": items}
	if nextURL != "" {
		body["volumes_links"] = []map[string]string{{"rel": "next", "href": nextURL}}
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(body))
}

func writeTestComputeVolumeV3Fault(
	t *testing.T,
	response http.ResponseWriter,
	status int,
	message string,
) {
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

func writeTestComputeVolumeV3OverQuota(t *testing.T, response http.ResponseWriter) {
	t.Helper()

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusRequestEntityTooLarge)
	require.NoError(t, json.NewEncoder(response).Encode(map[string]any{
		"overLimit": map[string]any{
			"code":       http.StatusRequestEntityTooLarge,
			"message":    "VolumeLimitExceeded: maximum number of volumes exceeded",
			"retryAfter": "0",
		},
	}))
}

func readTestComputeVolumeV3CreateToken(t *testing.T, request *http.Request) string {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
	volumeBody := body["volume"].(map[string]any)
	metadata := volumeBody["metadata"].(map[string]any)
	token := metadata[computeVolumeV3ReservedMetadataKey].(string)
	require.NotEmpty(t, token)

	return token
}

type testComputeVolumeV3CreateErrorClient struct {
	delegate  *http.Client
	createErr error
	token     *string
	postCount int
}

func (c *testComputeVolumeV3CreateErrorClient) Do(request *http.Request) (*http.Response, error) {
	if request.Method == http.MethodPost && request.URL.Path == testComputeVolumeV3CollectionPath() {
		c.postCount++

		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			return nil, err
		}
		volumeBody := body["volume"].(map[string]any)
		metadata := volumeBody["metadata"].(map[string]any)
		*c.token = metadata[computeVolumeV3ReservedMetadataKey].(string)

		return nil, c.createErr
	}

	return c.delegate.Do(request) //nolint:gosec // G704: test-only client forwards requests to its local httptest server.
}
