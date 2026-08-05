package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingSecGroupRuleV2ResourceName = "selectel_vpc_secgroup_rule_v2.rule_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingSecGroupRuleV2Schema covers REQ-111 and REQ-115: the rule is
// immutable, so every attribute replaces it.
func TestNetworkingSecGroupRuleV2Schema(t *testing.T) {
	s := resourceVPCSecGroupRuleV2().Schema

	t.Run("the required attributes", func(t *testing.T) {
		for _, name := range []string{
			"project_id", "region", "security_group_id", "direction", "ethertype",
		} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required, name+" must be required")
		}
	})

	t.Run("the optional attributes", func(t *testing.T) {
		for _, name := range []string{
			"description", "protocol", "port_range_min", "port_range_max",
			"remote_group_id", "remote_ip_prefix",
		} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Optional, name+" must be optional")
		}
	})

	// REQ-115: there is no update path, so nothing may be updated in place.
	t.Run("every attribute is ForceNew and there is no update path", func(t *testing.T) {
		for name, attr := range s {
			assert.True(t, attr.ForceNew, name+" must replace the rule when it changes")
		}
		assert.Nil(t, resourceVPCSecGroupRuleV2().UpdateContext)
	})

	// REQ-112: the two remote selectors are mutually exclusive at plan time.
	t.Run("the remote selectors conflict with each other", func(t *testing.T) {
		assert.Equal(t, []string{"remote_ip_prefix"}, s["remote_group_id"].ConflictsWith)
		assert.Equal(t, []string{"remote_group_id"}, s["remote_ip_prefix"].ConflictsWith)
	})

	// REQ-112: the local checks mirror the asymmetry the fork enforces — either
	// bound needs a protocol, and a code needs a type, but a lone type is fine.
	t.Run("a port range requires protocol, and a code requires a type", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"protocol"}, s["port_range_min"].RequiredWith,
			"a lone port_range_min is a valid ICMP type and must not be rejected locally")
		assert.ElementsMatch(t, []string{"protocol", "port_range_min"}, s["port_range_max"].RequiredWith)
	})

	// remote_address_group_id is out of scope: address groups cannot be created
	// by a project member in this cloud, so there is nothing to reference.
	t.Run("remote_address_group_id is not exposed", func(t *testing.T) {
		assert.NotContains(t, s, "remote_address_group_id")
	})
}

// TestNetworkingSecGroupRuleV2PrefixDiffSuppression covers SCN-115-01: a prefix
// written differently from the form Neutron stores is not a change.
func TestNetworkingSecGroupRuleV2PrefixDiffSuppression(t *testing.T) {
	cases := []struct {
		name       string
		stored     string
		configured string
		suppressed bool
	}{
		{
			// The scenario's own example: Neutron lowercases what it stores.
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
			// Neutron canonicalises a host address to its network address.
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
				suppressNetworkingSecGroupRuleV2PrefixDiff(
					"remote_ip_prefix", testCase.stored, testCase.configured, nil))
		})
	}
}

// TestNetworkingSecGroupRuleV2ExpandPortRange covers the port-range half of
// REQ-111: with an ICMP protocol, port_range_min and port_range_max carry the
// ICMP type and code, and Neutron accepts 0..255 for both
// (neutron/db/securitygroups_db.py, _validate_port_range). A configured zero
// must therefore reach the API — dropping it would create a rule matching every
// ICMP type instead of type 0, which is wider than the configuration asked for.
func TestNetworkingSecGroupRuleV2ExpandPortRange(t *testing.T) {
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

		request := expandNetworkingSecGroupRuleV2Request(d)

		require.NotNil(t, request.PortRangeMin, "a configured ICMP type of 0 must be sent")
		require.NotNil(t, request.PortRangeMax, "a configured ICMP code of 0 must be sent")
		assert.Equal(t, 0, *request.PortRangeMin)
		assert.Equal(t, 0, *request.PortRangeMax)
	})

	// SCN-112-04: a lone port_range_min is a valid ICMP type with any code, so it
	// must survive local validation and reach the request.
	t.Run("a lone icmp type without a code reaches the request", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "ingress",
			"ethertype":         "IPv4",
			"protocol":          "icmp",
			"port_range_min":    8,
		})

		request := expandNetworkingSecGroupRuleV2Request(d)

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

		request := expandNetworkingSecGroupRuleV2Request(d)

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

		assert.NotNil(t, expandNetworkingSecGroupRuleV2Request(prefix).RemoteSource)
		assert.NotNil(t, expandNetworkingSecGroupRuleV2Request(group).RemoteSource)

		// Neither set means no remote restriction at all.
		empty := schema.TestResourceDataRaw(t, schemaMap, map[string]interface{}{
			"security_group_id": "sg-id",
			"direction":         "egress",
			"ethertype":         "IPv4",
		})
		assert.Nil(t, expandNetworkingSecGroupRuleV2Request(empty).RemoteSource)
	})
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingSecGroupRuleV2Basic covers SCN-111-01, SCN-115-01, SCN-117-01
// and SCN-118-01.
func TestAccNetworkingSecGroupRuleV2Basic(t *testing.T) {
	var found securitygroup.SecurityGroupRule
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupRuleV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "direction", "ingress"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "port_range_min", "22"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSecGroupRuleV2ResourceName, "security_group_id",
						testAccNetworkingSecGroupV2ResourceName, "id",
					),
				),
			},
			{
				Config:   testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				PlanOnly: true,
			},
			{
				// SCN-117-01: import by the rule identifier alone.
				ResourceName:      testAccNetworkingSecGroupRuleV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2UpperCasePrefix covers SCN-115-01's own example:
// an IPv6 prefix written in upper case must not produce a permanent diff.
func TestAccNetworkingSecGroupRuleV2UpperCasePrefix(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupRuleV2IPv6Prefix(projectID, region, "2001:558:FC00::/39"),
			},
			{
				// The API stored it lowercased; the plan must still be empty.
				Config:   testAccNetworkingSecGroupRuleV2IPv6Prefix(projectID, region, "2001:558:FC00::/39"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2RemoteGroup covers SCN-111-02: a rule that
// references another security group as its source.
func TestAccNetworkingSecGroupRuleV2RemoteGroup(t *testing.T) {
	var found securitygroup.SecurityGroupRule
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupRuleV2RemoteGroup(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupRuleV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "direction", "egress"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSecGroupRuleV2ResourceName, "remote_group_id",
						"selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2", "id",
					),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "remote_ip_prefix", ""),
				),
			},
			{
				Config:   testAccNetworkingSecGroupRuleV2RemoteGroup(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2LocalValidation covers SCN-112-02 and SCN-112-03:
// the schema rejects these before any API call.
func TestAccNetworkingSecGroupRuleV2LocalValidation(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				// Both remote selectors at once.
				Config:      testAccNetworkingSecGroupRuleV2BothRemotes(projectID, region),
				ExpectError: regexp.MustCompile(`(?s)"remote_group_id".*conflicts with.*"remote_ip_prefix"|"remote_ip_prefix".*conflicts with.*"remote_group_id"`),
			},
			{
				// A port range without a protocol.
				Config:      testAccNetworkingSecGroupRuleV2PortRangeWithoutProtocol(projectID, region),
				ExpectError: regexp.MustCompile(`(?s)"port_range_min".*all of .*protocol|protocol.*must be specified`),
			},
			{
				// SCN-112-03: port_range_max without port_range_min. The reverse is
				// deliberately allowed — see TestAccNetworkingSecGroupRuleV2ICMPType.
				Config:      testAccNetworkingSecGroupRuleV2MaxWithoutMin(projectID, region),
				ExpectError: regexp.MustCompile(`(?s)port_range_max.*all of|must be specified`),
			},
			{
				// A rejected configuration is never applied, so the last step has to
				// leave a valid one behind: the framework destroys the configuration
				// of the final step, and a configuration that fails validation cannot
				// be destroyed.
				Config: testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.0.0/16"),
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2ICMPType covers SCN-112-04: a single
// port_range_min is an ICMP type without a code, so local validation must accept
// it and the rule must be created. This is the asymmetry SCN-112-03 rests on — the
// same lone bound with a TCP protocol is rejected by the API, not by the schema.
func TestAccNetworkingSecGroupRuleV2ICMPType(t *testing.T) {
	var found securitygroup.SecurityGroupRule
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupRuleV2ICMPType(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupRuleV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "protocol", "icmp"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupRuleV2ResourceName, "port_range_min", "8"),
				),
			},
			{
				Config:   testAccNetworkingSecGroupRuleV2ICMPType(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2APIValidation covers SCN-113-01 and SCN-113-02:
// values the schema passes through and the API rejects.
func TestAccNetworkingSecGroupRuleV2APIValidation(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-113-01: port_range_min greater than port_range_max.
				Config:      testAccNetworkingSecGroupRuleV2PortRange(projectID, region, 100, 50),
				ExpectError: regexp.MustCompile(`(?i)error creating|invalid|bad request`),
			},
			{
				// SCN-113-02: an IPv6 prefix with ethertype IPv4.
				Config:      testAccNetworkingSecGroupRuleV2EtherTypeMismatch(projectID, region),
				ExpectError: regexp.MustCompile(`(?i)error creating|invalid|ethertype|conflict`),
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2Duplicate covers SCN-114-01: an identical rule
// on the same group is refused by the API.
func TestAccNetworkingSecGroupRuleV2Duplicate(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingSecGroupRuleV2Duplicate(projectID, region),
				ExpectError: regexp.MustCompile(`(?i)already exists|conflict|duplicate`),
			},
		},
	})
}

// TestAccNetworkingSecGroupRuleV2ExternalDeletion covers SCN-116-01: the rule
// removed outside Terraform leaves the state without an error.
func TestAccNetworkingSecGroupRuleV2ExternalDeletion(t *testing.T) {
	var found securitygroup.SecurityGroupRule
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupRuleV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				Check:  testAccCheckNetworkingSecGroupRuleV2Exists(&found),
			},
			{
				Config:             testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				Check:              testAccCheckNetworkingSecGroupRuleV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24"),
				Check:  testAccCheckNetworkingSecGroupRuleV2Exists(&found),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingSecGroupRuleV2Exists(found *securitygroup.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupRuleV2ResourceName)
		if err != nil {
			return err
		}

		result, err := securitygroup.GetRule(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingSecGroupRuleV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupRuleV2ResourceName)
		if err != nil {
			return err
		}

		return securitygroup.DeleteRule(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingSecGroupRuleV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_secgroup_rule_v2" {
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

		if _, err := securitygroup.GetRule(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking security group rule still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, prefix string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
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

func testAccNetworkingSecGroupRuleV2IPv6Prefix(projectID, region, prefix string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv6"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "%s"
}
`, region, prefix)
}

func testAccNetworkingSecGroupRuleV2RemoteGroup(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
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

func testAccNetworkingSecGroupRuleV2BothRemotes(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_group_id   = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  remote_ip_prefix  = "192.168.199.0/24"
}
`, region)
}

func testAccNetworkingSecGroupRuleV2PortRangeWithoutProtocol(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  port_range_min    = 22
  port_range_max    = 22
}
`, region)
}

func testAccNetworkingSecGroupRuleV2ICMPType(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  port_range_min    = 8
}
`, region)
}

func testAccNetworkingSecGroupRuleV2MaxWithoutMin(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_max    = 22
}
`, region)
}

func testAccNetworkingSecGroupRuleV2PortRange(projectID, region string, minPort, maxPort int) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = %d
  port_range_max    = %d
}
`, region, minPort, maxPort)
}

func testAccNetworkingSecGroupRuleV2EtherTypeMismatch(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_ip_prefix  = "2001:db8::/64"
}
`, region)
}

// testAccNetworkingSecGroupRuleV2Duplicate declares two rules with the same
// comparable fields on the same group.
func testAccNetworkingSecGroupRuleV2Duplicate(projectID, region string) string {
	return testAccNetworkingSecGroupRuleV2IngressPrefix(projectID, region, "192.168.199.0/24") + fmt.Sprintf(`
resource "selectel_vpc_secgroup_rule_v2" "rule_tf_acc_test_2" {
  project_id        = local.project_id
  region            = "%s"
  security_group_id = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "192.168.199.0/24"

  depends_on = [selectel_vpc_secgroup_rule_v2.rule_tf_acc_test_1]
}
`, region)
}
