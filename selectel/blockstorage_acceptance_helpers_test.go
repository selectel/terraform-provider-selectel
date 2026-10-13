package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

const (
	testAccBlockStorageRegion = "ru-6"
	testAccBlockStorageZone   = "ru-6b"
	testAccBlockStorageType   = "universal2.ru-6b"
)

// testAccBlockStorageSetup runs the tests in the INFRA_PROJECT_ID project like
// the other acceptance tests. The framework provider uses the SDKv2 Config,
// which keeps the import scope of the first configuration in the process, so
// a project of its own would send the imports of later tests there.
func testAccBlockStorageSetup(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	testAccSelectelPreCheckWithAuth(t)
	testAccSelectelPreCheckWithProjectID(t)
	if region := os.Getenv("INFRA_REGION"); region != testAccBlockStorageRegion {
		t.Fatalf("INFRA_REGION must be %s for Block Storage acceptance tests, got %q",
			testAccBlockStorageRegion, region)
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
		testAccBlockStorageZone,
		testAccBlockStorageType,
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

  timeouts {
    create = "30m"
    update = "31m"
    delete = "32m"
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
  source_volume_id      = selectel_compute_volume_v3.source.id
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		testAccBlockStorageZone,
		testAccBlockStorageType,
		sourceName,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		testAccBlockStorageZone,
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
		testAccBlockStorageZone,
		testAccBlockStorageType,
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

		client, err := testAccBlockStorageClient(t, projectID, region)
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

	return testAccBlockStorageClient(
		t,
		resourceState.Primary.Attributes["project_id"],
		resourceState.Primary.Attributes["region"],
	)
}

func testAccConfigFromEnv() *Config {
	return &Config{
		AuthURL:        os.Getenv("OS_AUTH_URL"),
		AuthRegion:     os.Getenv("OS_REGION_NAME"),
		DomainName:     os.Getenv("OS_DOMAIN_NAME"),
		UserDomainName: os.Getenv("OS_USER_DOMAIN_NAME"),
		Username:       os.Getenv("OS_USERNAME"),
		Password:       os.Getenv("OS_PASSWORD"),
		UserAgent:      "terraform-provider-selectel/acceptance-tests",
	}
}

func testAccBlockStorageClient(t *testing.T, projectID, region string) (*blockstorage.Client, error) {
	t.Helper()

	return testAccConfigFromEnv().BlockStorageClient(t.Context(), projectID, region)
}

// testAccOpenStackServiceClient is for fixture operations that blockstorage-go
// does not implement, such as snapshot writes and Glance images.
func testAccOpenStackServiceClient(projectID, region, serviceType string) (*gophercloud.ServiceClient, error) {
	config := testAccConfigFromEnv()
	endpoint, token, err := config.projectEndpoint(context.Background(), projectID, region, serviceType)
	if err != nil {
		return nil, err
	}
	if endpoint == "" {
		return nil, fmt.Errorf("%s endpoint is empty for region %s", serviceType, region)
	}
	providerClient := &gophercloud.ProviderClient{TokenID: token}
	providerClient.HTTPClient.Timeout = 60 * time.Second
	providerClient.UserAgent.Prepend(config.UserAgent)

	return &gophercloud.ServiceClient{
		ProviderClient: providerClient,
		Endpoint:       gophercloud.NormalizeURL(endpoint),
		Type:           serviceType,
	}, nil
}
