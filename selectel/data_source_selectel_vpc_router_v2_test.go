package selectel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccVPCRouterV2DataSourceName = "data.selectel_vpc_router_v2.router_tf_acc_test_1"

func TestAccVPCRouterV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	routerName := acctest.RandomWithPrefix("tf-acc-router")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2DataSource(projectID, region, routerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2DataSourceName, "router_id",
						testAccVPCRouterV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "name", routerName),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2DataSourceName, "status",
						testAccVPCRouterV2ResourceName, "status",
					),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "external_network_id", ""),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "external_fixed_ip.#", "0"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "routes.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "routes.0.destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "routes.0.next_hop", "192.168.199.10"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2DataSourceName, "all_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(testAccVPCRouterV2DataSourceName, "all_tags.*", routerName),
					resource.TestCheckResourceAttrSet(testAccVPCRouterV2DataSourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2DataSourceName, "availability_zone_hints.#",
						testAccVPCRouterV2ResourceName, "availability_zone_hints.#",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2DataSourceName, "blocked",
						testAccVPCRouterV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccVPCRouterV2DataSourceAmbiguous(projectID, region, routerName),
				ExpectError: regexp.MustCompile(`2 VPC routers found`),
			},
			{
				Config:      testAccVPCRouterV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC router found`),
			},
		},
	})
}

func testAccVPCRouterV2DataSource(projectID, region, routerName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}

resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%[1]s"
  name           = "%[2]s"
  description    = "created by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc", "%[2]s"]
}

resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}

resource "selectel_vpc_router_routes_v2" "routes_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id

  routes {
    destination_cidr = "10.0.0.0/8"
    next_hop         = "192.168.199.10"
  }

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}

data "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%[1]s"
  tags           = ["%[2]s"]
  admin_state_up = true
  description    = selectel_vpc_router_v2.router_tf_acc_test_1.description

  depends_on = [selectel_vpc_router_routes_v2.routes_tf_acc_test_1]
}
`, region, routerName)
}

func testAccVPCRouterV2DataSourceAmbiguous(projectID, region, routerName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s"
  tags       = ["%[2]s"]
}

resource "selectel_vpc_router_v2" "router_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s-twin"
  tags       = ["%[2]s"]
}

data "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  tags       = ["%[2]s"]

  depends_on = [
    selectel_vpc_router_v2.router_tf_acc_test_1,
    selectel_vpc_router_v2.router_tf_acc_test_2,
  ]
}
`, region, routerName)
}

func testAccVPCRouterV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
data "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-router"
}
`, region)
}
