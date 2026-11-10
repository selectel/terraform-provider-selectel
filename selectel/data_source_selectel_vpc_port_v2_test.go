package selectel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCPortV2DataSourceName = "data.selectel_vpc_port_v2.port_tf_acc_test_1"

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
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "port_id",
						testAccVPCPortV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "name", portName),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "network_id",
						testAccVPCPortV2ResourceName, "network_id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "mac_address",
						testAccVPCPortV2ResourceName, "mac_address",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "status",
						testAccVPCPortV2ResourceName, "status",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "device_owner",
						testAccVPCPortV2ResourceName, "device_owner",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "dns_name",
						testAccVPCPortV2ResourceName, "dns_name",
					),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "all_fixed_ips.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "all_fixed_ips.0.ip_address", testAccVPCPortV2FixedIPAddress),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "all_fixed_ips.0.subnet_id",
						testAccVPCNetworkSubnetV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "all_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "allowed_address_pairs.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(testAccVPCPortV2DataSourceName, "allowed_address_pairs.*", map[string]string{
						"ip_address": "192.168.199.100",
					}),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "extra_dhcp_option.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(testAccVPCPortV2DataSourceName, "extra_dhcp_option.*", map[string]string{
						"name":       "tftp-server",
						"value":      "192.168.199.2",
						"ip_version": "4",
					}),
					resource.TestCheckResourceAttr(testAccVPCPortV2DataSourceName, "all_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(testAccVPCPortV2DataSourceName, "all_tags.*", portName),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "port_security_enabled",
						testAccVPCPortV2ResourceName, "port_security_enabled",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "dhcp_blocked",
						testAccVPCPortV2ResourceName, "dhcp_blocked",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortV2DataSourceName, "blocked",
						testAccVPCPortV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccVPCPortV2DataSourceAmbiguous(projectID, region, portName),
				ExpectError: regexp.MustCompile(`2 VPC ports found`),
			},
			{
				Config:      testAccVPCPortV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC port found`),
			},
		},
	})
}

func testAccVPCPortV2DataSource(projectID, region, portName string) string {
	return testAccVPCPortV2SecGroups(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%[1]s"
  network_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name               = "%[2]s"
  description        = "created by the acceptance test"
  admin_state_up     = true
  tags               = ["tf-acc", "%[2]s"]
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id]

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "%[3]s"
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  extra_dhcp_option {
    name  = "tftp-server"
    value = "192.168.199.2"
  }
}
data "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%[1]s"
  tags               = ["%[2]s"]
  admin_state_up     = true
  network_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  mac_address        = selectel_vpc_port_v2.port_tf_acc_test_1.mac_address
  status             = selectel_vpc_port_v2.port_tf_acc_test_1.status
  description        = selectel_vpc_port_v2.port_tf_acc_test_1.description
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id]
}
`, region, portName, testAccVPCPortV2FixedIPAddress)
}

func testAccVPCPortV2DataSourceAmbiguous(projectID, region, portName string) string {
	return testAccVPCPortV2SecGroups(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%[1]s"
  network_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name               = "%[2]s"
  tags               = ["%[2]s"]
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id]

  depends_on = [selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1]
}
resource "selectel_vpc_port_v2" "port_tf_acc_test_2" {
  project_id         = local.project_id
  region             = "%[1]s"
  network_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  name               = "%[2]s-twin"
  tags               = ["%[2]s"]
  no_fixed_ip        = true
  no_security_groups = true

  depends_on = [selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1]
}

data "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  tags       = ["%[2]s"]

  depends_on = [
    selectel_vpc_port_v2.port_tf_acc_test_1,
    selectel_vpc_port_v2.port_tf_acc_test_2,
  ]
}
`, region, portName)
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
