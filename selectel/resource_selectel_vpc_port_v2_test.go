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
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccVPCPortV2ResourceName   = "selectel_vpc_port_v2.port_tf_acc_test_1"
	testAccVPCPortV2DataSourceName = "data.selectel_vpc_port_v2.port_tf_acc_test_1"
	testAccVPCPortV2FixedIPAddress = "192.168.199.42"
)

func TestAccVPCPortV2Basic(t *testing.T) {
	var found port.Port
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortV2Exists(&found),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "mac_address"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "status"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "all_fixed_ips.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "all_security_group_ids.#", "1"),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "port_security_enabled"),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "dhcp_blocked"),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "blocked"),
				),
			},
			{
				Config:   testAccVPCPortV2Basic(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:            testAccVPCPortV2ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"no_fixed_ip", "no_security_groups"},
			},
		},
	})
}

func TestAccVPCPortV2ExtendedAttributes(t *testing.T) {
	var found port.Port
	projectID, region := vpcV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortV2Extended(projectID, region, portName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "name", portName),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "fixed_ip.0.ip_address", testAccVPCPortV2FixedIPAddress),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "all_fixed_ips.0.ip_address", testAccVPCPortV2FixedIPAddress),
					resource.TestCheckResourceAttrSet(testAccVPCPortV2ResourceName, "all_fixed_ips.0.subnet_id"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "allowed_address_pairs.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "extra_dhcp_option.#", "2"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config:   testAccVPCPortV2Extended(projectID, region, portName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCPortV2Update(t *testing.T) {
	var found port.Port
	projectID, region := vpcV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")
	updatedName := acctest.RandomWithPrefix("tf-acc-port-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortV2Extended(projectID, region, portName),
				Check:  testAccCheckVPCPortV2Exists(&found),
			},
			{
				Config: testAccVPCPortV2ExtendedUpdated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "allowed_address_pairs.#", "2"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "extra_dhcp_option.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCPortV2ResourceName, "tags.#", "1"),
				),
			},
			{
				Config:   testAccVPCPortV2ExtendedUpdated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCPortV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortV2DataSource(projectID, region, portName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "name", portName),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "port_id",
						testAccVPCPortV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "blocked",
						testAccVPCPortV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccVPCPortV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC port found`),
			},
		},
	})
}

func testAccVPCPortV2Handle(s *terraform.State) (*vpcNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[testAccVPCPortV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccVPCPortV2ResourceName)
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

func testAccCheckVPCPortV2Exists(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCPortV2Handle(s)
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

func testAccCheckVPCPortV2SameID(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCPortV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCPortV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC port was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCPortV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_port_v2" {
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
			return errors.New("VPC port still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCPortV2Subnet(projectID, region string) string {
	return testAccVPCNetworkSubnetV2Basic(projectID, region, "192.168.199.0/24")
}

func testAccVPCPortV2Basic(projectID, region string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  depends_on = [selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1]
}
`, region)
}

func testAccVPCPortV2Named(projectID, region, portName string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name       = "%s"

  depends_on = [selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1]
}
`, region, portName)
}

func testAccVPCPortV2FixedIP(projectID, region, address string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "%s"
  }
}
`, region, address)
}

func testAccVPCPortV2Extended(projectID, region, portName string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  network_id     = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name           = "%s"
  description    = "created by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc", "VPC"]

  fixed_ip {
    ip_address = "%s"
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  extra_dhcp_option {
    name  = "tftp-server"
    value = "192.168.199.2"
  }

  extra_dhcp_option {
    name  = "bootfile-name"
    value = "pxelinux.0"
  }

  depends_on = [selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1]
}
`, region, portName, testAccVPCPortV2FixedIPAddress)
}

func testAccVPCPortV2ExtendedUpdated(projectID, region, portName string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  network_id     = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name           = "%s"
  description    = "updated by the acceptance test"
  admin_state_up = false
  tags           = ["tf-acc"]

  fixed_ip {
    subnet_id = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  allowed_address_pairs {
    ip_address = "192.168.199.101"
  }

  extra_dhcp_option {
    name  = "tftp-server"
    value = "192.168.199.3"
  }
}
`, region, portName)
}

func testAccVPCPortV2DataSource(projectID, region, portName string) string {
	return testAccVPCPortV2Named(projectID, region, portName) + fmt.Sprintf(`
data "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_vpc_port_v2.port_tf_acc_test_1.name
  fixed_ip   = selectel_vpc_port_v2.port_tf_acc_test_1.all_fixed_ips[0].ip_address
}
`, region)
}

func testAccVPCPortV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
data "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-port"
}
`, region)
}

func TestVPCPortV2Expanders(t *testing.T) {
	t.Run("fixed ips round trip", func(t *testing.T) {
		raw := []any{
			map[string]any{"subnet_id": "s-1", "ip_address": "192.168.199.10"},
		}

		fixedIPs := expandVPCPortV2FixedIPs(raw)

		require.Len(t, fixedIPs, 1)
		assert.Equal(t, port.FixedIP{SubnetID: "s-1", IPAddress: "192.168.199.10"}, fixedIPs[0])
		assert.Equal(t, raw, flattenVPCPortV2FixedIPs(fixedIPs))
	})

	t.Run("allowed address pairs round trip", func(t *testing.T) {
		raw := []any{
			map[string]any{"ip_address": "192.168.199.20", "mac_address": "fa:16:3e:00:00:01"},
		}

		pairs := expandVPCPortV2AllowedAddressPairs(raw)

		require.Len(t, pairs, 1)
		assert.Equal(t, "192.168.199.20", pairs[0].IPAddress)
		assert.Equal(t, raw, flattenVPCPortV2AllowedAddressPairs(pairs))
	})

	t.Run("extra dhcp options round trip with the ip version", func(t *testing.T) {
		raw := []any{
			map[string]any{"name": "tftp-server", "value": "10.0.0.1", "ip_version": 4},
		}

		options := expandVPCPortV2ExtraDHCPOptions(raw)

		require.Len(t, options, 1)
		assert.Equal(t, "tftp-server", options[0].Name)
		assert.Equal(t, vpcNetworkSubnetV2IPVersion4, options[0].IPVersion)
		assert.Equal(t, raw, flattenVPCPortV2ExtraDHCPOptions(options))
	})

	t.Run("extra dhcp options dropped from the configuration are sent as removals", func(t *testing.T) {
		option := func(name, value string) any {
			return map[string]any{"name": name, "value": value, "ip_version": 4}
		}
		set := func(items ...any) *schema.Set {
			nested, ok := resourceVPCPortV2().Schema["extra_dhcp_option"].Elem.(*schema.Resource)
			require.True(t, ok)

			return schema.NewSet(schema.HashResource(nested), items)
		}

		options := expandVPCPortV2ExtraDHCPOptionsUpdate(
			set(option("tftp-server", "10.0.0.1"), option("bootfile-name", "pxelinux.0")),
			set(option("tftp-server", "10.0.0.2")),
		)

		require.Len(t, options, 2)

		byName := make(map[string]port.UpdateExtraDHCPOption, len(options))
		for _, item := range options {
			byName[item.Name] = item
		}

		require.Contains(t, byName, "tftp-server")
		require.NotNil(t, byName["tftp-server"].Value)
		assert.Equal(t, "10.0.0.2", *byName["tftp-server"].Value)

		require.Contains(t, byName, "bootfile-name")
		assert.Nil(t, byName["bootfile-name"].Value,
			"an option no longer configured must be sent with a null value")
	})
}

func TestVPCPortV2ListFilters(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceVPCPortV2().Schema, map[string]any{
		"fixed_ip": "192.168.199.11",
	})

	filters := vpcPortV2ListFilters(d)

	assert.Equal(t, []string{"ip_address=192.168.199.11"}, filters["fixed_ips"])
	assert.Contains(t,
		vpc.SelectionOptions{Filters: filters}.Values().Encode(),
		"fixed_ips=ip_address%3D192.168.199.11",
	)
}

func TestVPCPortV2SecurityGroupFilter(t *testing.T) {
	candidates := []port.Port{
		{
			ID:             "p-1",
			SecurityGroups: []string{"sg-1", "sg-2"},
		},
		{
			ID:             "p-2",
			SecurityGroups: []string{"sg-3"},
		},
	}

	newData := func(raw map[string]any) *schema.ResourceData {
		return schema.TestResourceDataRaw(t, dataSourceVPCPortV2().Schema, raw)
	}

	t.Run("security_group_ids matches an intersection, not an exact set", func(t *testing.T) {
		matched := filterVPCPortV2BySecurityGroups(newData(map[string]any{
			"security_group_ids": []any{"sg-2", "sg-9"},
		}), candidates)

		require.Len(t, matched, 1)
		assert.Equal(t, "p-1", matched[0].ID)
	})

	t.Run("without a security group selector everything passes through", func(t *testing.T) {
		matched := filterVPCPortV2BySecurityGroups(newData(map[string]any{}), candidates)

		assert.Len(t, matched, 2)
	})
}
