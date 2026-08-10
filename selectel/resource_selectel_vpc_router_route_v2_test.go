package selectel

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingRouterRouteV2ResourceName = "selectel_vpc_router_route_v2.route_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema and identifier unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingRouterRouteV2Schema covers REQ-073 and REQ-074: every attribute
// is required and replaces the route, because there is no update path.
func TestNetworkingRouterRouteV2Schema(t *testing.T) {
	s := resourceVPCRouterRouteV2().Schema

	assert.Len(t, s, 5, "the route is fully described by its five arguments")

	for _, name := range []string{"project_id", "region", "router_id", "destination_cidr", "next_hop"} {
		t.Run(name+" is required and ForceNew", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
			assert.False(t, s[name].Computed)
		})
	}

	// REQ-074: no update path means the resource declares no UpdateContext.
	t.Run("the resource has no update path", func(t *testing.T) {
		assert.Nil(t, resourceVPCRouterRouteV2().UpdateContext)
	})
}

// TestNetworkingRouterRouteV2ID covers REQ-077: the composite identifier round
// trips, and a malformed one is rejected locally.
func TestNetworkingRouterRouteV2ID(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		route := router.Route{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"}
		id := networkingRouterRouteV2ID("router-id", route)

		assert.Equal(t, "router-id-route-10.0.0.0/8-192.168.199.10", id)

		routerID, parsed, err := parseNetworkingRouterRouteV2ID(id)

		require.NoError(t, err)
		assert.Equal(t, "router-id", routerID)
		assert.Equal(t, route, parsed)
	})

	// SCN-077-02: rejected before any API call.
	for _, id := range []string{
		"",
		"router-id",
		"router-id-route-10.0.0.0/8",
		"router-id-route-10.0.0.0/8-192.168.199.10-extra",
		"-route-10.0.0.0/8-192.168.199.10",
		"router-id-route--192.168.199.10",
	} {
		t.Run("rejects "+id, func(t *testing.T) {
			_, _, err := parseNetworkingRouterRouteV2ID(id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}

	// REQ-077 splits at the FIRST separator and then requires the remainder to be
	// exactly two dash-separated parts. An identifier carrying a second "-route-"
	// therefore yields four parts and is rejected rather than guessed at. Router
	// identifiers are UUIDs, so this cannot arise from a real router.
	t.Run("a second separator is rejected rather than guessed at", func(t *testing.T) {
		_, _, err := parseNetworkingRouterRouteV2ID("a-route-b-route-10.0.0.0/8-192.168.199.10")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected the format")
	})
}

// TestNetworkingRouterRouteV2Index covers the matching rule behind REQ-073's
// idempotent create and REQ-076's detection of an externally removed route: the
// pair matches, not the destination alone.
func TestNetworkingRouterRouteV2Index(t *testing.T) {
	routes := []router.Route{
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
		{Destination: "172.16.0.0/12", NextHop: "192.168.199.11"},
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.12"},
	}

	t.Run("finds an exact pair", func(t *testing.T) {
		assert.Equal(t, 0, networkingRouterRouteV2Index(routes, routes[0]))
		assert.Equal(t, 2, networkingRouterRouteV2Index(routes, routes[2]))
	})

	t.Run("the same destination with another next hop is a different route", func(t *testing.T) {
		assert.Equal(t, -1, networkingRouterRouteV2Index(routes,
			router.Route{Destination: "10.0.0.0/8", NextHop: "192.168.199.99"}))
	})

	t.Run("reports absence", func(t *testing.T) {
		assert.Equal(t, -1, networkingRouterRouteV2Index(nil, routes[0]))
		assert.Equal(t, -1, networkingRouterRouteV2Index(routes,
			router.Route{Destination: "192.0.2.0/24", NextHop: "192.168.199.10"}))
	})
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingRouterRouteV2Basic covers SCN-073-01, SCN-073-03, SCN-077-01
// and SCN-078-03.
func TestAccNetworkingRouterRouteV2Basic(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					resource.TestCheckResourceAttr(testAccNetworkingRouterRouteV2ResourceName, "destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccNetworkingRouterRouteV2ResourceName, "next_hop", "192.168.199.10"),
				),
			},
			{
				// SCN-073-03: the follow-up plan is empty.
				Config:   testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				PlanOnly: true,
			},
			{
				// SCN-077-01: import by the composite identifier.
				ResourceName:      testAccNetworkingRouterRouteV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// SCN-074-01: a different destination replaces the route rather
				// than updating it, and the old entry is removed from the router.
				Config: testAccNetworkingRouterRouteV2Basic(projectID, region, "172.16.0.0/12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRouteV2Exists("172.16.0.0/12", "192.168.199.10"),
					testAccCheckNetworkingRouterRouteV2Absent("10.0.0.0/8", "192.168.199.10"),
				),
			},
		},
	})
}

// TestAccNetworkingRouterRouteV2Adopt covers SCN-073-02: an identical entry that
// already exists is adopted without a conflict and without rewriting routes.
func TestAccNetworkingRouterRouteV2Adopt(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				// Only the router and the interface, so the route can be added
				// out of band before the route resource is applied.
				Config: testAccNetworkingRouterRouteV2Prerequisites(projectID, region),
				Check:  testAccCheckNetworkingRouterRouteV2AddOutOfBand("10.0.0.0/8", "192.168.199.10"),
			},
			{
				Config: testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					// The adopted entry is the only one: routes was not rewritten
					// with a duplicate.
					testAccCheckNetworkingRouterRouteV2Count(1),
				),
			},
			{
				Config:   testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingRouterRouteV2NextHopValidation covers SCN-075-01 and
// SCN-075-02: the base check, and the Selectel fork excluding the router's
// external-network ports from the acceptable next hops.
func TestAccNetworkingRouterRouteV2NextHopValidation(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-075-01: the next hop is in no subnet the router serves.
				Config:      testAccNetworkingRouterRouteV2NextHop(projectID, region, false, "203.0.113.200"),
				ExpectError: regexp.MustCompile(`(?i)invalid|not connected|next.?hop`),
			},
			{
				// SCN-075-02: an address inside the router's external gateway
				// subnet. Upstream Neutron would accept this; the Selectel fork
				// rejects it with the same class of error as SCN-075-01, because
				// external-network ports no longer count.
				Config:      testAccNetworkingRouterRouteV2NextHop(projectID, region, true, ""),
				ExpectError: regexp.MustCompile(`(?i)invalid|not connected|next.?hop`),
			},
		},
	})
}

// TestAccNetworkingRouterRouteV2ExternalRemoval covers SCN-076-01: the route
// removed from the router outside Terraform leaves the state without an error.
func TestAccNetworkingRouterRouteV2ExternalRemoval(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check:  testAccCheckNetworkingRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			},
			{
				// The router stays, only its routes are cleared.
				Config:             testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check:              testAccCheckNetworkingRouterRouteV2ClearOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				// The plan re-creates the entry rather than erroring out.
				Config: testAccNetworkingRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
				Check:  testAccCheckNetworkingRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

// testAccNetworkingRouterRouteV2Router returns a client handle and the router the
// route under test belongs to, resolved from the router resource in the state so
// it also works in the step where the route resource does not exist yet.
func testAccNetworkingRouterRouteV2Router(s *terraform.State) (*networkingNetworkV2TestHandle, error) {
	return testAccNetworkingNetworkV2Handle(s, "selectel_vpc_router_v2.router_tf_acc_test_1")
}

func testAccCheckNetworkingRouterRouteV2Exists(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRouteV2Routes(s)
		if err != nil {
			return err
		}

		route := router.Route{Destination: destination, NextHop: nextHop}
		if networkingRouterRouteV2Index(routes, route) < 0 {
			return fmt.Errorf("route %s via %s is not present on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRouteV2Absent(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRouteV2Routes(s)
		if err != nil {
			return err
		}

		route := router.Route{Destination: destination, NextHop: nextHop}
		if networkingRouterRouteV2Index(routes, route) >= 0 {
			return fmt.Errorf("route %s via %s is still present on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRouteV2Count(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRouteV2Routes(s)
		if err != nil {
			return err
		}

		if len(routes) != expected {
			return fmt.Errorf("the router has %d routes, expected %d", len(routes), expected)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRouteV2AddOutOfBand(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingRouterRouteV2Router(s)
		if err != nil {
			return err
		}

		routes := []router.Route{{Destination: destination, NextHop: nextHop}}
		_, err = router.Update(context.Background(), handle.client, handle.id,
			router.UpdateRequest{Routes: &routes})

		return err
	}
}

func testAccCheckNetworkingRouterRouteV2ClearOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingRouterRouteV2Router(s)
		if err != nil {
			return err
		}

		routes := []router.Route{}
		_, err = router.Update(context.Background(), handle.client, handle.id,
			router.UpdateRequest{Routes: &routes})

		return err
	}
}

func testAccNetworkingRouterRouteV2Routes(s *terraform.State) ([]router.Route, error) {
	handle, err := testAccNetworkingRouterRouteV2Router(s)
	if err != nil {
		return nil, err
	}

	found, err := router.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.Routes, nil
}

// testAccCheckNetworkingRouterRouteV2Destroy asserts that destroying the resource
// left no entry behind on a router that still exists. A router that is gone
// satisfies it too: the routes went with it.
func testAccCheckNetworkingRouterRouteV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_route_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		found, err := router.Get(context.Background(), client, rs.Primary.Attributes["router_id"])
		if err != nil {
			if networkingIsNotFound(err) {
				continue
			}

			return err
		}

		route := router.Route{
			Destination: rs.Primary.Attributes["destination_cidr"],
			NextHop:     rs.Primary.Attributes["next_hop"],
		}
		if networkingRouterRouteV2Index(found.Routes, route) >= 0 {
			return fmt.Errorf("networking router route %s via %s still exists",
				route.Destination, route.NextHop)
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingRouterRouteV2Prerequisites is a router with one internal
// interface, which is what makes 192.168.199.10 an acceptable next hop.
func testAccNetworkingRouterRouteV2Prerequisites(projectID, region string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}
`, region)
}

func testAccNetworkingRouterRouteV2Basic(projectID, region, destination string) string {
	return testAccNetworkingRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
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

// testAccNetworkingRouterRouteV2NextHop optionally attaches an external gateway,
// which is what distinguishes SCN-075-02 from SCN-075-01.
// testAccNetworkingRouterRouteV2NextHop points a route at a next hop that no
// internal port of the router serves.
//
// Without an external gateway the next hop is the literal address given. With
// one, the router takes its gateway from the floating IP's external network and
// the next hop becomes the allocated address — which is what
// SCN-075-02 needs: a real address of the external subnet. A project cannot read
// the subnets of an external network at all (they answer 404), so an allocated
// floating IP is the only way to name one of their addresses without admin rights.
func testAccNetworkingRouterRouteV2NextHop(projectID, region string, withExternal bool, nextHop string) string {
	gateway, allocation := "", ""
	if withExternal {
		gateway = "external_network_id = data.selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.floating_network_id"
		nextHop = "${selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.floating_ip_address}"
		allocation = fmt.Sprintf(`
resource "selectel_vpc_floatingip_v2" "next_hop_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}

data "selectel_vpc_floatingip_v2" "next_hop_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%s"
  floatingip_id = selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.id
}
`, region, region)
	}

	return testAccNetworkingV2Base(projectID) + allocation + fmt.Sprintf(`
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
  project_id = local.project_id
  region     = "%[1]s"
  name       = "router-tf-acc-test-1"
  %[2]s
}

resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}

resource "selectel_vpc_router_route_v2" "route_tf_acc_test_1" {
  project_id       = local.project_id
  region           = "%[1]s"
  router_id        = selectel_vpc_router_v2.router_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "%[3]s"

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region, gateway, nextHop)
}
