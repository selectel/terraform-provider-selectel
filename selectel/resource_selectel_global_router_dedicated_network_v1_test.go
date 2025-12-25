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

const resourceGlobalRouterDedicatedNetworkName = "selectel_global_router_dedicated_network_v1.network_tf_acc_test_1"

func TestAccGlobalRouterDedicatedNetworkV1Basic(t *testing.T) {
	var network globalrouter.Network
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedNetworkV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedNetworkV1Exists(resourceGlobalRouterDedicatedNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "zone_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "vlan", globalRouterDedicatedNetworkVLAN),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "inner_vlan", "0"),
				),
			},
		},
	})
}

func TestAccGlobalRouterDedicatedNetworkV1Update(t *testing.T) {
	var network globalrouter.Network
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedNetworkV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedNetworkV1Exists(resourceGlobalRouterDedicatedNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "zone_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "vlan", globalRouterDedicatedNetworkVLAN),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "inner_vlan", "0"),
				),
			},
			{
				Config: testAccGlobalRouterDedicatedNetworkV1WithTags(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedNetworkV1Exists(resourceGlobalRouterDedicatedNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "zone_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "vlan", globalRouterDedicatedNetworkVLAN),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "inner_vlan", "0"),
				),
			},
			{
				Config: testAccGlobalRouterDedicatedNetworkV1Basic(routerName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterDedicatedNetworkV1Exists(resourceGlobalRouterDedicatedNetworkName, &network),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "name", networkName),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "netops_vlan_uuid"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "router_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "sv_network_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterDedicatedNetworkName, "zone_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "vlan", globalRouterDedicatedNetworkVLAN),
					resource.TestCheckResourceAttr(resourceGlobalRouterDedicatedNetworkName, "inner_vlan", "0"),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterDedicatedNetworkV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test network object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_router_v1" {
			continue
		}

		_, _, err := globalrouterClient.Network(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("network still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterDedicatedNetworkV1Exists(n string, network *globalrouter.Network) resource.TestCheckFunc {
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

func testAccGlobalRouterDedicatedNetworkV1Basic(routerName string, networkName string) string {
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
}`, globalRouterDedicatedRegion, routerName, globalRouterDedicatedNetworkVLAN, networkName)
}

func testAccGlobalRouterDedicatedNetworkV1WithTags(routerName string, networkName string) string {
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
  tags      = ["blue", "red"]
}`, globalRouterDedicatedRegion, routerName, globalRouterDedicatedNetworkVLAN, networkName)
}
