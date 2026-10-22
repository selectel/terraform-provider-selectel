package selectel

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	images "github.com/selectel/images-go/pkg/v1"
	"github.com/selectel/images-go/pkg/v1/image"
	"github.com/selectel/images-go/pkg/v1/member"
	"github.com/stretchr/testify/require"
)

const (
	testAccImagesRegion     = "ru-6"
	testAccImagesNamePrefix = "tf-acc-images"
	testAccImageResource    = "selectel_compute_image_v2.test"
)

// testAccImagesSetup returns the INFRA_PROJECT_ID project, in which the
// image tests run like the other acceptance tests. The import scope of the
// provider is the same for every test in the process, so the tests also need
// the INFRA_REGION pool.
func testAccImagesSetup(t *testing.T) string {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	testAccSelectelPreCheckWithAuth(t)
	testAccSelectelPreCheckWithProjectID(t)
	if region := os.Getenv("INFRA_REGION"); region != testAccImagesRegion {
		t.Fatalf("INFRA_REGION must be %s for image acceptance tests, got %q", testAccImagesRegion, region)
	}

	return os.Getenv("INFRA_PROJECT_ID")
}

func testAccImagesClient(ctx context.Context, projectID string) (*images.Client, error) {
	return testAccConfigFromEnv().ImagesClient(ctx, projectID, testAccImagesRegion)
}

// testAccImageFile is a small raw disk with random data, so every test
// uploads data with its own checksum.
func testAccImageFile(t *testing.T) string {
	t.Helper()
	data := make([]byte, 1<<20)
	_, err := rand.Read(data)
	require.NoError(t, err)
	file := filepath.Join(t.TempDir(), "disk.raw")
	require.NoError(t, os.WriteFile(file, data, 0o600))

	return file
}

// testAccCheckImageDestroyed accepts an image that the API still returns as
// deleted: the Image service keeps it readable after deletion.
func testAccCheckImageDestroyed(state *terraform.State) error {
	ctx := context.Background()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "selectel_compute_image_v2" {
			continue
		}
		client, err := testAccImagesClient(ctx, rs.Primary.Attributes["project_id"])
		if err != nil {
			return err
		}
		current, _, err := image.Get(ctx, client, rs.Primary.ID)
		switch {
		case images.IsKind(err, images.KindNotFound):
		case err != nil:
			return fmt.Errorf("read image %s: %w", rs.Primary.ID, err)
		case current.Status != image.StatusDeleted && current.Status != image.StatusPendingDelete:
			return fmt.Errorf("image %s still exists with status %s", rs.Primary.ID, current.Status)
		}
	}

	return nil
}

// testAccStoreImageID keeps the image ID, so a later step can check that an
// update did not replace the image.
func testAccStoreImageID(id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[testAccImageResource]
		if !ok {
			return fmt.Errorf("resource %s not found", testAccImageResource)
		}
		*id = rs.Primary.ID

		return nil
	}
}

func testAccImageImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            testAccImageResource,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"local_file_path", "image_source_url", "properties", "timeouts"},
	}
}

func testAccImageFromFileConfig(projectID, name, file, extra string) string {
	return fmt.Sprintf(`
resource "selectel_compute_image_v2" "test" {
  project_id       = %q
  region           = %q
  name             = %q
  container_format = "bare"
  disk_format      = "raw"
  local_file_path  = %q
%s
}
`, projectID, testAccImagesRegion, name, file, extra)
}

func TestAccSelectelComputeImageV2_Upload(t *testing.T) {
	projectID := testAccImagesSetup(t)
	name := acctest.RandomWithPrefix(testAccImagesNamePrefix)
	file := testAccImageFile(t)
	var id string
	lookups := fmt.Sprintf(`
data "selectel_compute_image_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
  name       = selectel_compute_image_v2.test.name
}

data "selectel_compute_image_ids_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
  name_regex = "^%[3]s$"
  tags       = ["b"]
}
`, projectID, testAccImagesRegion, name+"-updated")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccImageFromFileConfig(projectID, name, file, `
  min_ram_mb = 512
  tags       = ["a"]
  properties = {
    os_distro = "ubuntu"
  }`),
				Check: resource.ComposeTestCheckFunc(
					testAccStoreImageID(&id),
					resource.TestCheckResourceAttr(testAccImageResource, "status", "active"),
					resource.TestCheckResourceAttr(testAccImageResource, "size_bytes", "1048576"),
					// The API derives min_disk from the image data.
					resource.TestCheckResourceAttr(testAccImageResource, "min_disk_gb", "1"),
					resource.TestCheckResourceAttr(testAccImageResource, "min_ram_mb", "512"),
					resource.TestCheckResourceAttr(testAccImageResource, "owner", projectID),
					resource.TestCheckTypeSetElemAttr(testAccImageResource, "tags.*", "a"),
					resource.TestCheckResourceAttr(testAccImageResource, "properties.os_distro", "ubuntu"),
					resource.TestCheckResourceAttr(testAccImageResource, "service_properties.os_distro", "ubuntu"),
				),
			},
			{
				Config: testAccImageFromFileConfig(projectID, name+"-updated", file, `
  protected  = true
  tags       = ["b"]
  properties = {
    os_type = "linux"
  }`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(testAccImageResource, "id", &id),
					resource.TestCheckResourceAttr(testAccImageResource, "name", name+"-updated"),
					resource.TestCheckResourceAttr(testAccImageResource, "protected", "true"),
					resource.TestCheckResourceAttr(testAccImageResource, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(testAccImageResource, "tags.*", "b"),
					resource.TestCheckNoResourceAttr(testAccImageResource, "properties.os_distro"),
					resource.TestCheckNoResourceAttr(testAccImageResource, "service_properties.os_distro"),
					resource.TestCheckResourceAttr(testAccImageResource, "properties.os_type", "linux"),
				),
			},
			{
				Config: testAccImageFromFileConfig(projectID, name+"-updated", file, `
  tags       = ["b"]
  properties = {
    os_type = "linux"
  }`) + lookups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccImageResource, "protected", "false"),
					resource.TestCheckResourceAttrPair("data.selectel_compute_image_v2.test", "id", testAccImageResource, "id"),
					resource.TestCheckResourceAttr("data.selectel_compute_image_v2.test", "properties.os_type", "linux"),
					resource.TestCheckResourceAttr("data.selectel_compute_image_ids_v2.test", "ids.#", "1"),
					resource.TestCheckResourceAttrPair("data.selectel_compute_image_ids_v2.test", "ids.0", testAccImageResource, "id"),
				),
			},
			testAccImageImportStep(),
		},
	})
}

func TestAccSelectelComputeImageV2_WebDownloadInvalidURL(t *testing.T) {
	projectID := testAccImagesSetup(t)
	name := acctest.RandomWithPrefix(testAccImagesNamePrefix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroyed,
		Steps: []resource.TestStep{
			{
				Config:      testAccImageWebDownloadConfig(projectID, name, "https://"+testAccImagesNamePrefix+".invalid/disk.img"),
				ExpectError: regexp.MustCompile(`could not\s+download the image data: url_not_accessible`),
			},
		},
	})
}

func TestAccSelectelComputeImageV2_WebDownload(t *testing.T) {
	projectID := testAccImagesSetup(t)
	sourceURL := os.Getenv("SELECTEL_ACC_IMAGES_SOURCE_URL")
	if sourceURL == "" {
		t.Skip("SELECTEL_ACC_IMAGES_SOURCE_URL must be set to the URL of a small qcow2 image for the web-download test, " +
			"for example, https://cloud-images.ubuntu.com/minimal/releases/noble/release/ubuntu-24.04-minimal-cloudimg-amd64.img")
	}
	name := acctest.RandomWithPrefix(testAccImagesNamePrefix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccImageWebDownloadConfig(projectID, name, sourceURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccImageResource, "status", "active"),
					resource.TestCheckResourceAttr(testAccImageResource, "service_properties.os_distro", "ubuntu"),
				),
			},
			testAccImageImportStep(),
		},
	})
}

func testAccImageWebDownloadConfig(projectID, name, url string) string {
	return fmt.Sprintf(`
resource "selectel_compute_image_v2" "test" {
  project_id       = %q
  region           = %q
  name             = %q
  container_format = "bare"
  disk_format      = "qcow2"
  image_source_url = %q
  properties = {
    os_distro = "ubuntu"
  }
}
`, projectID, testAccImagesRegion, name, url)
}

func TestAccSelectelComputeImageAccessV2_Basic(t *testing.T) {
	ownerID := testAccImagesSetup(t)
	memberID := testAccExtraProject(t)
	name := acctest.RandomWithPrefix(testAccImagesNamePrefix)
	file := testAccImageFile(t)
	config := func(status string) string {
		return testAccImageFromFileConfig(ownerID, name, file, `  visibility = "shared"`) + fmt.Sprintf(`
resource "selectel_compute_image_access_v2" "test" {
  project_id = %[1]q
  region     = %[3]q
  image_id   = selectel_compute_image_v2.test.id
  member_id  = %[2]q
}

resource "selectel_compute_image_access_accept_v2" "test" {
  project_id = %[2]q
  region     = %[3]q
  image_id   = selectel_compute_image_access_v2.test.image_id
  status     = %[4]q
}

data "selectel_compute_image_v2" "shared" {
  project_id    = %[2]q
  region        = %[3]q
  name          = %[5]q
  visibility    = "shared"
  member_status = selectel_compute_image_access_accept_v2.test.status
}
`, ownerID, memberID, testAccImagesRegion, status, name)
	}
	access := "selectel_compute_image_access_v2.test"
	accept := "selectel_compute_image_access_accept_v2.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config("accepted"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accept, "status", "accepted"),
					resource.TestCheckResourceAttr(accept, "member_id", memberID),
					resource.TestCheckResourceAttrPair(accept, "id", access, "id"),
					resource.TestCheckResourceAttrPair("data.selectel_compute_image_v2.shared", "id", testAccImageResource, "id"),
				),
			},
			{
				Config: config("rejected"),
				Check:  resource.TestCheckResourceAttr(accept, "status", "rejected"),
			},
			{
				ResourceName:      access,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// The ID names the project that received access, so the import
				// does not depend on the provider project.
				ResourceName:      accept,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Destroying the access revokes it while the image stays.
				Config: testAccImageFromFileConfig(ownerID, name, file, `  visibility = "shared"`),
				Check:  testAccCheckImageNotShared(ownerID, memberID),
			},
		},
	})
}

func testAccCheckImageNotShared(ownerID, memberID string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()
		client, err := testAccImagesClient(ctx, ownerID)
		if err != nil {
			return err
		}
		imageID := state.RootModule().Resources[testAccImageResource].Primary.ID
		_, _, err = member.Get(ctx, client, imageID, memberID)
		if !images.IsKind(err, images.KindNotFound) {
			return fmt.Errorf("access of project %s to image %s is not revoked: %v", memberID, imageID, err)
		}

		return nil
	}
}
