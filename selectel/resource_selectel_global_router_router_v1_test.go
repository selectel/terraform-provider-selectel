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

const resourceGlobalRouterRouterName = "selectel_global_router_router_v1.router_tf_acc_test_1"

func TestAccGlobalRouterRouterV1Basic(t *testing.T) {
	var router globalrouter.Router
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterRouterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterRouterV1Basic(routerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterRouterV1Exists(resourceGlobalRouterRouterName, &router),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "name", routerName),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "netops_router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "leak_uuid", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "prefix_pool_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "vpn_id"),
				),
			},
		},
	})
}

func TestAccGlobalRouterRouterV1Update(t *testing.T) {
	var router globalrouter.Router
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterRouterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterRouterV1Basic(routerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterRouterV1Exists(resourceGlobalRouterRouterName, &router),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "name", routerName),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "netops_router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "leak_uuid", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "prefix_pool_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "vpn_id"),
				),
			},
			{
				Config: testAccGlobalRouterRouterV1WithTags(routerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterRouterV1Exists(resourceGlobalRouterRouterName, &router),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "name", routerName),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "tags.0", "blue"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "tags.1", "red"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "netops_router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "leak_uuid", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "prefix_pool_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "vpn_id"),
				),
			},
			{
				Config: testAccGlobalRouterRouterV1Basic(routerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRouterRouterV1Exists(resourceGlobalRouterRouterName, &router),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "name", routerName),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "tags.#", "0"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "updated_at"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "account_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "project_id", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "netops_router_id"),
					resource.TestCheckResourceAttr(resourceGlobalRouterRouterName, "leak_uuid", ""),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "prefix_pool_id"),
					resource.TestCheckResourceAttrSet(resourceGlobalRouterRouterName, "vpn_id"),
				),
			},
		},
	})
}

func testAccCheckGlobalRouterRouterV1Destroy(s *terraform.State) error {
	globalrouterClient, diagErr := getGlobalRouterClient(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get global_routerclient for test router object")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_global_router_router_v1" {
			continue
		}

		_, _, err := globalrouterClient.Router(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("router still exists")
		}
	}

	return nil
}

func testAccCheckGlobalRouterRouterV1Exists(n string, router *globalrouter.Router) resource.TestCheckFunc {
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
			return fmt.Errorf("can't get global_routerclient for test router object")
		}

		u, _, err := globalrouterClient.Router(context.Background(), rs.Primary.ID)
		if err != nil {
			return errors.New("router not found")
		}

		*router = *u

		return nil
	}
}

func testAccGlobalRouterRouterV1Basic(routerName string) string {
	return fmt.Sprintf(`
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
}`, routerName)
}

func testAccGlobalRouterRouterV1WithTags(routerName string) string {
	return fmt.Sprintf(`
resource "selectel_global_router_router_v1" "router_tf_acc_test_1" {
  name = "%s"
  tags = ["blue", "red"]
}`, routerName)
}
