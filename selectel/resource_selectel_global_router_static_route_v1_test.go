package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

const resourceGlobalRouterStaticRouteName = "selectel_global_router_static_route_v1.static_route_tf_acc_test_1"

func TestAccGlobalRouterStaticRouteV1Basic(t *testing.T) {
	var staticRoute globalrouter.StaticRoute
	staticRouteName := acctest.RandomWithPrefix("tf-acc") + "_static_route"
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subent"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
			testAccGlobalRouterStaticRoutePreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterStaticRouteV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterStaticRouteV1Basic(
					routerName, networkName, subnetName, staticRouteName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterStaticRouteV1Exists(resourceGlobalRouterStaticRouteName, &staticRoute),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "name", staticRouteName),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "cidr", globalRouterStaticRouteCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "next_hop", globalRouterNextHop),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "netops_static_route_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "subnet_id"),
				),
			},
		},
	})
}

func TestAccGlobalRouterStaticRouteV1Update(t *testing.T) {
	var staticRoute globalrouter.StaticRoute
	staticRouteName := acctest.RandomWithPrefix("tf-acc") + "_static_route"
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subent"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
			testAccGlobalRouterStaticRoutePreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterStaticRouteV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterStaticRouteV1Basic(
					routerName, networkName, subnetName, staticRouteName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterStaticRouteV1Exists(resourceGlobalRouterStaticRouteName, &staticRoute),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "name", staticRouteName),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "cidr", globalRouterStaticRouteCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "next_hop", globalRouterNextHop),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "netops_static_route_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "subnet_id"),
				),
			},
			{
				Config: testAccGlobalRouterStaticRouteV1WithTags(
					routerName, networkName, subnetName, staticRouteName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterStaticRouteV1Exists(resourceGlobalRouterStaticRouteName, &staticRoute),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "name", staticRouteName),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "cidr", globalRouterStaticRouteCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "next_hop", globalRouterNextHop),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "netops_static_route_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "subnet_id"),
				),
			},
			{
				Config: testAccGlobalRouterStaticRouteV1Basic(
					routerName, networkName, subnetName, staticRouteName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterStaticRouteV1Exists(resourceGlobalRouterStaticRouteName, &staticRoute),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "name", staticRouteName),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "cidr", globalRouterStaticRouteCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "next_hop", globalRouterNextHop),
					resource.TestCheckResourceAttr(resourceGlobalRouterStaticRouteName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "netops_static_route_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterStaticRouteName, "subnet_id"),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterStaticRouteV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test static route object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_static_route_v1" {
			continue
		}

		_, _, err := globalrouterClient.StaticRoute(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("static route still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterStaticRouteV1Exists(n string, staticRoute *globalrouter.StaticRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
		if diagErr != nil {
			return fmt.Errorf("can't get global_router client for test static route object")
		}

		u, _, err := globalrouterClient.StaticRoute(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("static_route not found")
		}

		*staticRoute = *u

		return nil
	}
}

func testAccGlobalRouterStaticRouteV1Basic(
	routerName string, networkName string, subnetName string, staticRouteName string,
) string {
	return fmt.Sprintf(`
provider openstack {
	tenant_id = "%s"
}
resource "openstack_networking_network_v2" "os_network_tf_acc_test_1" {
 	region = "%s"
  	name   = "network_one"
}
resource "openstack_networking_subnet_v2" "os_subnet_tf_acc_test_1" {
  network_id  = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  cidr        = "%s"
  ip_version  = 4
  enable_dhcp = false
  name        = "subnet"
}

data "selectel_global_router_zone_v1" "cloud_zone" {
  name    = "%s"
  service = "vpc"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_cloud_network_v1" "network_tf_acc_test_1" {
  router_id     = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id       = data.selectel_global_router_zone_v1.cloud_zone.id
  os_network_id = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  project_id    = "%s"
  name          = "%s"
}
resource "selectel_global_router_cloud_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_cloud_network_v1.network_tf_acc_test_1.id
  os_subnet_id      = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.id
  cidr              = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.cidr
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
}
resource "selectel_global_router_static_route_v1" "static_route_tf_acc_test_1" {
  router_id = selectel_global_router_router_v1.router_tf_acc_test_1.id
  cidr      = "%s"
  next_hop  = "%s"
  name      = "%s"
  # explicit dependency, because next_hop should be taken from subnet_tf_acc_test_1
  depends_on = [
    selectel_global_router_cloud_subnet_v1.subnet_tf_acc_test_1
  ]
}`, globalRouterVPCProjectID, globalRouterVPCRegion, globalRouterSubnetCidr,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName,
		globalRouterSubnetGateway, globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName,
		globalRouterStaticRouteCidr, globalRouterNextHop, staticRouteName)
}

func testAccGlobalRouterStaticRouteV1WithTags(
	routerName string, networkName string, subnetName string, staticRouteName string,
) string {
	return fmt.Sprintf(`
provider openstack {
	tenant_id = "%s"
}
resource "openstack_networking_network_v2" "os_network_tf_acc_test_1" {
 	region = "%s"
  	name   = "network_one"
}
resource "openstack_networking_subnet_v2" "os_subnet_tf_acc_test_1" {
  network_id  = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  cidr        = "%s"
  ip_version  = 4
  enable_dhcp = false
  name        = "subnet"
}

data "selectel_global_router_zone_v1" "cloud_zone" {
  name    = "%s"
  service = "vpc"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_cloud_network_v1" "network_tf_acc_test_1" {
  router_id     = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id       = data.selectel_global_router_zone_v1.cloud_zone.id
  os_network_id = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  project_id    = "%s"
  name          = "%s"
}
resource "selectel_global_router_cloud_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_cloud_network_v1.network_tf_acc_test_1.id
  os_subnet_id      = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.id
  cidr              = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.cidr
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
}
resource "selectel_global_router_static_route_v1" "static_route_tf_acc_test_1" {
  router_id = selectel_global_router_router_v1.router_tf_acc_test_1.id
  cidr      = "%s"
  next_hop  = "%s"
  name      = "%s"
  tags      = ["blue", "red"]
  # explicit dependency, because next_hop should be taken from subnet_tf_acc_test_1
  depends_on = [
    selectel_global_router_cloud_subnet_v1.subnet_tf_acc_test_1
  ]
}`, globalRouterVPCProjectID, globalRouterVPCRegion, globalRouterSubnetCidr,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName,
		globalRouterSubnetGateway, globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName,
		globalRouterStaticRouteCidr, globalRouterNextHop, staticRouteName)
}
