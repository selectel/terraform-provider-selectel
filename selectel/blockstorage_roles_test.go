package selectel

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/stretchr/testify/require"
)

const (
	testAccComputeBlockStorageRoleResourceName = "selectel_compute_volume_v3.test"
	testAccComputeBlockStorageRoleLookupName   = "data.selectel_compute_volume_v3.viewer"
	testAccComputeVolumeTypeRoleLookupName     = "data.selectel_compute_volume_type_v3.viewer"

	testAccComputeBlockStorageViewerProcessEnv = "SELECTEL_ACC_BLOCKSTORAGE_VIEWER_PROCESS"
	testAccComputeBlockStorageViewerVolumeID   = "SELECTEL_ACC_BLOCKSTORAGE_VIEWER_VOLUME_ID"
	testAccComputeBlockStorageViewerVolumeName = "SELECTEL_ACC_BLOCKSTORAGE_VIEWER_VOLUME_NAME"
)

var testAccComputeBlockStorageRoleEnvironmentVariables = []string{
	"OS_AUTH_URL",
	"OS_REGION_NAME",
	"OS_DOMAIN_NAME",
	"OS_USERNAME",
	"OS_PASSWORD",
	"INFRA_PROJECT_ID",
	"INFRA_REGION",
	"SELECTEL_ACC_BLOCKSTORAGE_VIEWER_USER",
	"SELECTEL_ACC_BLOCKSTORAGE_VIEWER_PASSWORD",
}

// The provider keeps one configuration per process, so each role runs in its own process.
var testAccComputeBlockStorageRoleProviders = map[string]func() (*schema.Provider, error){
	"selectel": func() (*schema.Provider, error) {
		return Provider("test"), nil
	},
}

func TestAccSelectelComputeBlockStorageRoles(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for Block Storage role acceptance tests")
	}
	testAccComputeBlockStorageRolesPreCheck(t)

	initialName := acctest.RandomWithPrefix("tf-acc-selectel-volume-role")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccComputeBlockStorageRolesPreCheck(t) },
		ProviderFactories: testAccComputeBlockStorageRoleProviders,
		CheckDestroy:      testAccCheckComputeBlockStorageRolesDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeBlockStorageWriteRoleConfig(initialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccComputeBlockStorageRoleResourceName, "name", initialName),
					resource.TestCheckResourceAttr(testAccComputeBlockStorageRoleResourceName, "project_id", os.Getenv("INFRA_PROJECT_ID")),
					resource.TestCheckResourceAttr(testAccComputeBlockStorageRoleResourceName, "region", os.Getenv("INFRA_REGION")),
					testAccCheckComputeBlockStorageViewerProcess(t, initialName),
				),
			},
		},
	})
}

func TestAccSelectelComputeBlockStorageRolesViewerProcess(t *testing.T) {
	if os.Getenv(testAccComputeBlockStorageViewerProcessEnv) != "1" {
		t.Skip("the viewer role check runs from TestAccSelectelComputeBlockStorageRoles")
	}
	if os.Getenv("TF_ACC") == "" {
		t.Fatal("TF_ACC must be set for Block Storage role acceptance tests")
	}

	testAccComputeBlockStorageViewerPreCheck(t)

	volumeID := os.Getenv(testAccComputeBlockStorageViewerVolumeID)
	volumeName := os.Getenv(testAccComputeBlockStorageViewerVolumeName)
	if volumeID == "" || volumeName == "" {
		t.Fatal("the viewer role process requires the parent volume ID and name")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccComputeBlockStorageViewerPreCheck(t) },
		ProviderFactories: testAccComputeBlockStorageRoleProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeBlockStorageViewerRoleConfig(volumeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccComputeBlockStorageRoleLookupName, "volume_id", volumeID),
					resource.TestCheckResourceAttr(testAccComputeBlockStorageRoleLookupName, "name", volumeName),
					resource.TestCheckResourceAttrSet(testAccComputeVolumeTypeRoleLookupName, "id"),
					resource.TestCheckResourceAttr(testAccComputeVolumeTypeRoleLookupName, "volume_type_id", "default"),
				),
			},
		},
	})

	testAccCheckComputeBlockStorageViewerUpdateForbidden(t, volumeID, volumeName)
}

func testAccComputeBlockStorageRolesPreCheck(t *testing.T) {
	t.Helper()

	missing := testAccComputeBlockStorageRolesMissingEnvironment(os.Getenv)
	if len(missing) != 0 {
		t.Fatalf(
			"missing environment variables for Block Storage role acceptance tests: %s",
			strings.Join(missing, ", "),
		)
	}

	t.Setenv("TF_VAR_selectel_username", os.Getenv("OS_USERNAME"))
	t.Setenv("TF_VAR_selectel_password", os.Getenv("OS_PASSWORD"))
}

func testAccComputeBlockStorageViewerPreCheck(t *testing.T) {
	t.Helper()

	missing := testAccComputeBlockStorageRolesMissingEnvironment(os.Getenv)
	if len(missing) != 0 {
		t.Fatalf(
			"missing environment variables for Block Storage role acceptance tests: %s",
			strings.Join(missing, ", "),
		)
	}

	t.Setenv("TF_VAR_selectel_username", os.Getenv("SELECTEL_ACC_BLOCKSTORAGE_VIEWER_USER"))
	t.Setenv("TF_VAR_selectel_password", os.Getenv("SELECTEL_ACC_BLOCKSTORAGE_VIEWER_PASSWORD"))
}

func testAccComputeBlockStorageRolesMissingEnvironment(getenv func(string) string) []string {
	missing := make([]string, 0)
	for _, envName := range testAccComputeBlockStorageRoleEnvironmentVariables {
		if getenv(envName) == "" {
			missing = append(missing, envName)
		}
	}

	return missing
}

func testAccComputeBlockStorageWriteRoleConfig(name string) string {
	return testAccComputeBlockStorageRoleProviderConfig() + fmt.Sprintf(`
resource "selectel_compute_volume_v3" "test" {
  provider   = selectel
  project_id = %q
  region     = %q
  name       = %q
  size       = 1
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		name,
	)
}

func testAccComputeBlockStorageViewerRoleConfig(volumeID string) string {
	return testAccComputeBlockStorageRoleProviderConfig() + fmt.Sprintf(`
data "selectel_compute_volume_v3" "viewer" {
  provider   = selectel
  project_id = %q
  region     = %q
  volume_id  = %q
}

data "selectel_compute_volume_type_v3" "viewer" {
  provider       = selectel
  project_id     = %q
  region         = %q
  volume_type_id = "default"
}
`,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
		volumeID,
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
	)
}

func testAccComputeBlockStorageRoleProviderConfig() string {
	return fmt.Sprintf(`
variable "selectel_username" {
  type      = string
  sensitive = true
}

variable "selectel_password" {
  type      = string
  sensitive = true
}

provider "selectel" {
  auth_url    = %q
  auth_region = %q
  domain_name = %q
  username    = var.selectel_username
  password    = var.selectel_password
  project_id  = %q
  region      = %q
}
`,
		os.Getenv("OS_AUTH_URL"),
		os.Getenv("OS_REGION_NAME"),
		os.Getenv("OS_DOMAIN_NAME"),
		os.Getenv("INFRA_PROJECT_ID"),
		os.Getenv("INFRA_REGION"),
	)
}

func testAccCheckComputeBlockStorageViewerProcess(t *testing.T, volumeName string) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[testAccComputeBlockStorageRoleResourceName]
		if !ok {
			return fmt.Errorf("Block Storage role-test resource was not found in state")
		}
		if resourceState.Primary.ID == "" {
			return fmt.Errorf("Block Storage role-test resource has no ID")
		}

		command := exec.CommandContext( //nolint:gosec // The command is the current test binary, not user input.
			t.Context(),
			os.Args[0],
			"-test.run=^TestAccSelectelComputeBlockStorageRolesViewerProcess$",
			"-test.v",
			"-test.timeout=30m",
		)
		command.Env = append(
			os.Environ(),
			testAccComputeBlockStorageViewerProcessEnv+"=1",
			testAccComputeBlockStorageViewerVolumeID+"="+resourceState.Primary.ID,
			testAccComputeBlockStorageViewerVolumeName+"="+volumeName,
		)

		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("viewer role acceptance process failed: %w\n%s", err, output)
		}

		client, err := testAccComputeBlockStorageRoleWriteClient(t)
		if err != nil {
			return err
		}
		actual, _, err := volume.Get(t.Context(), client, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read the role-test volume after the viewer check: %w", err)
		}
		if actual.Name != volumeName {
			return fmt.Errorf(
				"viewer role changed Block Storage volume %s name from %q to %q",
				resourceState.Primary.ID,
				volumeName,
				actual.Name,
			)
		}

		return nil
	}
}

func testAccCheckComputeBlockStorageViewerUpdateForbidden(t *testing.T, volumeID, volumeName string) {
	t.Helper()

	require.NotNil(t, cfgSingletone, "the viewer provider configuration was not initialized")

	oldValues := map[string]any{
		"project_id": os.Getenv("INFRA_PROJECT_ID"),
		"region":     os.Getenv("INFRA_REGION"),
		"size":       1,
		"name":       volumeName,
	}
	newValues := map[string]any{
		"project_id": os.Getenv("INFRA_PROJECT_ID"),
		"region":     os.Getenv("INFRA_REGION"),
		"size":       1,
		"name":       volumeName + "-forbidden-update",
	}

	resourceData := schema.TestResourceDataRaw(t, resourceComputeVolumeV3Schema(), oldValues)
	resourceData.SetId(volumeID)
	testResource := resourceComputeVolumeV3()
	diff, err := testResource.Diff(
		t.Context(),
		resourceData.State(),
		terraform.NewResourceConfigRaw(newValues),
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, diff)

	updatedState, diagnostics := testResource.Apply(t.Context(), resourceData.State(), diff, cfgSingletone)
	require.True(t, diagnostics.HasError(), "viewer role unexpectedly updated the Block Storage volume")
	require.Regexp(
		t,
		regexp.MustCompile(`(?is)failed to update Block Storage volume.*(forbidden|403)`),
		diagnostics[0].Summary,
	)
	require.NotNil(t, updatedState)
	require.Equal(t, volumeID, updatedState.ID)
	require.Equal(t, volumeName, updatedState.Attributes["name"])
}

func testAccCheckComputeBlockStorageRolesDestroy(t *testing.T) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		client, err := testAccComputeBlockStorageRoleWriteClient(t)
		if err != nil {
			return err
		}
		for _, resourceState := range state.RootModule().Resources {
			if resourceState.Type != "selectel_compute_volume_v3" {
				continue
			}
			if err := testAccVerifyComputeVolumeV3Deleted(
				t.Context(),
				client,
				resourceState.Primary.ID,
			); err != nil {
				return fmt.Errorf(
					"failed to verify role-test volume %s deletion: %w",
					resourceState.Primary.ID,
					err,
				)
			}
		}

		return nil
	}
}

func testAccComputeBlockStorageRoleWriteClient(t *testing.T) (*blockstorage.Client, error) {
	t.Helper()

	resourceData := schema.TestResourceDataRaw(t, resourceComputeVolumeV3Schema(), map[string]any{
		"project_id": os.Getenv("INFRA_PROJECT_ID"),
		"region":     os.Getenv("INFRA_REGION"),
		"size":       1,
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
	if diagnostics.HasError() {
		return nil, fmt.Errorf("failed to create Block Storage write-role client: %s", diagnostics[0].Summary)
	}

	return client, nil
}
