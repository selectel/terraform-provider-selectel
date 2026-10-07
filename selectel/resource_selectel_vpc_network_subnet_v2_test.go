package selectel

import (
	"context"
	"errors"
	"fmt"
	"maps"
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
	raw := []any{
		map[string]any{"start": "192.168.0.10", "end": "192.168.0.20"},
	}

	pools := expandVPCNetworkSubnetV2AllocationPools(raw)

	require.Len(t, pools, 1)
	assert.Equal(t, subnet.AllocationPool{Start: "192.168.0.10", End: "192.168.0.20"}, pools[0])
	assert.Equal(t, raw, flattenVPCNetworkSubnetV2AllocationPools(pools))
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
	want := []any{
		map[string]any{"destination_cidr": "10.0.0.0/8", "next_hop": "192.168.0.1"},
	}

	assert.Equal(t, want, flattenVPCNetworkSubnetV2HostRoutes(routes))

	d := schema.TestResourceDataRaw(t, resourceVPCNetworkSubnetV2().Schema, nil)
	flattenVPCNetworkSubnetV2(d, &subnet.Subnet{HostRoutes: routes})
	assert.Equal(t, want, d.Get("host_routes"))
}

func TestVPCNetworkSubnetV2PlanNoGateway(t *testing.T) {
	state := func(noGateway, gatewayIP string) *terraform.InstanceState {
		return &terraform.InstanceState{
			ID: "subnet-id",
			Attributes: map[string]string{
				"id":         "subnet-id",
				"project_id": "project-id",
				"region":     "ru-3",
				"network_id": "network-id",
				"cidr":       "192.168.199.0/24",
				"no_gateway": noGateway,
				"gateway_ip": gatewayIP,
			},
		}
	}
	config := func(extra map[string]any) *terraform.ResourceConfig {
		raw := map[string]any{
			"project_id": "project-id",
			"region":     "ru-3",
			"network_id": "network-id",
			"cidr":       "192.168.199.0/24",
		}
		maps.Copy(raw, extra)

		return terraform.NewResourceConfigRaw(raw)
	}

	for name, c := range map[string]struct {
		state   *terraform.InstanceState
		config  *terraform.ResourceConfig
		wantErr string
	}{
		"no_gateway true to false without gateway_ip is rejected": {
			state:   state("true", ""),
			config:  config(map[string]any{"no_gateway": false}),
			wantErr: "gateway_ip",
		},
		"no_gateway attribute removed without gateway_ip is rejected": {
			state:   state("true", ""),
			config:  config(nil),
			wantErr: "gateway_ip",
		},
		"no_gateway true to false with gateway_ip is allowed": {
			state:  state("true", ""),
			config: config(map[string]any{"no_gateway": false, "gateway_ip": "192.168.199.1"}),
		},
		"no_gateway stays true is allowed": {
			state:  state("true", ""),
			config: config(map[string]any{"no_gateway": true}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := resourceVPCNetworkSubnetV2().Diff(context.Background(), c.state, c.config, nil)

			if c.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), c.wantErr)
		})
	}
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
			{
				Config:             testAccVPCNetworkSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				Check:              testAccCheckVPCV2DeleteOutOfBand(testAccVPCNetworkSubnetV2ResourceName, subnet.Delete),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetV2IPv6(t *testing.T) {
	var found subnet.Subnet
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkSubnetV2IPv6(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2Exists(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "ip_version", "6"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "cidr", "fd00:c0de:199::/64"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "ipv6_address_mode", "dhcpv6-stateful"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "ipv6_ra_mode", "dhcpv6-stateful"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "no_gateway", "true"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip", ""),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2IPv6(projectID, region),
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
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "all_tags.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkSubnetV2NoGateway(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2SameID(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "no_gateway", "true"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip", ""),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2NoGateway(projectID, region, updatedName),
				PlanOnly: true,
			},
			{
				Config:      testAccVPCNetworkSubnetV2GatewayEnabled(projectID, region, updatedName, ""),
				ExpectError: regexp.MustCompile("`gateway_ip` must be set"),
			},
			{
				Config: testAccVPCNetworkSubnetV2GatewayEnabled(projectID, region, updatedName, "192.168.199.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2SameID(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "no_gateway", "false"),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "gateway_ip", "192.168.199.1"),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2GatewayEnabled(projectID, region, updatedName, "192.168.199.1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkSubnetV2TagsOutOfBand(t *testing.T) {
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
					testAccCheckVPCTagsV2AddOutOfBand(testAccVPCNetworkSubnetV2TagOperations, "foreign"),
				),
			},
			{
				Config: testAccVPCNetworkSubnetV2Updated(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkSubnetV2SameID(testAccVPCNetworkSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkSubnetV2ResourceName, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(testAccVPCNetworkSubnetV2ResourceName, "all_tags.*", "foreign"),
					testAccCheckVPCTagsV2Match(testAccVPCNetworkSubnetV2TagOperations, []string{"tf-acc", "updated", "foreign"}),
				),
			},
			{
				Config:   testAccVPCNetworkSubnetV2Updated(projectID, region, subnetName),
				PlanOnly: true,
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

func testAccVPCNetworkSubnetV2TagOperations(s *terraform.State) (vpcTagsV2AccOperations, error) {
	handle, err := testAccVPCNetworkSubnetV2Handle(s)
	if err != nil {
		return nil, err
	}

	return subnet.TagOperations(handle.client, handle.id), nil
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

func testAccVPCNetworkSubnetV2NoGateway(projectID, region, subnetName string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id      = local.project_id
  region          = "%s"
  network_id      = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr            = "192.168.199.0/24"
  name            = "%s"
  description     = "updated by the acceptance test"
  no_gateway      = true
  enable_dhcp     = true
  dns_nameservers = ["8.8.8.8"]
  tags            = ["tf-acc", "updated"]

  allocation_pool {
    start = "192.168.199.30"
    end   = "192.168.199.40"
  }
}
`, region, subnetName)
}

func testAccVPCNetworkSubnetV2GatewayEnabled(projectID, region, subnetName, gatewayIP string) string {
	gateway := ""
	if gatewayIP != "" {
		gateway = fmt.Sprintf("gateway_ip      = %q", gatewayIP)
	}

	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id      = local.project_id
  region          = "%s"
  network_id      = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr            = "192.168.199.0/24"
  name            = "%s"
  description     = "updated by the acceptance test"
  %s
  enable_dhcp     = true
  dns_nameservers = ["8.8.8.8"]
  tags            = ["tf-acc", "updated"]

  allocation_pool {
    start = "192.168.199.30"
    end   = "192.168.199.40"
  }
}
`, region, subnetName, gateway)
}

func testAccVPCNetworkSubnetV2IPv6(projectID, region string) string {
	return testAccVPCNetworkSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  network_id        = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr              = "fd00:c0de:199::/64"
  ip_version        = 6
  ipv6_address_mode = "dhcpv6-stateful"
  ipv6_ra_mode      = "dhcpv6-stateful"
  no_gateway        = true
}
`, region)
}
