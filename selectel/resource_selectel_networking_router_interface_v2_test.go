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
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingRouterInterfaceV2ResourceName = "selectel_networking_router_interface_v2.interface_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingRouterInterfaceV2Schema covers REQ-067 and REQ-068: the two
// addressing attributes are mutually exclusive, everything that identifies the
// interface replaces it, and force_destroy does not.
func TestNetworkingRouterInterfaceV2Schema(t *testing.T) {
	s := resourceNetworkingRouterInterfaceV2().Schema

	t.Run("project_id, region and router_id are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region", "router_id"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-067: whichever of the two is omitted is computed from the attached port.
	t.Run("subnet_id and port_id are optional, computed and ForceNew", func(t *testing.T) {
		for _, name := range []string{"subnet_id", "port_id"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Optional)
			assert.True(t, s[name].Computed)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-068: the pair is rejected at plan time rather than at apply.
	t.Run("subnet_id and port_id conflict with each other", func(t *testing.T) {
		assert.Equal(t, []string{"port_id"}, s["subnet_id"].ConflictsWith)
		assert.Equal(t, []string{"subnet_id"}, s["port_id"].ConflictsWith)
	})

	// REQ-072: force_destroy steers Delete, so changing it must not replace the
	// interface.
	t.Run("force_destroy is optional, defaults to false and is not ForceNew", func(t *testing.T) {
		require.Contains(t, s, "force_destroy")
		assert.True(t, s["force_destroy"].Optional)
		assert.Equal(t, false, s["force_destroy"].Default)
		assert.False(t, s["force_destroy"].ForceNew)
	})

	// REQ-069: there is no updatable remote attribute, so the follow-up plan is
	// empty by construction — the Update exists only for force_destroy.
	t.Run("no attribute other than force_destroy is updatable", func(t *testing.T) {
		for name, attr := range s {
			if name == "force_destroy" {
				continue
			}
			assert.True(t, attr.ForceNew, name+" must replace the interface when it changes")
		}
	})
}

// TestNetworkingRouterInterfaceV2SubnetFromPort covers the subnet_id computation
// of REQ-067 and REQ-071: exactly one fixed IP resolves the subnet, and anything
// else leaves it unset.
func TestNetworkingRouterInterfaceV2SubnetFromPort(t *testing.T) {
	cases := []struct {
		name     string
		fixedIPs []port.FixedIP
		expected string
	}{
		{
			name:     "a single fixed ip resolves the subnet",
			fixedIPs: []port.FixedIP{{SubnetID: "subnet-id", IPAddress: "192.168.199.1"}},
			expected: "subnet-id",
		},
		{
			// SCN-071-02: a dual-stack port has no single subnet to report.
			name: "two fixed ips leave the subnet unset",
			fixedIPs: []port.FixedIP{
				{SubnetID: "subnet-v4", IPAddress: "192.168.199.1"},
				{SubnetID: "subnet-v6", IPAddress: "2001:db8::1"},
			},
			expected: "",
		},
		{
			name:     "no fixed ip leaves the subnet unset",
			fixedIPs: nil,
			expected: "",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected,
				networkingRouterInterfaceV2SubnetFromPort(testCase.fixedIPs))
		})
	}
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingRouterInterfaceV2Subnet covers SCN-067-01, SCN-069-01,
// SCN-071-01 and SCN-072-01.
func TestAccNetworkingRouterInterfaceV2Subnet(t *testing.T) {
	var found port.Port
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterInterfaceV2Subnet(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterInterfaceV2Exists(&found),
					// SCN-067-01: the API creates the port, and its id is the
					// identity of the resource.
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterInterfaceV2ResourceName, "port_id"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRouterInterfaceV2ResourceName, "subnet_id",
						"selectel_networking_subnet_v2.subnet_tf_acc_test_1", "id",
					),
					testAccCheckNetworkingRouterInterfaceV2IDIsPort(),
				),
			},
			{
				// SCN-069-01: the follow-up plan is empty.
				Config:   testAccNetworkingRouterInterfaceV2Subnet(projectName),
				PlanOnly: true,
			},
			{
				// SCN-071-01: import by the port id alone, no composite format.
				ResourceName:      testAccNetworkingRouterInterfaceV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// force_destroy steers this resource's Delete and has no remote
				// counterpart to import.
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

// TestAccNetworkingRouterInterfaceV2Port covers SCN-067-02: attaching through a
// pre-created port, with subnet_id computed from it.
func TestAccNetworkingRouterInterfaceV2Port(t *testing.T) {
	var found port.Port
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterInterfaceV2Port(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterInterfaceV2Exists(&found),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRouterInterfaceV2ResourceName, "port_id",
						testAccNetworkingPortV2ResourceName, "id",
					),
					// subnet_id is computed from the port's single fixed IP.
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRouterInterfaceV2ResourceName, "subnet_id",
						"selectel_networking_subnet_v2.subnet_tf_acc_test_1", "id",
					),
					testAccCheckNetworkingRouterInterfaceV2AttachedToRouter(&found),
				),
			},
			{
				Config:   testAccNetworkingRouterInterfaceV2Port(projectName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingRouterInterfaceV2PortFixedIPCount covers SCN-067-03: a port
// with two fixed IPs is refused before the router is touched.
func TestAccNetworkingRouterInterfaceV2PortFixedIPCount(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingRouterInterfaceV2PortTwoFixedIPs(projectName),
				ExpectError: regexp.MustCompile(`has 2 fixed ips, a router interface requires exactly one`),
			},
		},
	})
}

// TestAccNetworkingRouterInterfaceV2Exclusivity covers SCN-068-01 and SCN-068-02.
func TestAccNetworkingRouterInterfaceV2Exclusivity(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-068-01: rejected by the schema, before any API call.
				Config:      testAccNetworkingRouterInterfaceV2Both(projectName),
				ExpectError: regexp.MustCompile(`(?s)"subnet_id".*conflicts with.*"port_id"|"port_id".*conflicts with.*"subnet_id"`),
			},
			{
				// SCN-068-02: neither one set.
				Config:      testAccNetworkingRouterInterfaceV2Neither(projectName),
				ExpectError: regexp.MustCompile(`exactly one of subnet_id or port_id must be set`),
			},
		},
	})
}

// TestAccNetworkingRouterInterfaceV2ExternalDetach covers SCN-070-01: an
// interface detached outside Terraform leaves the state without an error.
func TestAccNetworkingRouterInterfaceV2ExternalDetach(t *testing.T) {
	var found port.Port
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterInterfaceV2Subnet(projectName),
				Check:  testAccCheckNetworkingRouterInterfaceV2Exists(&found),
			},
			{
				Config:             testAccNetworkingRouterInterfaceV2Subnet(projectName),
				Check:              testAccCheckNetworkingRouterInterfaceV2DetachOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				// The plan re-creates the interface rather than updating it.
				Config: testAccNetworkingRouterInterfaceV2Subnet(projectName),
				Check:  testAccCheckNetworkingRouterInterfaceV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingRouterInterfaceV2ForceDestroy covers SCN-072-02 and
// SCN-072-03: a static route pointing into the attached subnet blocks the
// detach, and force_destroy clears exactly those routes.
//
// The blocking route is added through the SDK rather than through
// selectel_networking_router_route_v2 so this test does not depend on another
// resource's behaviour, and the delete timeout is shortened in the configuration
// so the blocked case fails in seconds instead of minutes.
func TestAccNetworkingRouterInterfaceV2ForceDestroy(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				// force_destroy=false, plus a route whose next hop is inside the
				// attached subnet.
				Config: testAccNetworkingRouterInterfaceV2ForceDestroy(projectName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkingRouterInterfaceV2ResourceName, "force_destroy", "false"),
					testAccCheckNetworkingRouterInterfaceV2AddBlockingRoute(),
				),
			},
			{
				// SCN-072-02: the detach is retried until the delete timeout and
				// then reported as an error; the route is not removed.
				Config:      testAccNetworkingRouterInterfaceV2ForceDestroy(projectName, false),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)timeout|conflict|in use|still in use`),
			},
			{
				// SCN-072-03: force_destroy changes in place — it steers Delete and
				// has no remote counterpart — and the following destroy clears the
				// conflicting route and succeeds.
				Config: testAccNetworkingRouterInterfaceV2ForceDestroy(projectName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkingRouterInterfaceV2ResourceName, "force_destroy", "true"),
					testAccCheckNetworkingRouterInterfaceV2RouteStillPresent(),
				),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingRouterInterfaceV2Exists(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterInterfaceV2ResourceName)
		if err != nil {
			return err
		}

		result, err := port.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

// testAccCheckNetworkingRouterInterfaceV2IDIsPort asserts REQ-067's identity
// rule: the resource id is the port the router is attached through.
func testAccCheckNetworkingRouterInterfaceV2IDIsPort() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterInterfaceV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterInterfaceV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["port_id"] {
			return fmt.Errorf("resource id %s is not the interface port %s",
				rs.Primary.ID, rs.Primary.Attributes["port_id"])
		}

		return nil
	}
}

func testAccCheckNetworkingRouterInterfaceV2AttachedToRouter(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterInterfaceV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterInterfaceV2ResourceName)
		}

		if found.DeviceID != rs.Primary.Attributes["router_id"] {
			return fmt.Errorf("port device_id is %q, expected the router %q",
				found.DeviceID, rs.Primary.Attributes["router_id"])
		}

		return nil
	}
}

func testAccCheckNetworkingRouterInterfaceV2DetachOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterInterfaceV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterInterfaceV2ResourceName)
		}

		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterInterfaceV2ResourceName)
		if err != nil {
			return err
		}

		_, err = router.RemoveInterface(
			context.Background(), handle.client,
			rs.Primary.Attributes["router_id"], router.ByPort(rs.Primary.ID))

		return err
	}
}

// testAccNetworkingRouterInterfaceV2BlockingRoute is the route whose next hop
// lies inside the attached subnet, which is what makes Neutron refuse the detach.
var testAccNetworkingRouterInterfaceV2BlockingRoute = router.Route{
	Destination: "10.0.0.0/8",
	NextHop:     "192.168.199.10",
}

func testAccCheckNetworkingRouterInterfaceV2AddBlockingRoute() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, routerID, err := testAccNetworkingRouterInterfaceV2Router(s)
		if err != nil {
			return err
		}

		routes := []router.Route{testAccNetworkingRouterInterfaceV2BlockingRoute}
		_, err = router.Update(context.Background(), handle.client, routerID,
			router.UpdateRequest{Routes: &routes})

		return err
	}
}

// testAccCheckNetworkingRouterInterfaceV2RouteStillPresent asserts the route is
// untouched while force_destroy has only been flipped: it is cleared by Delete,
// not by the update that sets the flag.
func testAccCheckNetworkingRouterInterfaceV2RouteStillPresent() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, routerID, err := testAccNetworkingRouterInterfaceV2Router(s)
		if err != nil {
			return err
		}

		found, err := router.Get(context.Background(), handle.client, routerID)
		if err != nil {
			return err
		}

		for _, route := range found.Routes {
			if route == testAccNetworkingRouterInterfaceV2BlockingRoute {
				return nil
			}
		}

		return fmt.Errorf("route %s via %s is gone before destroy ran",
			testAccNetworkingRouterInterfaceV2BlockingRoute.Destination,
			testAccNetworkingRouterInterfaceV2BlockingRoute.NextHop)
	}
}

// testAccNetworkingRouterInterfaceV2Router returns a client handle plus the id of
// the router the interface under test is attached to.
func testAccNetworkingRouterInterfaceV2Router(s *terraform.State) (*networkingNetworkV2TestHandle, string, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingRouterInterfaceV2ResourceName]
	if !ok {
		return nil, "", fmt.Errorf("not found: %s", testAccNetworkingRouterInterfaceV2ResourceName)
	}

	handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterInterfaceV2ResourceName)
	if err != nil {
		return nil, "", err
	}

	return handle, rs.Primary.Attributes["router_id"], nil
}

func testAccCheckNetworkingRouterInterfaceV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_router_interface_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return nil //nolint:nilerr // the project is gone together with the interface
		}

		if _, err := port.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking router interface still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingRouterInterfaceV2Base is the network, subnet and router every
// interface configuration attaches together.
func testAccNetworkingRouterInterfaceV2Base(projectName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}

resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  name       = "router-tf-acc-test-1"
}
`, networkingV2TestRegion())
}

func testAccNetworkingRouterInterfaceV2Subnet(projectName string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
}
`, networkingV2TestRegion())
}

func testAccNetworkingRouterInterfaceV2Port(projectName string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.60"
  }
}

resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
  port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
}
`, networkingV2TestRegion())
}

func testAccNetworkingRouterInterfaceV2PortTwoFixedIPs(projectName string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.61"
  }

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.62"
  }
}

resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
  port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
}
`, networkingV2TestRegion())
}

func testAccNetworkingRouterInterfaceV2Both(projectName string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.63"
  }
}

resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
}
`, networkingV2TestRegion())
}

func testAccNetworkingRouterInterfaceV2Neither(projectName string) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
}
`, networkingV2TestRegion())
}

// testAccNetworkingRouterInterfaceV2ForceDestroy shortens the delete timeout so
// the blocked-detach case fails fast instead of retrying for the default period.
func testAccNetworkingRouterInterfaceV2ForceDestroy(projectName string, forceDestroy bool) string {
	return testAccNetworkingRouterInterfaceV2Base(projectName) + fmt.Sprintf(`
resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id    = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region        = "%[1]s"
  router_id     = selectel_networking_router_v2.router_tf_acc_test_1.id
  subnet_id     = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  force_destroy = %[2]t

  timeouts {
    delete = "45s"
  }
}
`, networkingV2TestRegion(), forceDestroy)
}
