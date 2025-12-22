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

const resourceGlobalRouterVPCNetworkName = "selectel_global_router_vpc_network_v1.network_tf_acc_test_1"

func TestAccGlobalRouterVPCNetworkV1Basic(t *testing.T) {
	var network globalrouter.Network
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterVPCNetworkV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterVPCNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCNetworkV1Exists(resourceGlobalRouterVPCNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "vlan"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "project_id", globalRouterVPCProjectID),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "os_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "zone_id"),
				),
			},
		},
	})
}

func TestAccGlobalRouterVPCNetworkV1Update(t *testing.T) {
	var network globalrouter.Network
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterVPCNetworkV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterVPCNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCNetworkV1Exists(resourceGlobalRouterVPCNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "vlan"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "project_id", globalRouterVPCProjectID),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "os_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "zone_id"),
				),
			},
			{
				Config: testAccGlobalRouterVPCNetworkV1WithTags(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCNetworkV1Exists(resourceGlobalRouterVPCNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "vlan"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "project_id", globalRouterVPCProjectID),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "os_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "zone_id"),
				),
			},
			{
				Config: testAccGlobalRouterVPCNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterVPCNetworkV1Exists(resourceGlobalRouterVPCNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "vlan"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterVPCNetworkName, "project_id", globalRouterVPCProjectID),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "os_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterVPCNetworkName, "zone_id"),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterVPCNetworkV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test network object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_vpc_network_v1" {
			continue
		}

		_, _, err := globalrouterClient.Network(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("network still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterVPCNetworkV1Exists(n string, network *globalrouter.Network) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get global_router client for test network object")
		}

		u, _, err := globalrouterClient.Network(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("network not found")
		}

		*network = *u

		return nil
	}
}

func testAccGlobalRouterVPCNetworkV1Basic(routerName string, networkName string) string {
	return fmt.Sprintf(`
provider openstack {
	tenant_id = "%s"
}
resource "openstack_networking_network_v2" "os_network_tf_acc_test_1" {
 	region = "%s"
  	name   = "network_one"
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
}`, globalRouterVPCProjectID, globalRouterVPCRegion,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName)
}

func testAccGlobalRouterVPCNetworkV1WithTags(routerName string, networkName string) string {
	return fmt.Sprintf(`
provider openstack {
	tenant_id = "%s"
}
resource "openstack_networking_network_v2" "os_network_tf_acc_test_1" {
 	region = "%s"
  	name   = "network_one"
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
  tags          = ["blue", "red"]
}`, globalRouterVPCProjectID, globalRouterVPCRegion,
		globalRouterVPCRegion, routerName, globalRouterVPCProjectID, networkName)
}
