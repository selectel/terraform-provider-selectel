package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
)

const testAccVPCRouterInterfaceV2ResourceName = "selectel_vpc_router_interface_v2.interface_tf_acc_test_1"

func TestVPCRouterInterfaceV2SubnetFromPort(t *testing.T) {
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
				vpcRouterInterfaceV2SubnetFromPort(testCase.fixedIPs))
		})
	}
}

func TestAccVPCRouterInterfaceV2Subnet(t *testing.T) {
	var found port.Port
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterInterfaceV2Subnet(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterInterfaceV2Exists(&found),
					resource.TestCheckResourceAttrSet(testAccVPCRouterInterfaceV2ResourceName, "port_id"),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterInterfaceV2ResourceName, "subnet_id",
						"selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterInterfaceV2ResourceName, "network_id",
						"selectel_vpc_network_v2.network_tf_acc_test_1", "id",
					),
					testAccCheckVPCRouterInterfaceV2IDIsPort(),
				),
			},
			{
				Config:   testAccVPCRouterInterfaceV2Subnet(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:            testAccVPCRouterInterfaceV2ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccVPCRouterInterfaceV2Port(t *testing.T) {
	var found port.Port
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterInterfaceV2Port(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterInterfaceV2Exists(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterInterfaceV2ResourceName, "port_id",
						testAccVPCPortV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterInterfaceV2ResourceName, "subnet_id",
						"selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterInterfaceV2ResourceName, "network_id",
						"selectel_vpc_network_v2.network_tf_acc_test_1", "id",
					),
					testAccCheckVPCRouterInterfaceV2AttachedToRouter(&found),
				),
			},
			{
				Config:   testAccVPCRouterInterfaceV2Port(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCRouterInterfaceV2ForceDestroy(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterInterfaceV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterInterfaceV2ForceDestroy(projectID, region, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCRouterInterfaceV2ResourceName, "force_destroy", "false"),
					testAccCheckVPCRouterInterfaceV2AddBlockingRoute(),
				),
			},
			{
				Config:      testAccVPCRouterInterfaceV2ForceDestroy(projectID, region, false),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)timeout|conflict|in ?use|required by one or more routes`),
			},
			{
				Config: testAccVPCRouterInterfaceV2ForceDestroy(projectID, region, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCRouterInterfaceV2ResourceName, "force_destroy", "true"),
					testAccCheckVPCRouterInterfaceV2RouteStillPresent(),
				),
			},
		},
	})
}

func testAccCheckVPCRouterInterfaceV2Exists(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterInterfaceV2ResourceName)
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

func testAccCheckVPCRouterInterfaceV2IDIsPort() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRouterInterfaceV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRouterInterfaceV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["port_id"] {
			return fmt.Errorf("resource id %s is not the interface port %s",
				rs.Primary.ID, rs.Primary.Attributes["port_id"])
		}

		return nil
	}
}

func testAccCheckVPCRouterInterfaceV2AttachedToRouter(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRouterInterfaceV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRouterInterfaceV2ResourceName)
		}

		if found.DeviceID != rs.Primary.Attributes["router_id"] {
			return fmt.Errorf("port device_id is %q, expected the router %q",
				found.DeviceID, rs.Primary.Attributes["router_id"])
		}

		return nil
	}
}

var testAccVPCRouterInterfaceV2BlockingRoute = router.Route{
	Destination: "10.0.0.0/8",
	NextHop:     "192.168.199.10",
}

func testAccCheckVPCRouterInterfaceV2AddBlockingRoute() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, routerID, err := testAccVPCRouterInterfaceV2Router(s)
		if err != nil {
			return err
		}

		routes := []router.Route{testAccVPCRouterInterfaceV2BlockingRoute}
		_, err = router.Update(context.Background(), handle.client, routerID,
			router.UpdateRequest{Routes: &routes})

		return err
	}
}

func testAccCheckVPCRouterInterfaceV2RouteStillPresent() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, routerID, err := testAccVPCRouterInterfaceV2Router(s)
		if err != nil {
			return err
		}

		found, err := router.Get(context.Background(), handle.client, routerID)
		if err != nil {
			return err
		}

		if slices.Contains(found.Routes, testAccVPCRouterInterfaceV2BlockingRoute) {
			return nil
		}

		return fmt.Errorf("route %s via %s is gone before destroy ran",
			testAccVPCRouterInterfaceV2BlockingRoute.Destination,
			testAccVPCRouterInterfaceV2BlockingRoute.NextHop)
	}
}

func testAccVPCRouterInterfaceV2Router(s *terraform.State) (*vpcNetworkV2TestHandle, string, error) {
	rs, ok := s.RootModule().Resources[testAccVPCRouterInterfaceV2ResourceName]
	if !ok {
		return nil, "", fmt.Errorf("not found: %s", testAccVPCRouterInterfaceV2ResourceName)
	}

	handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterInterfaceV2ResourceName)
	if err != nil {
		return nil, "", err
	}

	return handle, rs.Primary.Attributes["router_id"], nil
}

func testAccCheckVPCRouterInterfaceV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_interface_v2" {
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

		if _, err := port.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC router interface still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCRouterInterfaceV2Base(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
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
}
`, region)
}

func testAccVPCRouterInterfaceV2Subnet(projectID, region string) string {
	return testAccVPCRouterInterfaceV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}
`, region)
}

func testAccVPCRouterInterfaceV2Port(projectID, region string) string {
	return testAccVPCRouterInterfaceV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.60"
  }
}

resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  port_id    = selectel_vpc_port_v2.port_tf_acc_test_1.id
}
`, region)
}

func testAccVPCRouterInterfaceV2ForceDestroy(projectID, region string, forceDestroy bool) string {
	return testAccVPCRouterInterfaceV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  router_id     = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id     = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  force_destroy = %[2]t
}
`, region, forceDestroy)
}

func TestVPCRouterInterfaceV2ConflictIsTransient(t *testing.T) {
	conflict := func(neutronType string) error {
		return &vpc.APIError{
			StatusCode: http.StatusConflict,
			Type:       neutronType,
			Message:    "test",
			Class:      vpc.ErrorClassConflict,
		}
	}

	t.Run("floating ip still bound through the interface", func(t *testing.T) {
		assert.True(t, vpcRouterInterfaceV2ConflictIsTransient(conflict("RouterInterfaceInUseByFloatingIP")))
	})

	t.Run("routes require the interface", func(t *testing.T) {
		assert.False(t, vpcRouterInterfaceV2ConflictIsTransient(conflict("RouterInterfaceInUseByRoute")),
			"routes never go away on their own; force_destroy is the way to clear them")
	})
}
