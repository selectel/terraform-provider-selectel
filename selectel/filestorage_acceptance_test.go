package selectel

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	filestorage "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/shareaccess"
	"github.com/selectel/filestorage-go/pkg/v1/sharenetwork"
	"github.com/selectel/filestorage-go/pkg/v1/sharetype"
)

// testAccFileStorageRegions are the production pools whose File Storage API
// serves microversion 2.75.
var testAccFileStorageRegions = []string{
	"gis-1", "gis-2", "ke-1", "kz-1", "ru-1", "ru-2", "ru-3", "ru-6", "ru-7", "ru-8", "ru-9", "uz-1", "uz-2",
}

const testAccFileStorageShare = "selectel_filestorage_share_v2.test"

// testAccFileStorageSetup checks the environment before the tests create
// anything and returns the share type to use: the first one available to the
// project in INFRA_REGION.
func testAccFileStorageSetup(t *testing.T) string {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	testAccSelectelPreCheckWithAuth(t)
	testAccSelectelPreCheckWithProjectID(t)
	region := os.Getenv("INFRA_REGION")
	if !slices.Contains(testAccFileStorageRegions, region) {
		t.Skipf("INFRA_REGION %q has no File Storage with API microversion 2.75; use one of %v",
			region, testAccFileStorageRegions)
	}
	client := testAccFileStorageClient(t)
	types, _, err := sharetype.List(t.Context(), client)
	if err != nil {
		t.Fatalf("list file storage types: %v", err)
	}
	if len(types) == 0 {
		t.Skipf("no file storage type is available to project %s in %s", os.Getenv("INFRA_PROJECT_ID"), region)
	}
	names := make([]string, 0, len(types))
	for _, typ := range types {
		names = append(names, typ.Name)
	}
	slices.Sort(names)

	return names[0]
}

func testAccFileStorageClient(t *testing.T) *filestorage.Client {
	t.Helper()
	client, err := testAccConfigFromEnv().FileStorageClient(t.Context(), os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"))
	if err != nil {
		t.Fatalf("connect to File Storage: %v", err)
	}

	return client
}

// testAccFileStorageNetwork is a private network with one subnet for the share.
func testAccFileStorageNetwork(name, cidr string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_network_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
  name       = %[3]q
}

resource "selectel_vpc_network_subnet_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
  network_id = selectel_vpc_network_v2.test.id
  cidr       = %[4]q
}

`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), name, cidr)
}

// testAccShareSize is the minimum size of the basic file storage types, in GB.
const testAccShareSize = 50

type testAccShare struct {
	name, proto, shareType, ip, owner string
	size                              int
}

func testAccFileStorageShareConfig(network string, s testAccShare) string {
	ip := ""
	if s.ip != "" {
		ip = fmt.Sprintf("ip = %q", s.ip)
	}

	return network + fmt.Sprintf(`
resource "selectel_filestorage_share_v2" "test" {
  project_id  = %q
  region      = %q
  name        = %q
  share_proto = %q
  share_type  = %q
  size        = %d
  network_id  = selectel_vpc_network_v2.test.id
  subnet_id   = selectel_vpc_network_subnet_v2.test.id
  %s

  metadata = {
    owner = %q
  }
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), s.name, s.proto, s.shareType, s.size, ip, s.owner)
}

func testAccFileStorageAccessConfig(name, accessTo, level string) string {
	return fmt.Sprintf(`
resource "selectel_filestorage_share_access_v2" %q {
  project_id   = %q
  region       = %q
  share_id     = selectel_filestorage_share_v2.test.id
  access_to    = %q
  access_level = %q
}
`, name, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"), accessTo, level)
}

func testAccFileStorageDataConfig() string {
	return fmt.Sprintf(`
data "selectel_filestorage_share_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
  name       = selectel_filestorage_share_v2.test.name
}

data "selectel_filestorage_share_v2" "by_id" {
  project_id = %[1]q
  region     = %[2]q
  share_id   = selectel_filestorage_share_v2.test.id
}

data "selectel_filestorage_availability_zones_v2" "test" {
  project_id = %[1]q
  region     = %[2]q
}
`, os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION"))
}

func TestAccSelectelFileStorageShareV2_NFS(t *testing.T) {
	shareType := testAccFileStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-filestorage-nfs")
	network := testAccFileStorageNetwork(name, "192.168.211.0/24")
	created := testAccShare{name: name, proto: "NFS", shareType: shareType, size: testAccShareSize, owner: "initial"}
	updated := created
	updated.name, updated.owner, updated.size = name+"-updated", "updated", testAccShareSize+10
	rules := testAccFileStorageAccessConfig("ro", "192.168.211.0/24", "ro") +
		testAccFileStorageAccessConfig("any", "0.0.0.0/0", "rw")
	var observed share.View

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFileStorageDestroyed(t, &observed),
		Steps: []resource.TestStep{
			{
				Config: testAccFileStorageShareConfig(network, created) + rules + testAccFileStorageDataConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileStorageShareExists(t, &observed),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "status", "available"),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "share_type", shareType),
					resource.TestCheckResourceAttrSet(testAccFileStorageShare, "ip"),
					resource.TestCheckResourceAttrSet(testAccFileStorageShare, "export_locations.0.path"),
					resource.TestCheckResourceAttrSet(testAccFileStorageShare, "availability_zone"),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "metadata.%", "1"),
					resource.TestCheckResourceAttrSet(testAccFileStorageShare, "all_metadata.delete_empty_share_network"),
					resource.TestCheckResourceAttr("selectel_filestorage_share_access_v2.ro", "state", "active"),
					resource.TestCheckResourceAttr("selectel_filestorage_share_access_v2.any", "state", "active"),
					resource.TestCheckResourceAttrPair("data.selectel_filestorage_share_v2.test", "id", testAccFileStorageShare, "id"),
					resource.TestCheckResourceAttrPair("data.selectel_filestorage_share_v2.test", "ip", testAccFileStorageShare, "ip"),
					resource.TestCheckResourceAttrPair("data.selectel_filestorage_share_v2.test", "subnet_id", testAccFileStorageShare, "subnet_id"),
					resource.TestCheckResourceAttrPair("data.selectel_filestorage_share_v2.by_id", "name", testAccFileStorageShare, "name"),
					resource.TestCheckResourceAttrPair("data.selectel_filestorage_share_v2.by_id", "export_locations.0.path", testAccFileStorageShare, "export_locations.0.path"),
					resource.TestCheckResourceAttrSet("data.selectel_filestorage_availability_zones_v2.test", "names.0"),
				),
			},
			{
				Config: testAccFileStorageShareConfig(network, updated) + rules,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileStorageShareExists(t, &observed),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "name", name+"-updated"),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "size", strconv.Itoa(testAccShareSize+10)),
					resource.TestCheckResourceAttr(testAccFileStorageShare, "metadata.owner", "updated"),
				),
			},
			{
				ResourceName:            testAccFileStorageShare,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata", "timeouts"},
			},
			{
				ResourceName:            "selectel_filestorage_share_access_v2.ro",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccFileStorageAccessImportID("selectel_filestorage_share_access_v2.ro"),
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func TestAccSelectelFileStorageShareV2_CIFS(t *testing.T) {
	shareType := testAccFileStorageSetup(t)
	name := acctest.RandomWithPrefix("tf-acc-filestorage-cifs")
	network := testAccFileStorageNetwork(name, "192.168.212.0/24")
	created := testAccShare{
		name: name, proto: "CIFS", shareType: shareType, size: testAccShareSize, owner: "initial", ip: "192.168.212.50",
	}
	var observed share.View

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFileStorageDestroyed(t, &observed),
		Steps: []resource.TestStep{{
			Config: testAccFileStorageShareConfig(network, created) +
				testAccFileStorageAccessConfig("rw", "192.168.212.0/24", "rw"),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckFileStorageShareExists(t, &observed),
				resource.TestCheckResourceAttr(testAccFileStorageShare, "ip", "192.168.212.50"),
				resource.TestCheckResourceAttr(testAccFileStorageShare, "share_proto", "CIFS"),
				resource.TestCheckResourceAttr("selectel_filestorage_share_access_v2.rw", "state", "active"),
			),
		}},
	})
}

// testAccCheckFileStorageShareExists reads the share through the SDK and keeps
// it, so the destroy check knows its service share network.
func testAccCheckFileStorageShareExists(t *testing.T, observed *share.View) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[testAccFileStorageShare]
		if !ok {
			return fmt.Errorf("%s is not in state", testAccFileStorageShare)
		}
		current, _, err := share.Get(t.Context(), testAccFileStorageClient(t), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("read file storage %s: %w", resourceState.Primary.ID, err)
		}
		*observed = *current

		return nil
	}
}

func testAccFileStorageAccessImportID(name string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("%s is not in state", name)
		}

		return resourceState.Primary.Attributes["share_id"] + "/" + resourceState.Primary.ID, nil
	}
}

// testAccCheckFileStorageDestroyed confirms that the shares, their access
// rules and the service share network of the observed share are gone.
func testAccCheckFileStorageDestroyed(t *testing.T, observed *share.View) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		client := testAccFileStorageClient(t)
		var leftovers []error
		for _, resourceState := range state.RootModule().Resources {
			var err error
			switch resourceState.Type {
			case "selectel_filestorage_share_v2":
				_, _, err = share.Get(t.Context(), client, resourceState.Primary.ID)
			case "selectel_filestorage_share_access_v2":
				_, _, err = shareaccess.Get(t.Context(), client, resourceState.Primary.ID)
			default:
				continue
			}
			if !filestorage.IsKind(err, filestorage.KindNotFound) {
				leftovers = append(leftovers, fmt.Errorf("%s %s still exists: %v", resourceState.Type, resourceState.Primary.ID, err))
			}
		}
		if observed.ShareNetworkID != "" {
			_, _, err := sharenetwork.Get(t.Context(), client, observed.ShareNetworkID)
			if !filestorage.IsKind(err, filestorage.KindNotFound) {
				leftovers = append(leftovers, fmt.Errorf("share network %s still exists: %v", observed.ShareNetworkID, err))
			}
		}

		return errors.Join(leftovers...)
	}
}
