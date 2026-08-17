package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCSecGroupRuleV2ResourceName = "selectel_vpc_secgroup_rule_v2.rule_tf_acc_test_1"

func TestVPCSecGroupRuleV2PrefixDiffSuppression(t *testing.T) {
	cases := []struct {
		name       string
		stored     string
		configured string
		suppressed bool
	}{
		{
			name:       "an ipv6 prefix in upper case is the same network",
			stored:     "2001:558:fc00::/39",
			configured: "2001:558:FC00::/39",
			suppressed: true,
		},
		{
			name:       "identical values",
			stored:     "192.168.199.0/24",
			configured: "192.168.199.0/24",
			suppressed: true,
		},
		{
			name:       "a non-canonical prefix is the same network",
			stored:     "192.168.199.0/24",
			configured: "192.168.199.7/24",
			suppressed: true,
		},
		{
			name:       "different networks are a change",
			stored:     "192.168.199.0/24",
			configured: "192.168.200.0/24",
			suppressed: false,
		},
		{
			name:       "a cleared prefix is a change",
			stored:     "192.168.199.0/24",
			configured: "",
			suppressed: false,
		},
		{
			name:       "a newly set prefix is a change",
			stored:     "",
			configured: "192.168.199.0/24",
			suppressed: false,
		},
		{
			name:       "unparseable values fall back to a case-insensitive comparison",
			stored:     "not-a-cidr",
			configured: "NOT-A-CIDR",
			suppressed: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.suppressed,
				suppressVPCSecGroupRuleV2PrefixDiff(
					"remote_ip_prefix", testCase.stored, testCase.configured, nil))
		})
	}
}

func TestVPCSecGroupRuleV2ExpandPortRange(t *testing.T) {
	schemaMap := resourceVPCSecGroupRuleV2().Schema

	t.Run("an icmp type and code of zero reach the request", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"protocol":          "icmp",
			"port_range_min":    0,
			"port_range_max":    0,
		})

		request := expandVPCSecGroupRuleV2Request(d)

		require.NotNil(t, request.PortRangeMin, "a configured ICMP type of 0 must be sent")
		require.NotNil(t, request.PortRangeMax, "a configured ICMP code of 0 must be sent")
		assert.Equal(t, 0, *request.PortRangeMin)
		assert.Equal(t, 0, *request.PortRangeMax)
	})

	t.Run("a lone icmp type without a code reaches the request", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"protocol":          "icmp",
			"port_range_min":    8,
		})

		request := expandVPCSecGroupRuleV2Request(d)

		require.NotNil(t, request.PortRangeMin)
		assert.Equal(t, 8, *request.PortRangeMin)
		assert.Nil(t, request.PortRangeMax, "an unset ICMP code must stay unset")
	})

	t.Run("a tcp port range is carried through", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"protocol":          "tcp",
			"port_range_min":    22,
			"port_range_max":    22,
		})

		request := expandVPCSecGroupRuleV2Request(d)

		require.NotNil(t, request.PortRangeMin)
		assert.Equal(t, 22, *request.PortRangeMin)
		assert.Equal(t, "tcp", *request.Protocol)
	})

	t.Run("the remote selectors are mutually exclusive in the request", func(t *testing.T) {
		prefix := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"remote_ip_prefix":  "192.168.199.0/24",
		})
		group := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"remote_group_id":   "other-sg-id",
		})

		prefixRequest := expandVPCSecGroupRuleV2Request(prefix)
		assert.NotNil(t, prefixRequest.RemoteIPPrefix)
		assert.Nil(t, prefixRequest.RemoteGroupID)

		groupRequest := expandVPCSecGroupRuleV2Request(group)
		assert.NotNil(t, groupRequest.RemoteGroupID)
		assert.Nil(t, groupRequest.RemoteIPPrefix)

		empty := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "egress",
			"ethertype":         "IPv4",
		})
		emptyRequest := expandVPCSecGroupRuleV2Request(empty)
		assert.Nil(t, emptyRequest.RemoteIPPrefix)
		assert.Nil(t, emptyRequest.RemoteGroupID)
	})
}

func TestAccVPCSecGroupRuleV2Basic(t *testing.T) {
	var found securitygrouprule.Rule
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupRuleV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCSecGroupRuleV2ResourceName, "direction", "ingress"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupRuleV2ResourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupRuleV2ResourceName, "port_range_min", "22"),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupRuleV2ResourceName, "security_group_id",
						testAccVPCSecGroupV2ResourceName, "id",
					),
				),
			},
			{
				Config:   testAccVPCSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCSecGroupRuleV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecGroupRuleV2RemoteGroup(t *testing.T) {
	var found securitygrouprule.Rule
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupRuleV2RemoteGroup(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupRuleV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCSecGroupRuleV2ResourceName, "direction", "egress"),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupRuleV2ResourceName, "remote_group_id",
						"selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2", "id",
					),
					resource.TestCheckResourceAttr(testAccVPCSecGroupRuleV2ResourceName, "remote_ip_prefix", ""),
				),
			},
			{
				Config:   testAccVPCSecGroupRuleV2RemoteGroup(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVPCSecGroupRuleV2Exists(found *securitygrouprule.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCSecGroupRuleV2ResourceName)
		if err != nil {
			return err
		}

		result, err := securitygrouprule.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckVPCSecGroupRuleV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_secgroup_rule_v2" {
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

		if _, err := securitygrouprule.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC security group rule still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCSecGroupRuleV2IngressPrefix(projectID, region, prefix string) string {
	return testAccVPCSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "%s"
}
`, region, prefix)
}

func testAccVPCSecGroupRuleV2RemoteGroup(projectID, region string) string {
	return testAccVPCSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "secgroup-2"
}

resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%[1]s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "egress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2.id
}
`, region)
}
