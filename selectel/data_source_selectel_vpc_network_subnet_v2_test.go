package selectel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCNetworkSubnetV2DataSourceName = "data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1"

func TestVPCNetworkSubnetV2DataSourceSchema(t *testing.T) {
	dataSource := dataSourceVPCNetworkSubnetV2()

	_, hasDeprecatedDHCPSelector := dataSource.Schema["dhcp_enabled"]
	assert.False(t, hasDeprecatedDHCPSelector)

	enableDHCP := dataSource.Schema["enable_dhcp"]
	require.NotNil(t, enableDHCP)
	assert.True(t, enableDHCP.Optional)
	assert.True(t, enableDHCP.Computed)

	_, hasDeprecatedAllocationPools := dataSource.Schema["allocation_pools"]
	assert.False(t, hasDeprecatedAllocationPools)

	allocationPool := dataSource.Schema["allocation_pool"]
	require.NotNil(t, allocationPool)
	assert.Equal(t, schema.TypeSet, allocationPool.Type)
	assert.True(t, allocationPool.Computed)
}

func TestAccVPCNetworkSubnetV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "name", subnetName),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "cidr", "192.168.199.0/24"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "gateway_ip", "192.168.199.1"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "ip_version", "4"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "enable_dhcp", "false"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "dns_nameservers.#", "2"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "allocation_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(testAccVPCNetworkSubnetV2DataSourceName, "allocation_pool.*", map[string]string{
						"start": "192.168.199.10",
						"end":   "192.168.199.20",
					}),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2DataSourceName, "all_tags.#", "1"),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2DataSourceName, "subnet_id",
						testAccVPCNetworkSubnetV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2DataSourceName, "network_id",
						testAccVPCNetworkSubnetV2ResourceName, "network_id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2DataSourceName, "service_types.#",
						testAccVPCNetworkSubnetV2ResourceName, "service_types.#",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2DataSourceName, "ipv6_address_mode",
						testAccVPCNetworkSubnetV2ResourceName, "ipv6_address_mode",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2DataSourceName, "ipv6_ra_mode",
						testAccVPCNetworkSubnetV2ResourceName, "ipv6_ra_mode",
					),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.#", "1"),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.0.destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.0.next_hop", "192.168.199.10"),
					resource.TestCheckResourceAttrPair(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "blocked",
						testAccVPCNetworkSubnetV2ResourceName, "blocked",
					),
				),
			},
			{
				Config: testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "host_routes.#", "1"),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2ResourceName, "host_routes.#",
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.#",
					),
				),
			},
			{
				Config:      testAccVPCNetworkSubnetV2DataSourceAmbiguous(projectID, region, subnetName),
				ExpectError: regexp.MustCompile(`2 VPC network subnets found`),
			},
			{
				Config:      testAccVPCNetworkSubnetV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC network subnet found`),
			},
		},
	})
}

func testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_route_v2" "route_tf_acc_test_1" {
  project_id       = local.project_id
  region           = "%s"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
}
data "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%[2]s"
  name        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.name
  description = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.description
  network_id  = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.cidr
  gateway_ip  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.gateway_ip
  ip_version  = 4
  enable_dhcp = false
  tags        = ["tf-acc"]

  depends_on = [selectel_vpc_network_subnet_route_v2.route_tf_acc_test_1]
}
`, region, region)
}

func testAccVPCNetworkSubnetV2DataSourceAmbiguous(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
  name       = "%[2]s"
  tags       = ["%[2]s"]
}
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.200.0/24"
  name       = "%[2]s-twin"
  tags       = ["%[2]s"]
}

data "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  tags       = ["%[2]s"]

  depends_on = [
    selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1,
    selectel_vpc_network_subnet_v2.subnet_tf_acc_test_2,
  ]
}
`, region, subnetName)
}

func testAccVPCNetworkSubnetV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
data "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-subnet"
}
`, region)
}
