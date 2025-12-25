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

const resourceGlobalRouterVPCSubnetName = "selectel_global_router_vpc_subnet_v1.subnet_tf_acc_test_1"

func TestAccGlobalRouterVPCSubnetV1Basic(t *testing.T) {
	var subnet globalrouter.Subnet
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterVPCSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterVPCSubnetV1Basic(
					routerName, networkName, subnetName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCSubnetV1Exists(resourceGlobalRouterVPCSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "os_subnet_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "sv_subnet_id"),
				),
			},
		},
	})
}

func TestAccGlobalRouterVPCSubnetV1Update(t *testing.T) {
	var subnet globalrouter.Subnet
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterVPCSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterVPCSubnetV1Basic(
					routerName, networkName, subnetName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCSubnetV1Exists(resourceGlobalRouterVPCSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "os_subnet_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "sv_subnet_id"),
				),
			},
			{
				Config: testAccGlobalRouterVPCSubnetV1WithTags(
					routerName, networkName, subnetName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCSubnetV1Exists(resourceGlobalRouterVPCSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "os_subnet_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "sv_subnet_id"),
				),
			},
			{
				Config: testAccGlobalRouterVPCSubnetV1Basic(
					routerName, networkName, subnetName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCSubnetV1Exists(resourceGlobalRouterVPCSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "os_subnet_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCSubnetName, "sv_subnet_id"),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterVPCSubnetV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test subnet object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_vpc_subnet_v1" {
			continue
		}

		_, _, err := globalrouterClient.Subnet(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("subnet still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterVPCSubnetV1Exists(n string, subnet *globalrouter.Subnet) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get global_router client for test subnet object")
		}

		u, _, err := globalrouterClient.Subnet(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("subnet not found")
		}

		*subnet = *u

		return nil
	}
}

func testAccGlobalRouterVPCSubnetV1Basic(
	routerName string, networkName string, subnetName string,
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

data "selectel_global_router_zone_v1" "vpc_zone" {
  name    = "%s"
  service = "vpc"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_vpc_network_v1" "network_tf_acc_test_1" {
  router_id     = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id       = data.selectel_global_router_zone_v1.vpc_zone.id
  os_network_id = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  project_id    = "%s"
  name          = "%s"
}
resource "selectel_global_router_vpc_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_vpc_network_v1.network_tf_acc_test_1.id
  os_subnet_id      = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.id
  cidr              = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.cidr
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
}`, globalRouterVPCProjectID, globalRouterVPCRegion, globalRouterSubnetCidr,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName,
		globalRouterSubnetGateway, globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName)
}

func testAccGlobalRouterVPCSubnetV1WithTags(
	routerName string, networkName string, subnetName string,
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

data "selectel_global_router_zone_v1" "vpc_zone" {
  name    = "%s"
  service = "vpc"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_vpc_network_v1" "network_tf_acc_test_1" {
  router_id     = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id       = data.selectel_global_router_zone_v1.vpc_zone.id
  os_network_id = openstack_networking_network_v2.os_network_tf_acc_test_1.id
  project_id    = "%s"
  name          = "%s"
}
resource "selectel_global_router_vpc_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_vpc_network_v1.network_tf_acc_test_1.id
  os_subnet_id      = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.id
  cidr              = openstack_networking_subnet_v2.os_subnet_tf_acc_test_1.cidr
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
  tags              = ["blue", "red"]
}`, globalRouterVPCProjectID, globalRouterVPCRegion, globalRouterSubnetCidr,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName,
		globalRouterSubnetGateway, globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName)
}
