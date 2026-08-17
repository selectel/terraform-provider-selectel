package selectel

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCRouterRouteV2ResourceName = "selectel_vpc_router_route_v2.route_tf_acc_test_1"

func TestVPCRouterRouteV2ID(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		route := router.Route{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"}
		id := vpcRouterRouteV2ID("router-id", route)

		assert.Equal(t, "router-id-route-10.0.0.0/8-192.168.199.10", id)

		routerID, parsed, err := parseVPCRouterRouteV2ID(id)

		require.NoError(t, err)
		assert.Equal(t, "router-id", routerID)
		assert.Equal(t, route, parsed)
	})

	for _, id := range []string{
		"",
		"router-id",
		"router-id-route-10.0.0.0/8",
		"router-id-route-10.0.0.0/8-192.168.199.10-extra",
		"-route-10.0.0.0/8-192.168.199.10",
		"router-id-route--192.168.199.10",
	} {
		t.Run("rejects "+id, func(t *testing.T) {
			_, _, err := parseVPCRouterRouteV2ID(id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}

	t.Run("a second separator is rejected rather than guessed at", func(t *testing.T) {
		_, _, err := parseVPCRouterRouteV2ID("a-route-b-route-10.0.0.0/8-192.168.199.10")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected the format")
	})
}

func TestAccVPCRouterRouteV2Basic(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					resource.TestCheckResourceAttr(testAccVPCRouterRouteV2ResourceName, "destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccVPCRouterRouteV2ResourceName, "next_hop", "192.168.199.10"),
				),
			},
			{
				Config:   testAccVPCRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCRouterRouteV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCRouterRouteV2Basic(projectID, region, "172.16.0.0/12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterRouteV2Exists("172.16.0.0/12", "192.168.199.10"),
					testAccCheckVPCRouterRouteV2Absent("10.0.0.0/8", "192.168.199.10"),
				),
			},
		},
	})
}

func testAccVPCRouterRouteV2Router(s *terraform.State) (*vpcNetworkV2TestHandle, error) {
	return testAccVPCNetworkV2Handle(s, "selectel_vpc_router_v2.router_tf_acc_test_1")
}

func testAccCheckVPCRouterRouteV2Exists(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccVPCRouterRouteV2Routes(s)
		if err != nil {
			return err
		}

		route := router.Route{Destination: destination, NextHop: nextHop}
		if !slices.Contains(routes, route) {
			return fmt.Errorf("route %s via %s is not present on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckVPCRouterRouteV2Absent(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccVPCRouterRouteV2Routes(s)
		if err != nil {
			return err
		}

		route := router.Route{Destination: destination, NextHop: nextHop}
		if slices.Contains(routes, route) {
			return fmt.Errorf("route %s via %s is still present on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccVPCRouterRouteV2Routes(s *terraform.State) ([]router.Route, error) {
	handle, err := testAccVPCRouterRouteV2Router(s)
	if err != nil {
		return nil, err
	}

	found, err := router.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.Routes, nil
}

func testAccCheckVPCRouterRouteV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_route_v2" {
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

		found, err := router.Get(context.Background(), client, rs.Primary.Attributes["router_id"])
		if err != nil {
			if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
				continue
			}

			return err
		}

		route := router.Route{
			Destination: rs.Primary.Attributes["destination_cidr"],
			NextHop:     rs.Primary.Attributes["next_hop"],
		}
		if slices.Contains(found.Routes, route) {
			return fmt.Errorf("VPC router route %s via %s still exists",
				route.Destination, route.NextHop)
		}
	}

	return nil
}

func testAccVPCRouterRouteV2Prerequisites(projectID, region string) string {
	return testAccVPCRouterInterfaceV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}
`, region)
}

func testAccVPCRouterRouteV2Basic(projectID, region, destination string) string {
	return testAccVPCRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_route_v2" "route_tf_acc_test_1" {
  project_id       = local.project_id
  region           = "%s"
  router_id        = selectel_vpc_router_v2.router_tf_acc_test_1.id
  destination_cidr = "%s"
  next_hop         = "192.168.199.10"

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region, destination)
}
