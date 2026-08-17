package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCNetworkSubnetV2ResourceName = "selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1"

func TestVPCNetworkSubnetV2CIDRValidation(t *testing.T) {
	for _, c := range []struct {
		value   string
		wantErr string
	}{
		{value: "192.168.0.0/24"},
		{value: "2001:db8::/64"},
		{value: ""},
		{value: "not-a-cidr", wantErr: "not a valid CIDR"},
		{value: "192.168.0.0", wantErr: "not a valid CIDR"},
		{value: "192.168.0.17/24", wantErr: "must be the network address"},
	} {
		t.Run(c.value, func(t *testing.T) {
			_, errs := validateVPCNetworkSubnetV2CIDR(c.value, "cidr")

			if c.wantErr == "" {
				assert.Empty(t, errs)

				return
			}

			require.NotEmpty(t, errs)
			assert.Contains(t, errs[0].Error(), c.wantErr)
		})
	}
}

func TestVPCNetworkSubnetV2AllocationPools(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"start": "192.168.0.10", "end": "192.168.0.20"},
	}

	pools := expandVPCNetworkSubnetV2AllocationPools(raw)

	require.Len(t, pools, 1)
	assert.Equal(t, subnet.AllocationPool{Start: "192.168.0.10", End: "192.168.0.20"}, pools[0])
	assert.Equal(t, raw, flattenVPCNetworkSubnetV2AllocationPools(pools))
}

func TestVPCNetworkSubnetV2DataSourceSchema(t *testing.T) {
	dataSource := dataSourceVPCNetworkSubnetV2()

	_, hasDeprecatedDHCPSelector := dataSource.Schema["dhcp_enabled"]
	assert.False(t, hasDeprecatedDHCPSelector)

	enableDHCP := dataSource.Schema["enable_dhcp"]
	require.NotNil(t, enableDHCP)
	assert.True(t, enableDHCP.Optional)
	assert.True(t, enableDHCP.Computed)

	_, hasDeprecatedAllocationPools := dataSource.Schema["allocation_pools"]
	assert.False(t, hasDeprecatedAllocationPools)

	allocationPool := dataSource.Schema["allocation_pool"]
	require.NotNil(t, allocationPool)
	assert.Equal(t, schema.TypeSet, allocationPool.Type)
	assert.True(t, allocationPool.Computed)
}

func TestVPCNetworkSubnetV2CollectionSchema(t *testing.T) {
	for name, resourceSchema := range map[string]map[string]*schema.Schema{
		"resource":    resourceVPCNetworkSubnetV2().Schema,
		"data source": dataSourceVPCNetworkSubnetV2().Schema,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, schema.TypeList, resourceSchema["dns_nameservers"].Type)
			assert.Equal(t, schema.TypeSet, resourceSchema["service_types"].Type)
		})
	}
}

func TestVPCNetworkSubnetV2HostRoutes(t *testing.T) {
	routes := []subnet.HostRoute{{Destination: "10.0.0.0/8", NextHop: "192.168.0.1"}}
	want := []interface{}{
		map[string]interface{}{"destination_cidr": "10.0.0.0/8", "next_hop": "192.168.0.1"},
	}

	assert.Equal(t, want, flattenVPCNetworkSubnetV2HostRoutes(routes))

	d := schema.TestResourceDataRaw(t, resourceVPCNetworkSubnetV2().Schema, nil)
	flattenVPCNetworkSubnetV2(d, &subnet.Subnet{HostRoutes: routes})
	assert.Equal(t, want, d.Get("host_routes"))
}

func TestAccVPCNetworkSubnetV2Basic(t *testing.T) {
	var found subnet.Subnet
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2Exists(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "cidr", "192.168.199.0/24"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "ip_version", "4"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "enable_dhcp", "true"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkSubnetV2ResourceName, "allocation_pool.#"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkSubnetV2ResourceName, "blocked"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkSubnetV2ResourceName, "host_routes.#"),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCNetworkSubnetV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetV2OptionalAttributes(t *testing.T) {
	var found subnet.Subnet
	projectID, region := vpcV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2Exists(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "name", subnetName),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "enable_dhcp", "false"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "dns_nameservers.#", "2"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "allocation_pool.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip", "192.168.199.1"),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetV2Update(t *testing.T) {
	var found subnet.Subnet
	projectID, region := vpcV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")
	updatedName := acctest.RandomWithPrefix("tf-acc-subnet-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2Exists(testAccVPCNetworkSubnetV2ResourceName, &found),
				),
			},
			{
				Config: testAccVPCNetworkSubnetV2Updated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2SameID(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "enable_dhcp", "true"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "dns_nameservers.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip", "192.168.199.2"),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2Updated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "name", subnetName),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "enable_dhcp", "false"),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "allocation_pool.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "subnet_id",
						testAccVPCNetworkSubnetV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.#", "1"),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.0.destination_cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.0.next_hop", "192.168.199.10"),
					resource.TestCheckResourceAttrPair(
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "blocked",
						testAccVPCNetworkSubnetV2ResourceName, "blocked",
					),
				),
			},
			{
				Config: testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "host_routes.#", "1"),
					resource.TestCheckResourceAttrPair(
						testAccVPCNetworkSubnetV2ResourceName, "host_routes.#",
						"data.selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1", "host_routes.#",
					),
				),
			},
			{
				Config:      testAccVPCNetworkSubnetV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC network subnet found`),
			},
		},
	})
}

func testAccVPCNetworkSubnetV2Handle(s *terraform.State) (*vpcNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[testAccVPCNetworkSubnetV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccVPCNetworkSubnetV2ResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, errors.New("no ID is set")
	}

	client, err := vpcV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, err
	}

	return &vpcNetworkV2TestHandle{client: client, id: rs.Primary.ID}, nil
}

func testAccCheckVPCNetworkSubnetV2Exists(_ string, found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkSubnetV2Handle(s)
		if err != nil {
			return err
		}

		result, err := subnet.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if result.ID != handle.id {
			return errors.New("VPC network subnet not found")
		}

		*found = *result

		return nil
	}
}

func testAccCheckVPCNetworkSubnetV2SameID(resourceName string, found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC network subnet was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCNetworkSubnetV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_network_subnet_v2" {
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

		if _, err := subnet.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC network subnet still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCNetworkSubnetV2Network(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-subnet-parent"
}
`, region)
}

func testAccVPCNetworkSubnetV2Basic(projectID, region, cidr string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "%s"
}
`, region, cidr)
}

func testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  network_id           = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr                 = "192.168.199.0/24"
  name                 = "%s"
  description          = "created by the acceptance test"
  gateway_ip           = "192.168.199.1"
  enable_dhcp          = false
  dns_nameservers      = ["8.8.8.8", "1.1.1.1"]
  tags                 = ["tf-acc"]

  allocation_pool {
    start = "192.168.199.10"
    end   = "192.168.199.20"
  }
}
`, region, subnetName)
}

func testAccVPCNetworkSubnetV2Updated(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  network_id           = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr                 = "192.168.199.0/24"
  name                 = "%s"
  description          = "updated by the acceptance test"
  gateway_ip           = "192.168.199.2"
  enable_dhcp          = true
  dns_nameservers      = ["8.8.8.8"]
  tags                 = ["tf-acc", "updated"]

  allocation_pool {
    start = "192.168.199.30"
    end   = "192.168.199.40"
  }
}
`, region, subnetName)
}

func testAccVPCNetworkSubnetV2DataSource(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2OptionalAttributes(projectID, region, subnetName) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_route_v2" "route_tf_acc_test_1" {
  project_id       = local.project_id
  region           = "%s"
  subnet_id        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  destination_cidr = "10.0.0.0/8"
  next_hop         = "192.168.199.10"
}

data "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  name        = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.name
  enable_dhcp = false

  depends_on = [selectel_vpc_network_subnet_route_v2.route_tf_acc_test_1]
}
`, region, region)
}

func testAccVPCNetworkSubnetV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
data "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-subnet"
}
`, region)
}
