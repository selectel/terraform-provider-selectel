package selectel

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCRouterRoutesV2ResourceName = "selectel_vpc_router_routes_v2.routes_tf_acc_test_1"

func TestVPCRouterRoutesV2Expand(t *testing.T) {
	s := resourceVPCRouterRoutesV2().Schema

	t.Run("an unset routes attribute expands to an empty list", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]any{})

		assert.Empty(t, expandVPCRouterRoutesV2(d.Get("routes")))
	})

	t.Run("a declared set expands to the same pairs", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]any{
			"routes": []any{
				map[string]any{"destination_cidr": "10.0.0.0/8", "next_hop": "192.168.199.10"},
				map[string]any{"destination_cidr": "172.16.0.0/12", "next_hop": "192.168.199.11"},
			},
		})

		assert.ElementsMatch(t, []router.Route{
			{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
			{Destination: "172.16.0.0/12", NextHop: "192.168.199.11"},
		}, expandVPCRouterRoutesV2(d.Get("routes")))
	})

	t.Run("a non-set value expands to an empty list", func(t *testing.T) {
		assert.Empty(t, expandVPCRouterRoutesV2(nil))
	})
}

func TestVPCRouterRoutesV2Flatten(t *testing.T) {
	routes := []router.Route{
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.11"},
	}

	d := schema.TestResourceDataRaw(t, resourceVPCRouterRoutesV2().Schema, map[string]any{})
	require.NoError(t, d.Set("routes", flattenVPCRouterRoutesV2(routes)))

	assert.ElementsMatch(t, routes, expandVPCRouterRoutesV2(d.Get("routes")))
}

func TestAccVPCRouterRoutesV2Basic(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRoutesV2IDIsRouter(),
					resource.TestCheckResourceAttr(testAccVPCRouterRoutesV2ResourceName, "routes.#", "2"),
					testAccCheckVPCRouterRoutesV2Count(2),
				),
			},
			{
				Config:   testAccVPCRouterRoutesV2Pair(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCRouterRoutesV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouterRoutesV2Update(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRoutesV2Has("10.0.0.0/8", "192.168.199.10"),
					testAccCheckVPCRouterRoutesV2Has("172.16.0.0/12", "192.168.199.10"),
				),
			},
			{
				Config: testAccVPCRouterRoutesV2Replaced(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRoutesV2IDIsRouter(),
					testAccCheckVPCRouterRoutesV2Has("10.0.0.0/8", "192.168.199.10"),
					testAccCheckVPCRouterRoutesV2Has("192.0.2.0/24", "192.168.199.10"),
					testAccCheckVPCRouterRoutesV2HasNot("172.16.0.0/12", "192.168.199.10"),
					testAccCheckVPCRouterRoutesV2Count(2),
				),
			},
			{
				Config: testAccVPCRouterRoutesV2Empty(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRoutesV2IDIsRouter(),
					testAccCheckVPCRouterRoutesV2Count(0),
				),
			},
		},
	})
}

func TestAccVPCRouterRoutesV2Authoritative(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterRoutesV2Pair(projectID, region),
				Check:  testAccCheckVPCRouterRoutesV2Count(2),
			},
			{
				Config:             testAccVPCRouterRoutesV2Pair(projectID, region),
				Check:              testAccCheckVPCRouterRoutesV2AddOutOfBand("192.0.2.0/24", "192.168.199.10"),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRoutesV2HasNot("192.0.2.0/24", "192.168.199.10"),
					testAccCheckVPCRouterRoutesV2Count(2),
				),
			},
		},
	})
}

func testAccCheckVPCRouterRoutesV2IDIsRouter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRouterRoutesV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRouterRoutesV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["router_id"] {
			return fmt.Errorf("resource id %s is not the router id %s",
				rs.Primary.ID, rs.Primary.Attributes["router_id"])
		}

		return nil
	}
}

func testAccCheckVPCRouterRoutesV2Has(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccVPCRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if !slices.Contains(routes, router.Route{Destination: destination, NextHop: nextHop}) {
			return fmt.Errorf("route %s via %s is missing from the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckVPCRouterRoutesV2HasNot(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccVPCRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if slices.Contains(routes, router.Route{Destination: destination, NextHop: nextHop}) {
			return fmt.Errorf("route %s via %s is still on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckVPCRouterRoutesV2Count(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccVPCRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if len(routes) != expected {
			return fmt.Errorf("the router has %d routes, expected %d", len(routes), expected)
		}

		return nil
	}
}

func testAccCheckVPCRouterRoutesV2AddOutOfBand(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterRoutesV2ResourceName)
		if err != nil {
			return err
		}

		found, err := router.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		routes := append(append([]router.Route{}, found.Routes...),
			router.Route{Destination: destination, NextHop: nextHop})
		_, err = router.Update(context.Background(), handle.client, handle.id,
			router.UpdateRequest{Routes: &routes})

		return err
	}
}

func testAccVPCRouterRoutesV2Routes(s *terraform.State) ([]router.Route, error) {
	handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterRoutesV2ResourceName)
	if err != nil {
		return nil, err
	}

	found, err := router.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.Routes, nil
}

func testAccCheckVPCRouterRoutesV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_routes_v2" {
			continue
		}

		client, err := vpcV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		found, err := router.Get(context.Background(), client, rs.Primary.ID)
		if err != nil {
			if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
				continue
			}

			return err
		}

		if len(found.Routes) != 0 {
			return fmt.Errorf("router %s still has %d routes after destroy",
				rs.Primary.ID, len(found.Routes))
		}
	}

	return nil
}

func testAccVPCRouterRoutesV2Pair(projectID, region string) string {
	return testAccVPCRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_routes_v2" "routes_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id

  routes {
    destination_cidr = "10.0.0.0/8"
    next_hop         = "192.168.199.10"
  }

  routes {
    destination_cidr = "172.16.0.0/12"
    next_hop         = "192.168.199.10"
  }

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region)
}

func testAccVPCRouterRoutesV2Replaced(projectID, region string) string {
	return testAccVPCRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_routes_v2" "routes_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id

  routes {
    destination_cidr = "10.0.0.0/8"
    next_hop         = "192.168.199.10"
  }

  routes {
    destination_cidr = "192.0.2.0/24"
    next_hop         = "192.168.199.10"
  }

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
	`, region)
}

func testAccVPCRouterRoutesV2Empty(projectID, region string) string {
	return testAccVPCRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_routes_v2" "routes_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region)
}
