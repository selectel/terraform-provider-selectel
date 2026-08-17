package selectel

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCNetworkSubnetRouteV2ResourceName = "selectel_vpc_network_subnet_route_v2.route_tf_acc_test_1"

func TestVPCNetworkSubnetRouteV2ID(t *testing.T) {
	subnetID := "8f6c1e9b-2a3d-4f50-9b8a-1c2d3e4f5a6b"
	route := subnet.HostRoute{Destination: "10.0.0.0/8", NextHop: "192.168.199.10"}

	t.Run("round trips a well-formed identifier", func(t *testing.T) {
		id := vpcNetworkSubnetRouteV2ID(subnetID, route)
		assert.Equal(t, subnetID+"-route-10.0.0.0/8-192.168.199.10", id)

		gotSubnetID, gotRoute, err := parseVPCNetworkSubnetRouteV2ID(id)

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
			_, _, err := parseVPCNetworkSubnetRouteV2ID(c.id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}
}

func TestAccVPCNetworkSubnetRouteV2Basic(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetRouteV2Basic(projectID, region, "10.0.0.0/8", "192.168.199.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetRouteV2ResourceName, "destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetRouteV2ResourceName, "next_hop", "192.168.199.10"),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetRouteV2Basic(projectID, region, "10.0.0.0/8", "192.168.199.10"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCNetworkSubnetRouteV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetRouteV2Coexist(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetRouteV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetRouteV2Pair(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
					testAccCheckVPCNetworkSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.11"),
					testAccCheckVPCNetworkSubnetRouteV2Count(2),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetRouteV2Pair(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

func testAccVPCNetworkSubnetRouteV2Subnet(s *terraform.State) (*subnet.Subnet, error) {
	rs, ok := s.RootModule().Resources[testAccVPCNetworkSubnetV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccVPCNetworkSubnetV2ResourceName)
	}

	client, err := vpcV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, err
	}

	return subnet.Get(context.Background(), client, rs.Primary.ID)
}

func testAccCheckVPCNetworkSubnetRouteV2Exists(destination, nextHop string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := testAccVPCNetworkSubnetRouteV2Subnet(s)
		if err != nil {
			return err
		}

		route := subnet.HostRoute{Destination: destination, NextHop: nextHop}
		if !slices.Contains(found.HostRoutes, route) {
			return fmt.Errorf("route %s via %s is missing from the subnet", destination, nextHop)
		}

		return nil
	}
}

func testAccCheckVPCNetworkSubnetRouteV2Count(want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found, err := testAccVPCNetworkSubnetRouteV2Subnet(s)
		if err != nil {
			return err
		}

		if len(found.HostRoutes) != want {
			return fmt.Errorf("the subnet has %d host routes, expected %d", len(found.HostRoutes), want)
		}

		return nil
	}
}

func testAccCheckVPCNetworkSubnetRouteV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_network_subnet_route_v2" {
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

		subnetID, route, err := parseVPCNetworkSubnetRouteV2ID(rs.Primary.ID)
		if err != nil {
			return err
		}

		found, err := subnet.Get(context.Background(), client, subnetID)
		if err != nil {
			if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
				continue
			}

			return err
		}

		if slices.Contains(found.HostRoutes, route) {
			return errors.New("VPC network subnet route still exists")
		}
	}

	return nil
}

func testAccVPCNetworkSubnetRouteV2Basic(projectID, region, destination, nextHop string) string {
	return testAccVPCNetworkSubnetV2Basic(projectID, region, "192.168.199.0/24") + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_route_v2" "route_tf_acc_test_1" {
  project_id       = local.project_id
  region           = "%s"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "%s"
  next_hop         = "%s"
}
`, region, destination, nextHop)
}

func testAccVPCNetworkSubnetRouteV2Pair(projectID, region string) string {
	return testAccVPCNetworkSubnetRouteV2Basic(projectID, region, "10.0.0.0/8", "192.168.199.10") + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_route_v2" "route_tf_acc_test_2" {
  project_id       = local.project_id
  region           = "%s"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.11"

  depends_on = [selectel_vpc_network_subnet_route_v2.route_tf_acc_test_1]
}
`, region)
}
