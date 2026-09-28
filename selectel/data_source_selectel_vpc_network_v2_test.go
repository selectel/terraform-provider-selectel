package selectel

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

const testAccVPCNetworkV2DataSourceName = "data.selectel_vpc_network_v2.network_tf_acc_test_1"

func TestFilterVPCNetworkV2BySubnetCIDR(t *testing.T) {
	client, err := vpc.NewClient(vpc.Config{
		Endpoint: "https://vpc.example.test",
		Token:    "token",
		HTTPClient: &http.Client{
			Transport: httptest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "/v2.0/subnets", req.URL.Path)
				assert.Equal(t, "192.168.199.0/24", req.URL.Query().Get("cidr"))

				return httptest.NewFakeResponse(http.StatusOK, `{"subnets": [
					{"id": "subnet-1", "network_id": "net-1"},
					{"id": "subnet-2", "network_id": "net-3"},
					{"id": "subnet-3", "network_id": "net-unknown"}
				]}`), nil
			}),
		},
	})
	require.NoError(t, err)

	candidates := []network.Network{{ID: "net-1"}, {ID: "net-2"}, {ID: "net-3"}}

	filtered, err := filterVPCNetworkV2BySubnetCIDR(context.Background(), client, candidates, "192.168.199.0/24")

	require.NoError(t, err)
	assert.Equal(t, []network.Network{{ID: "net-1"}, {ID: "net-3"}}, filtered,
		"only networks that own a subnet with the CIDR stay, in the candidate order")
}

func TestAccVPCNetworkV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkV2DataSource(projectID, region, networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "network_id",
						testAccVPCNetworkV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "name", networkName),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "external", "false"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "subnets.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2DataSourceName, "all_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(testAccVPCNetworkV2DataSourceName, "all_tags.*", networkName),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "mtu",
						testAccVPCNetworkV2ResourceName, "mtu",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "status",
						testAccVPCNetworkV2ResourceName, "status",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "dns_domain",
						testAccVPCNetworkV2ResourceName, "dns_domain",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "blocked",
						testAccVPCNetworkV2ResourceName, "blocked",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "shared",
						testAccVPCNetworkV2ResourceName, "shared",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "port_security_enabled",
						testAccVPCNetworkV2ResourceName, "port_security_enabled",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "is_public",
						testAccVPCNetworkV2ResourceName, "is_public",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "is_dns_enabled",
						testAccVPCNetworkV2ResourceName, "is_dns_enabled",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkV2DataSourceName, "availability_zone_hints.#",
						testAccVPCNetworkV2ResourceName, "availability_zone_hints.#",
					),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2DataSourceName, "availability_zones.#"),
				),
			},
			{
				Config:      testAccVPCNetworkV2DataSourceAmbiguous(projectID, region, networkName),
				ExpectError: regexp.MustCompile(`2 VPC networks found`),
			},
			{
				Config:      testAccVPCNetworkV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC network found`),
			},
		},
	})
}

func testAccVPCNetworkV2DataSource(projectID, region, networkName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%[1]s"
  name           = "%[2]s"
  description    = "created by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc", "%[2]s"]
}
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}

data "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%[1]s"
  tags                 = ["%[2]s"]
  mtu                  = selectel_vpc_network_v2.network_tf_acc_test_1.mtu
  status               = selectel_vpc_network_v2.network_tf_acc_test_1.status
  external             = false
  matching_subnet_cidr = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.cidr
}
`, region, networkName)
}

func testAccVPCNetworkV2DataSourceAmbiguous(projectID, region, networkName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s"
  tags       = ["%[2]s"]
}
resource "selectel_vpc_network_v2" "network_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s-twin"
  tags       = ["%[2]s"]
}

data "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  tags       = ["%[2]s"]

  depends_on = [
    selectel_vpc_network_v2.network_tf_acc_test_1,
    selectel_vpc_network_v2.network_tf_acc_test_2,
  ]
}
`, region, networkName)
}

func testAccVPCNetworkV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
data "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-network"
}
`, region)
}
