package selectel

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingRouterRoutesV2ResourceName = "selectel_vpc_router_routes_v2.routes_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingRouterRoutesV2Schema covers REQ-079 and REQ-082: the resource is
// bound to one router, and routes is an updatable authoritative declaration.
func TestNetworkingRouterRoutesV2Schema(t *testing.T) {
	s := resourceVPCRouterRoutesV2().Schema

	t.Run("project_id, region and router_id are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region", "router_id"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-082: routes is updated in place — that is the whole point of the
	// resource — and Read fills it, so it is Computed as well.
	t.Run("routes is optional, computed and updatable", func(t *testing.T) {
		require.Contains(t, s, "routes")
		assert.True(t, s["routes"].Optional)
		assert.True(t, s["routes"].Computed)
		assert.False(t, s["routes"].ForceNew)
		assert.Equal(t, schema.TypeSet, s["routes"].Type)
	})

	t.Run("each route requires both halves of the pair", func(t *testing.T) {
		element, ok := s["routes"].Elem.(*schema.Resource)
		require.True(t, ok)

		for _, name := range []string{"destination_cidr", "next_hop"} {
			require.Contains(t, element.Schema, name)
			assert.True(t, element.Schema[name].Required)
		}
	})
}

// TestNetworkingRouterRoutesV2Expand covers the conversion behind REQ-081 and
// REQ-082, including the empty declaration of SCN-081-02.
func TestNetworkingRouterRoutesV2Expand(t *testing.T) {
	s := resourceVPCRouterRoutesV2().Schema

	t.Run("an unset routes attribute expands to an empty list", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{})

		assert.Empty(t, expandNetworkingRouterRoutesV2(d.Get("routes")))
	})

	t.Run("a declared set expands to the same pairs", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, s, map[string]interface{}{
			"routes": []interface{}{
				map[string]interface{}{"destination_cidr": "10.0.0.0/8", "next_hop": "192.168.199.10"},
				map[string]interface{}{"destination_cidr": "172.16.0.0/12", "next_hop": "192.168.199.11"},
			},
		})

		assert.ElementsMatch(t, []router.Route{
			{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
			{Destination: "172.16.0.0/12", NextHop: "192.168.199.11"},
		}, expandNetworkingRouterRoutesV2(d.Get("routes")))
	})

	t.Run("a non-set value expands to an empty list", func(t *testing.T) {
		assert.Empty(t, expandNetworkingRouterRoutesV2(nil))
	})
}

// TestNetworkingRouterRoutesV2Flatten covers the Read side of REQ-082: the
// router's actual list round trips through the schema shape.
func TestNetworkingRouterRoutesV2Flatten(t *testing.T) {
	routes := []router.Route{
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.11"},
	}

	d := schema.TestResourceDataRaw(t, resourceVPCRouterRoutesV2().Schema, map[string]interface{}{})
	require.NoError(t, d.Set("routes", flattenNetworkingRouterRoutesV2(routes)))

	// The same destination with two next hops must survive as two entries.
	assert.ElementsMatch(t, routes, expandNetworkingRouterRoutesV2(d.Get("routes")))
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingRouterRoutesV2Basic covers SCN-079-01, SCN-081-01, SCN-080-01
// and SCN-084-01.
func TestAccNetworkingRouterRoutesV2Basic(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					// SCN-079-01: the identity is the router itself.
					testAccCheckNetworkingRouterRoutesV2IDIsRouter(),
					resource.TestCheckResourceAttr(testAccNetworkingRouterRoutesV2ResourceName, "routes.#", "2"),
					testAccCheckNetworkingRouterRoutesV2Count(2),
				),
			},
			{
				Config:   testAccNetworkingRouterRoutesV2Pair(projectID, region),
				PlanOnly: true,
			},
			{
				// SCN-080-01: import by router_id alone restores the full list.
				ResourceName:      testAccNetworkingRouterRoutesV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingRouterRoutesV2Empty covers SCN-081-02 and SCN-084-02: an empty
// declaration adds nothing and destroys as a no-op.
func TestAccNetworkingRouterRoutesV2Empty(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterRoutesV2Empty(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRoutesV2IDIsRouter(),
					resource.TestCheckResourceAttr(testAccNetworkingRouterRoutesV2ResourceName, "routes.#", "0"),
					testAccCheckNetworkingRouterRoutesV2Count(0),
				),
			},
			{
				Config:   testAccNetworkingRouterRoutesV2Empty(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingRouterRoutesV2Update covers SCN-082-01: the diff between the
// two declarations is applied, adding and removing exactly what changed.
func TestAccNetworkingRouterRoutesV2Update(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				// A and B.
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRoutesV2Has("10.0.0.0/8", "192.168.199.10"),
					testAccCheckNetworkingRouterRoutesV2Has("172.16.0.0/12", "192.168.199.10"),
				),
			},
			{
				// A and C: B is removed, C is added, A stays.
				Config: testAccNetworkingRouterRoutesV2Replaced(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRoutesV2IDIsRouter(),
					testAccCheckNetworkingRouterRoutesV2Has("10.0.0.0/8", "192.168.199.10"),
					testAccCheckNetworkingRouterRoutesV2Has("192.0.2.0/24", "192.168.199.10"),
					testAccCheckNetworkingRouterRoutesV2HasNot("172.16.0.0/12", "192.168.199.10"),
					testAccCheckNetworkingRouterRoutesV2Count(2),
				),
			},
		},
	})
}

// TestAccNetworkingRouterRoutesV2Authoritative covers SCN-082-02: a route added
// by another mechanism becomes drift and is removed by the next apply. This is
// the behaviour that distinguishes this resource from the single-route sibling.
func TestAccNetworkingRouterRoutesV2Authoritative(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check:  testAccCheckNetworkingRouterRoutesV2Count(2),
			},
			{
				// The undeclared route X appears as drift after refresh.
				Config:             testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check:              testAccCheckNetworkingRouterRoutesV2AddOutOfBand("192.0.2.0/24", "192.168.199.10"),
				ExpectNonEmptyPlan: true,
			},
			{
				// The apply removes X, leaving only the declared pair.
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterRoutesV2HasNot("192.0.2.0/24", "192.168.199.10"),
					testAccCheckNetworkingRouterRoutesV2Count(2),
				),
			},
		},
	})
}

// TestAccNetworkingRouterRoutesV2NextHopValidation covers SCN-083-01 and
// SCN-083-02: an accepted next hop, and the atomic rejection of the whole set
// when one next hop is only reachable through an external-network port.
func TestAccNetworkingRouterRoutesV2NextHopValidation(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterRoutesV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-083-01: a next hop inside an internal interface's CIDR.
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check:  testAccCheckNetworkingRouterRoutesV2Count(2),
			},
			{
				// SCN-083-02: one bad next hop rejects the whole set, so the
				// router keeps the two routes it already had.
				Config:      testAccNetworkingRouterRoutesV2ExternalNextHop(projectID, region),
				ExpectError: regexp.MustCompile(`(?i)invalid|not connected|next.?hop`),
			},
			{
				// The rejected set stays in state until a refresh corrects it, and the
				// framework destroys without one in between: Delete would then remove
				// only the rejected entries and leave the real routes holding the
				// router's interface. Re-applying the set the rejection left in place
				// refreshes state, so state and router agree before the destroy.
				//
				// An empty configuration would not do: routes is Optional+Computed, so
				// leaving the blocks out means "keep what is there" (SCN-081-02), not
				// "remove everything".
				Config: testAccNetworkingRouterRoutesV2Pair(projectID, region),
				Check:  testAccCheckNetworkingRouterRoutesV2Count(2),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingRouterRoutesV2IDIsRouter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterRoutesV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterRoutesV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["router_id"] {
			return fmt.Errorf("resource id %s is not the router id %s",
				rs.Primary.ID, rs.Primary.Attributes["router_id"])
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRoutesV2Has(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if networkingRouterRouteV2Index(routes, router.Route{Destination: destination, NextHop: nextHop}) < 0 {
			return fmt.Errorf("route %s via %s is missing from the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRoutesV2HasNot(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if networkingRouterRouteV2Index(routes, router.Route{Destination: destination, NextHop: nextHop}) >= 0 {
			return fmt.Errorf("route %s via %s is still on the router", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRoutesV2Count(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routes, err := testAccNetworkingRouterRoutesV2Routes(s)
		if err != nil {
			return err
		}

		if len(routes) != expected {
			return fmt.Errorf("the router has %d routes, expected %d", len(routes), expected)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterRoutesV2AddOutOfBand(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterRoutesV2ResourceName)
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

// testAccNetworkingRouterRoutesV2Routes reads the router's actual routes. The
// resource identity is the router, so the handle points at it directly.
func testAccNetworkingRouterRoutesV2Routes(s *terraform.State) ([]router.Route, error) {
	handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterRoutesV2ResourceName)
	if err != nil {
		return nil, err
	}

	found, err := router.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.Routes, nil
}

// testAccCheckNetworkingRouterRoutesV2Destroy asserts the managed entries are
// gone. A router that no longer exists satisfies it: the routes went with it.
func testAccCheckNetworkingRouterRoutesV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_routes_v2" {
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

		found, err := router.Get(context.Background(), client, rs.Primary.ID)
		if err != nil {
			if networkingIsNotFound(err) {
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

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingRouterRoutesV2Empty(projectID, region string) string {
	return testAccNetworkingRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_routes_v2" "routes_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region)
}

func testAccNetworkingRouterRoutesV2Pair(projectID, region string) string {
	return testAccNetworkingRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
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

func testAccNetworkingRouterRoutesV2Replaced(projectID, region string) string {
	return testAccNetworkingRouterRouteV2Prerequisites(projectID, region) + fmt.Sprintf(`
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

// testAccNetworkingRouterRoutesV2ExternalNextHop declares one acceptable route and
// one whose next hop is only reachable through the external gateway port, so the
// whole set must be rejected.
// The bad next hop is an address allocated from the external network the router
// uses as its gateway: the subnets of an external network are unreadable by a
// project, so a floating IP is the only way to name one of their addresses.
func testAccNetworkingRouterRoutesV2ExternalNextHop(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_floatingip_v2" "next_hop_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
}

data "selectel_vpc_floatingip_v2" "next_hop_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  floatingip_id = selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.id
}

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
  project_id          = local.project_id
  region              = "%[1]s"
  name                = "router-tf-acc-test-1"
  external_network_id = data.selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.floating_network_id
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

  routes {
    destination_cidr = "192.0.2.0/24"
    next_hop         = selectel_vpc_floatingip_v2.next_hop_tf_acc_test_1.floating_ip_address
  }

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region)
}
