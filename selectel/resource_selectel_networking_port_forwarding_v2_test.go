package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingPortForwardingV2ResourceName = "selectel_networking_port_forwarding_v2.pf_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingPortForwardingV2Schema covers REQ-049 and REQ-050: which
// attributes are required, which are updated in place, and which replace the
// rule.
func TestNetworkingPortForwardingV2Schema(t *testing.T) {
	s := resourceNetworkingPortForwardingV2().Schema

	required := []string{
		"project_id", "region", "floatingip_id", "internal_port_id",
		"internal_ip_address", "internal_port", "external_port", "protocol",
	}

	t.Run("the required set matches the requirement", func(t *testing.T) {
		got := make([]string, 0)
		for name, attr := range s {
			if attr.Required {
				got = append(got, name)
			}
		}
		assert.ElementsMatch(t, required, got)
	})

	t.Run("description is the only optional attribute", func(t *testing.T) {
		got := make([]string, 0)
		for name, attr := range s {
			if attr.Optional {
				got = append(got, name)
			}
		}
		assert.ElementsMatch(t, []string{"description"}, got)
	})

	// REQ-050: only the scope-defining attributes replace the rule.
	t.Run("only project_id, region and floatingip_id are ForceNew", func(t *testing.T) {
		got := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				got = append(got, name)
			}
		}
		assert.ElementsMatch(t, []string{"project_id", "region", "floatingip_id"}, got)
	})

	// REQ-051: the schema deliberately does no local validation of these values.
	t.Run("no local validation of protocol and ports", func(t *testing.T) {
		for _, name := range []string{"protocol", "internal_port", "external_port"} {
			assert.Nil(t, s[name].ValidateFunc, name+" must be validated by the API")
		}
	})
}

// TestNetworkingPortForwardingV2ID covers REQ-053: the composite import
// identifier is parsed locally, and a malformed one is rejected without calling
// the API.
func TestNetworkingPortForwardingV2ID(t *testing.T) {
	t.Run("splits a well-formed identifier", func(t *testing.T) {
		fipID, ruleID, err := parseNetworkingPortForwardingV2ID(
			"1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809/9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0")

		require.NoError(t, err)
		assert.Equal(t, "1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809", fipID)
		assert.Equal(t, "9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0", ruleID)
	})

	for _, c := range []struct{ name, id string }{
		{name: "no separator", id: "1a2b3c4d-5e6f"},
		{name: "empty floating ip", id: "/9f8e7d6c"},
		{name: "empty rule", id: "1a2b3c4d/"},
		{name: "three parts", id: "1a2b/3c4d/5e6f"},
		{name: "empty identifier", id: ""},
	} {
		t.Run("rejects "+c.name, func(t *testing.T) {
			_, _, err := parseNetworkingPortForwardingV2ID(c.id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingPortForwardingV2Basic covers SCN-049-01, SCN-049-02 (empty
// follow-up plan) and SCN-053-01 (import by the composite identifier).
func TestAccNetworkingPortForwardingV2Basic(t *testing.T) {
	var found floatingip.PortForwarding
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortForwardingV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortForwardingV2ResourceName, "external_port", "8080"),
					resource.TestCheckResourceAttr(testAccNetworkingPortForwardingV2ResourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttrSet(testAccNetworkingPortForwardingV2ResourceName, "internal_ip_address"),
				),
			},
			{
				Config:   testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingPortForwardingV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNetworkingPortForwardingV2ImportID,
			},
		},
	})
}

// TestAccNetworkingPortForwardingV2Update covers SCN-050-01: a mutable attribute
// changes in place, keeping the identifier and the floating IP.
func TestAccNetworkingPortForwardingV2Update(t *testing.T) {
	var found floatingip.PortForwarding
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				Check:  testAccCheckNetworkingPortForwardingV2Exists(&found),
			},
			{
				Config: testAccNetworkingPortForwardingV2Described(projectID, region, pool, 9090, "udp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortForwardingV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortForwardingV2ResourceName, "external_port", "9090"),
					resource.TestCheckResourceAttr(testAccNetworkingPortForwardingV2ResourceName, "protocol", "udp"),
					resource.TestCheckResourceAttr(testAccNetworkingPortForwardingV2ResourceName, "description", "updated by the acceptance test"),
				),
			},
			{
				Config:   testAccNetworkingPortForwardingV2Described(projectID, region, pool, 9090, "udp"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortForwardingV2Conflict covers SCN-051-01: a second rule with
// the same external port and protocol on the same floating IP is refused by the
// API.
func TestAccNetworkingPortForwardingV2Conflict(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingPortForwardingV2Duplicate(projectID, region, pool),
				ExpectError: regexp.MustCompile(`error creating networking port forwarding`),
			},
		},
	})
}

// TestAccNetworkingPortForwardingV2InvalidValue covers SCN-051-02: the API, not
// the schema, rejects a malformed protocol.
func TestAccNetworkingPortForwardingV2InvalidValue(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "not-a-protocol"),
				ExpectError: regexp.MustCompile(`error creating networking port forwarding`),
			},
		},
	})
}

// TestAccNetworkingPortForwardingV2ExternalRemoval covers SCN-052-01 and
// SCN-054-01: a rule removed outside Terraform leaves the state quietly and the
// next plan re-creates it.
func TestAccNetworkingPortForwardingV2ExternalRemoval(t *testing.T) {
	var found floatingip.PortForwarding
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				Check:  testAccCheckNetworkingPortForwardingV2Exists(&found),
			},
			{
				Config:             testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				Check:              testAccCheckNetworkingPortForwardingV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp"),
				Check:  testAccCheckNetworkingPortForwardingV2Exists(&found),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccNetworkingPortForwardingV2Handle(s *terraform.State) (*networkingNetworkV2TestHandle, string, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingPortForwardingV2ResourceName]
	if !ok {
		return nil, "", fmt.Errorf("not found: %s", testAccNetworkingPortForwardingV2ResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, "", errors.New("no ID is set")
	}

	client, err := networkingV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, "", err
	}

	return &networkingNetworkV2TestHandle{client: client, id: rs.Primary.ID},
		rs.Primary.Attributes["floatingip_id"], nil
}

func testAccNetworkingPortForwardingV2ImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingPortForwardingV2ResourceName]
	if !ok {
		return "", fmt.Errorf("not found: %s", testAccNetworkingPortForwardingV2ResourceName)
	}

	return rs.Primary.Attributes["floatingip_id"] + "/" + rs.Primary.ID, nil
}

func testAccCheckNetworkingPortForwardingV2Exists(found *floatingip.PortForwarding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, floatingIPID, err := testAccNetworkingPortForwardingV2Handle(s)
		if err != nil {
			return err
		}

		result, err := floatingip.GetPortForwarding(context.Background(), handle.client, floatingIPID, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingPortForwardingV2SameID(found *floatingip.PortForwarding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingPortForwardingV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingPortForwardingV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("port forwarding was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingPortForwardingV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, floatingIPID, err := testAccNetworkingPortForwardingV2Handle(s)
		if err != nil {
			return err
		}

		return floatingip.DeletePortForwarding(context.Background(), handle.client, floatingIPID, handle.id)
	}
}

func testAccCheckNetworkingPortForwardingV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_port_forwarding_v2" {
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

		_, err = floatingip.GetPortForwarding(
			context.Background(), client, rs.Primary.Attributes["floatingip_id"], rs.Primary.ID)
		if err == nil {
			return errors.New("networking port forwarding still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingPortForwardingV2Base gives the rule a floating IP and an
// internal port to forward to.
func testAccNetworkingPortForwardingV2Base(projectID, region, pool string) string {
	return testAccNetworkingPortV2FixedIP(projectID, region, "192.168.199.60") +
		testAccNetworkingFloatingIPV2Gateway(region, pool) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  pool       = "%s"
}
`, region, pool)
}

func testAccNetworkingPortForwardingV2Basic(projectID, region, pool string, externalPort int, protocol string) string {
	return testAccNetworkingPortForwardingV2Base(projectID, region, pool) + fmt.Sprintf(`
resource "selectel_networking_port_forwarding_v2" "pf_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_networking_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
  internal_ip_address = "192.168.199.60"
  internal_port       = 80
  external_port       = %d
  protocol            = "%s"

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_networking_router_interface_v2.interface_tf_acc_test_1]
}
`, region, externalPort, protocol)
}

func testAccNetworkingPortForwardingV2Described(projectID, region, pool string, externalPort int, protocol string) string {
	return testAccNetworkingPortForwardingV2Base(projectID, region, pool) + fmt.Sprintf(`
resource "selectel_networking_port_forwarding_v2" "pf_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_networking_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
  internal_ip_address = "192.168.199.60"
  internal_port       = 80
  external_port       = %d
  protocol            = "%s"
  description         = "updated by the acceptance test"

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_networking_router_interface_v2.interface_tf_acc_test_1]
}
`, region, externalPort, protocol)
}

func testAccNetworkingPortForwardingV2Duplicate(projectID, region, pool string) string {
	return testAccNetworkingPortForwardingV2Basic(projectID, region, pool, 8080, "tcp") + fmt.Sprintf(`
resource "selectel_networking_port_forwarding_v2" "pf_tf_acc_test_duplicate" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_networking_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_networking_port_v2.port_tf_acc_test_1.id
  internal_ip_address = "192.168.199.60"
  internal_port       = 81
  external_port       = 8080
  protocol            = "tcp"

  depends_on = [selectel_networking_port_forwarding_v2.pf_tf_acc_test_1]
}
`, region)
}
