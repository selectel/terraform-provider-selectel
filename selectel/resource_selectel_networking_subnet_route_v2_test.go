package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingSubnetRouteV2ResourceName = "selectel_networking_subnet_route_v2.route_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingSubnetRouteV2Schema covers REQ-023: every attribute is required
// and replaces the resource, because the route has no update path of its own.
func TestNetworkingSubnetRouteV2Schema(t *testing.T) {
	s := resourceNetworkingSubnetRouteV2().Schema

	expected := []string{"project_id", "region", "subnet_id", "destination_cidr", "next_hop"}

	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}
	assert.ElementsMatch(t, expected, names, "the route has exactly these attributes")

	for _, name := range expected {
		t.Run(name+" is required and ForceNew", func(t *testing.T) {
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		})
	}

	// REQ-025: the schema does no local format validation of its own.
	t.Run("no local format validation", func(t *testing.T) {
		assert.Nil(t, s["destination_cidr"].ValidateFunc)
		assert.Nil(t, s["next_hop"].ValidateFunc)
	})

	// The resource has no update path at all.
	t.Run("no update function", func(t *testing.T) {
		assert.Nil(t, resourceNetworkingSubnetRouteV2().UpdateContext)
	})
}

// TestNetworkingSubnetRouteV2ID covers REQ-028: the composite identifier round
// trips, and a malformed one is rejected locally without touching the API.
func TestNetworkingSubnetRouteV2ID(t *testing.T) {
	subnetID := "8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b"
	route := subnet.HostRoute{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"}

	t.Run("round trips a well-formed identifier", func(t *testing.T) {
		id := networkingSubnetRouteV2ID(subnetID, route)
		assert.Equal(t, subnetID+"-route-10.0.0.0/8-192.168.199.10", id)

		gotSubnetID, gotRoute, err := parseNetworkingSubnetRouteV2ID(id)

		require.NoError(t, err)
		assert.Equal(t, subnetID, gotSubnetID)
		assert.Equal(t, route, gotRoute)
	})

	for _, c := range []struct {
		name string
		id   string
	}{
		{name: "no separator", id: subnetID + "/10.0.0.0/8/192.168.199.10"},
		{name: "empty subnet id", id: "-route-10.0.0.0/8-192.168.199.10"},
		{name: "one part after separator", id: subnetID + "-route-10.0.0.0/8"},
		{name: "three parts after separator", id: subnetID + "-route-10.0.0.0/8-192.168.199.10-extra"},
		{name: "empty destination", id: subnetID + "-route--192.168.199.10"},
		{name: "empty next hop", id: subnetID + "-route-10.0.0.0/8-"},
	} {
		t.Run("rejects "+c.name, func(t *testing.T) {
			_, _, err := parseNetworkingSubnetRouteV2ID(c.id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}
}

// TestNetworkingSubnetRouteV2Index covers REQ-024: a match needs both fields, so
// the same destination with another next hop is a different route.
func TestNetworkingSubnetRouteV2Index(t *testing.T) {
	routes := []subnet.HostRoute{
		{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"},
		{Destination: "172.16.0.0/12", NextHop: "192.168.199.11"},
	}

	assert.Equal(t, 0, networkingSubnetRouteV2Index(routes, routes[0]))
	assert.Equal(t, 1, networkingSubnetRouteV2Index(routes, routes[1]))

	t.Run("same destination with another next hop does not match", func(t *testing.T) {
		assert.Equal(t, -1, networkingSubnetRouteV2Index(routes, subnet.HostRoute{
			Destination: "10.0.0.0/8", NextHop: "192.168.199.99",
		}))
	})

	t.Run("same next hop with another destination does not match", func(t *testing.T) {
		assert.Equal(t, -1, networkingSubnetRouteV2Index(routes, subnet.HostRoute{
			Destination: "10.10.0.0/16", NextHop: "192.168.199.10",
		}))
	})

	t.Run("an empty list matches nothing", func(t *testing.T) {
		assert.Equal(t, -1, networkingSubnetRouteV2Index(nil, routes[0]))
	})
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingSubnetRouteV2Basic covers SCN-023-01, SCN-023-02 (empty
// follow-up plan) and SCN-028-01 (import by the composite identifier).
func TestAccNetworkingSubnetRouteV2Basic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetRouteV2ResourceName, "destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetRouteV2ResourceName, "next_hop", "192.168.199.10"),
				),
			},
			{
				Config:   testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingSubnetRouteV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingSubnetRouteV2ForceNew covers SCN-023-03: changing the route
// pair replaces the resource instead of updating it.
func TestAccNetworkingSubnetRouteV2ForceNew(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10"),
				Check:  testAccCheckNetworkingSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			},
			{
				Config: testAccNetworkingSubnetRouteV2Basic(projectName, "172.16.0.0/12", "192.168.199.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetRouteV2Exists("172.16.0.0/12", "192.168.199.10"),
					// The replaced route is gone from the subnet.
					testAccCheckNetworkingSubnetRouteV2Absent("10.0.0.0/8", "192.168.199.10"),
				),
			},
		},
	})
}

// TestAccNetworkingSubnetRouteV2Coexist covers SCN-023-01's requirement that the
// resource leaves other host_routes alone, and SCN-024-02: the same destination
// with a different next hop is a separate route, not a conflict.
func TestAccNetworkingSubnetRouteV2Coexist(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetRouteV2Pair(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					testAccCheckNetworkingSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.11"),
					testAccCheckNetworkingSubnetRouteV2Count(2),
				),
			},
			{
				Config:   testAccNetworkingSubnetRouteV2Pair(projectName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetRouteV2Conflict covers SCN-024-01: an exact duplicate
// of an existing entry is refused without touching host_routes.
func TestAccNetworkingSubnetRouteV2Conflict(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingSubnetRouteV2Duplicate(projectName),
				ExpectError: regexp.MustCompile(`already exists on subnet`),
			},
		},
	})
}

// TestAccNetworkingSubnetRouteV2InvalidValues covers SCN-025-01: the API, not the
// schema, rejects a malformed route, and no resource appears in the state.
func TestAccNetworkingSubnetRouteV2InvalidValues(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingSubnetRouteV2Basic(projectName, "not-a-cidr", "192.168.199.10"),
				ExpectError: regexp.MustCompile(`error creating networking subnet route`),
			},
		},
	})
}

// TestAccNetworkingSubnetRouteV2ExternalRemoval covers SCN-027-01: while the
// subnet exists, a route that disappeared outside Terraform is reported as an
// error rather than silently dropped from the state.
func TestAccNetworkingSubnetRouteV2ExternalRemoval(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10"),
				Check:  testAccCheckNetworkingSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			},
			{
				Config:      testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10"),
				Check:       testAccCheckNetworkingSubnetRouteV2RemoveOutOfBand("10.0.0.0/8", "192.168.199.10"),
				ExpectError: regexp.MustCompile(`no longer present on subnet`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Acceptance test helpers.
// -----------------------------------------------------------------------------

// testAccNetworkingSubnetRouteV2Subnet reads the parent subnet of the route
// under test.
func testAccNetworkingSubnetRouteV2Subnet(s *terraform.State) (*subnet.Subnet, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingSubnetV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccNetworkingSubnetV2ResourceName)
	}

	client, err := networkingV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, err
	}

	return subnet.Get(context.Background(), client, rs.Primary.ID)
}

func testAccCheckNetworkingSubnetRouteV2Exists(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := testAccNetworkingSubnetRouteV2Subnet(s)
		if err != nil {
			return err
		}

		route := subnet.HostRoute{Destination: destination, NextHop: nextHop}
		if networkingSubnetRouteV2Index(found.HostRoutes, route) < 0 {
			return fmt.Errorf("route %s via %s is missing from the subnet", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckNetworkingSubnetRouteV2Absent(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := testAccNetworkingSubnetRouteV2Subnet(s)
		if err != nil {
			return err
		}

		route := subnet.HostRoute{Destination: destination, NextHop: nextHop}
		if networkingSubnetRouteV2Index(found.HostRoutes, route) >= 0 {
			return fmt.Errorf("route %s via %s is still present on the subnet", destination, nextHop)
		}

		return nil
	}
}

// testAccCheckNetworkingSubnetRouteV2Count asserts the resource did not disturb
// the other entries of host_routes.
func testAccCheckNetworkingSubnetRouteV2Count(want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := testAccNetworkingSubnetRouteV2Subnet(s)
		if err != nil {
			return err
		}

		if len(found.HostRoutes) != want {
			return fmt.Errorf("the subnet has %d host routes, expected %d", len(found.HostRoutes), want)
		}

		return nil
	}
}

// testAccCheckNetworkingSubnetRouteV2RemoveOutOfBand drops the route from the
// subnet without Terraform, producing the disappearance SCN-027-01 describes.
func testAccCheckNetworkingSubnetRouteV2RemoveOutOfBand(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingSubnetV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingSubnetV2ResourceName)
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		found, err := subnet.Get(context.Background(), client, rs.Primary.ID)
		if err != nil {
			return err
		}

		route := subnet.HostRoute{Destination: destination, NextHop: nextHop}
		index := networkingSubnetRouteV2Index(found.HostRoutes, route)
		if index < 0 {
			return errors.New("the route to remove out of band is already absent")
		}

		routes := make([]subnet.HostRoute, 0, len(found.HostRoutes)-1)
		routes = append(routes, found.HostRoutes[:index]...)
		routes = append(routes, found.HostRoutes[index+1:]...)

		_, err = subnet.Update(context.Background(), client, rs.Primary.ID, subnet.UpdateRequest{
			HostRoutes: &routes,
		})

		return err
	}
}

func testAccCheckNetworkingSubnetRouteV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_subnet_route_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return nil //nolint:nilerr // the project is gone together with the subnet
		}

		subnetID, route, err := parseNetworkingSubnetRouteV2ID(rs.Primary.ID)
		if err != nil {
			return err
		}

		found, err := subnet.Get(context.Background(), client, subnetID)
		if err != nil {
			if networkingIsNotFound(err) {
				continue
			}

			return err
		}

		if networkingSubnetRouteV2Index(found.HostRoutes, route) >= 0 {
			return errors.New("networking subnet route still exists")
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Acceptance test configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingSubnetRouteV2Basic(projectName, destination, nextHop string) string {
	return testAccNetworkingSubnetV2Basic(projectName, "192.168.199.0/24") + fmt.Sprintf(`
resource "selectel_networking_subnet_route_v2" "route_tf_acc_test_1" {
  project_id       = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region           = "%s"
  subnet_id        = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "%s"
  next_hop         = "%s"
}
`, networkingV2TestRegion(), destination, nextHop)
}

func testAccNetworkingSubnetRouteV2Pair(projectName string) string {
	return testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10") + fmt.Sprintf(`
resource "selectel_networking_subnet_route_v2" "route_tf_acc_test_2" {
  project_id       = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region           = "%s"
  subnet_id        = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.11"

  depends_on = [selectel_networking_subnet_route_v2.route_tf_acc_test_1]
}
`, networkingV2TestRegion())
}

func testAccNetworkingSubnetRouteV2Duplicate(projectName string) string {
	return testAccNetworkingSubnetRouteV2Basic(projectName, "10.0.0.0/8", "192.168.199.10") + fmt.Sprintf(`
resource "selectel_networking_subnet_route_v2" "route_tf_acc_test_duplicate" {
  project_id       = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region           = "%s"
  subnet_id        = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"

  depends_on = [selectel_networking_subnet_route_v2.route_tf_acc_test_1]
}
`, networkingV2TestRegion())
}
