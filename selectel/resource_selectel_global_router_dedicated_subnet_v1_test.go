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

const resourceGlobalRouterDedicatedSubnetName = "selectel_global_router_dedicated_subnet_v1.subnet_tf_acc_test_1"

func TestAccGlobalRouterDedicatedSubnetV1Basic(t *testing.T) {
	var subnet globalrouter.Subnet
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedSubnetV1Basic(routerName, networkName, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedSubnetV1Exists(resourceGlobalRouterDedicatedSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "network_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "sv_subnet_id", ""),
				),
			},
		},
	})
}

func TestAccGlobalRouterDedicatedSubnetV1Update(t *testing.T) {
	var subnet globalrouter.Subnet
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedSubnetV1Basic(routerName, networkName, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedSubnetV1Exists(resourceGlobalRouterDedicatedSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "network_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "sv_subnet_id", ""),
				),
			},
			{
				Config: testAccGlobalRouterDedicatedSubnetV1WithTags(routerName, networkName, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedSubnetV1Exists(resourceGlobalRouterDedicatedSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "network_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "sv_subnet_id", ""),
				),
			},
			{
				Config: testAccGlobalRouterDedicatedSubnetV1Basic(routerName, networkName, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedSubnetV1Exists(resourceGlobalRouterDedicatedSubnetName, &subnet),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "name", subnetName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "cidr", globalRouterSubnetCidr),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "gateway", globalRouterSubnetGateway),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.0", globalRouterSubnetServiceAddress1),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "service_addresses.1", globalRouterSubnetServiceAddress2),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "netops_subnet_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedSubnetName, "network_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedSubnetName, "sv_subnet_id", ""),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterDedicatedSubnetV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test subnet object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_dedicated_subnet_v1" {
			continue
		}

		_, _, err := globalrouterClient.Subnet(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("subnet still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterDedicatedSubnetV1Exists(n string, subnet *globalrouter.Subnet) resource.TestCheckFunc {
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

func testAccGlobalRouterDedicatedSubnetV1Basic(
	routerName string, networkName string, subnetName string,
) string {
	return fmt.Sprintf(`
data "selectel_global_router_zone_v1" "dedicated_zone" {
  name    = "%s"
  service = "dedicated"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_dedicated_network_v1" "network_tf_acc_test_1" {
  router_id = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id   = data.selectel_global_router_zone_v1.dedicated_zone.id
  vlan      = %s
  name      = "%s"
}
resource "selectel_global_router_dedicated_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_dedicated_network_v1.network_tf_acc_test_1.id
  cidr              = "%s"
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
}`, globalRouterDedicatedRegion, routerName, globalRouterDedicatedNetworkVLAN, networkName,
		globalRouterSubnetCidr, globalRouterSubnetGateway,
		globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName)
}

func testAccGlobalRouterDedicatedSubnetV1WithTags(
	routerName string, networkName string, subnetName string,
) string {
	return fmt.Sprintf(`
data "selectel_global_router_zone_v1" "dedicated_zone" {
  name    = "%s"
  service = "dedicated"
}
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}
resource "selectel_global_router_dedicated_network_v1" "network_tf_acc_test_1" {
  router_id = selectel_global_router_router_v1.router_tf_acc_test_1.id
  zone_id   = data.selectel_global_router_zone_v1.dedicated_zone.id
  vlan      = %s
  name      = "%s"
}
resource "selectel_global_router_dedicated_subnet_v1" "subnet_tf_acc_test_1" {
  network_id        = selectel_global_router_dedicated_network_v1.network_tf_acc_test_1.id
  cidr              = "%s"
  gateway           = "%s"
  service_addresses = ["%s", "%s"]
  name              = "%s"
  tags              = ["blue", "red"]
}`, globalRouterDedicatedRegion, routerName, globalRouterDedicatedNetworkVLAN, networkName,
		globalRouterSubnetCidr, globalRouterSubnetGateway,
		globalRouterSubnetServiceAddress1, globalRouterSubnetServiceAddress2, subnetName)
}
