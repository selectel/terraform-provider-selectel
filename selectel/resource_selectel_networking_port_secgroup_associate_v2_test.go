package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingPortSecGroupAssociateV2ResourceName = "selectel_networking_port_secgroup_associate_v2.associate_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingPortSecGroupAssociateV2Schema covers REQ-119: the resource
// addresses a port it does not own, and both the set and the mode are updatable.
func TestNetworkingPortSecGroupAssociateV2Schema(t *testing.T) {
	s := resourceNetworkingPortSecGroupAssociateV2().Schema

	assert.Len(t, s, 6, "the association is fully described by these six attributes")

	t.Run("project_id, region and port_id are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region", "port_id"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	t.Run("security_group_ids is required and updatable", func(t *testing.T) {
		require.Contains(t, s, "security_group_ids")
		assert.True(t, s["security_group_ids"].Required)
		assert.False(t, s["security_group_ids"].ForceNew)
	})

	t.Run("enforce is optional, defaults to false and is updatable", func(t *testing.T) {
		require.Contains(t, s, "enforce")
		assert.True(t, s["enforce"].Optional)
		assert.Equal(t, false, s["enforce"].Default)
		assert.False(t, s["enforce"].ForceNew)
	})

	t.Run("all_security_group_ids is computed-only", func(t *testing.T) {
		require.Contains(t, s, "all_security_group_ids")
		assert.True(t, s["all_security_group_ids"].Computed)
		assert.False(t, s["all_security_group_ids"].Optional)
	})

	// The resource owns the association, not the port: nothing else about the
	// port appears in its schema.
	t.Run("no other port attribute is exposed", func(t *testing.T) {
		for _, name := range []string{"network_id", "fixed_ip", "mac_address", "name"} {
			assert.NotContains(t, s, name)
		}
	})
}

// TestNetworkingPortSecGroupAssociateV2SetOperations covers the set algebra the
// two modes are built on: REQ-120 (additive create), REQ-122 (additive update
// replaces only the previously claimed subset) and REQ-124 (additive delete
// releases only what it claimed).
func TestNetworkingPortSecGroupAssociateV2SetOperations(t *testing.T) {
	t.Run("union keeps everything once, in a stable order", func(t *testing.T) {
		// REQ-120: an additive create adds to what the port already has.
		assert.Equal(t, []string{"a", "b", "c"},
			unionNetworkingSecGroupIDs([]string{"b", "a"}, []string{"c", "a"}))
		assert.Equal(t, []string{"a"}, unionNetworkingSecGroupIDs([]string{"a"}, nil))
		assert.Equal(t, []string{"a"}, unionNetworkingSecGroupIDs(nil, []string{"a"}))
		assert.Empty(t, unionNetworkingSecGroupIDs(nil, nil))
	})

	t.Run("intersect reports only what is on both sides", func(t *testing.T) {
		// REQ-123: in additive mode Read narrows the managed set to what the port
		// actually still carries.
		assert.Equal(t, []string{"a"},
			intersectNetworkingSecGroupIDs([]string{"a", "z"}, []string{"a", "b"}))
		assert.Empty(t, intersectNetworkingSecGroupIDs([]string{"z"}, []string{"a"}))
		assert.Empty(t, intersectNetworkingSecGroupIDs(nil, []string{"a"}))
	})

	t.Run("subtract removes only the named identifiers", func(t *testing.T) {
		// REQ-124: an additive delete releases its own subset and nothing else.
		assert.Equal(t, []string{"b"},
			subtractNetworkingSecGroupIDs([]string{"a", "b"}, []string{"a"}))
		assert.Equal(t, []string{"a", "b"},
			subtractNetworkingSecGroupIDs([]string{"b", "a"}, []string{"z"}))
		assert.Empty(t, subtractNetworkingSecGroupIDs([]string{"a"}, []string{"a"}))
	})

	// SCN-122-01 end to end: managed {A}, port {A, B} externally, new config {C}
	// must leave the port as {B, C}.
	t.Run("an additive update leaves foreign groups alone", func(t *testing.T) {
		onPort := []string{"A", "B"}
		previouslyManaged := []string{"A"}
		desired := []string{"C"}

		kept := subtractNetworkingSecGroupIDs(onPort, previouslyManaged)
		result := unionNetworkingSecGroupIDs(kept, desired)

		assert.Equal(t, []string{"B", "C"}, result)
	})

	// SCN-124-01 end to end: managed {A}, port {A, B} — the delete leaves {B}.
	t.Run("an additive delete leaves foreign groups alone", func(t *testing.T) {
		assert.Equal(t, []string{"B"},
			subtractNetworkingSecGroupIDs([]string{"A", "B"}, []string{"A"}))
	})
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingPortSecGroupAssociateV2Additive covers SCN-120-01, SCN-120-02,
// SCN-119-01 and SCN-124-01.
func TestAccNetworkingPortSecGroupAssociateV2Additive(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					// SCN-120-01: the group the port already had is still there,
					// alongside the configured one.
					resource.TestCheckResourceAttr(testAccNetworkingPortSecGroupAssociateV2ResourceName, "security_group_ids.#", "1"),
					testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast(2),
					testAccCheckNetworkingPortSecGroupAssociateV2IDIsPort(),
				),
			},
			{
				// SCN-120-02: the follow-up plan is empty.
				Config:   testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				PlanOnly: true,
			},
			{
				// SCN-119-01: import by port_id.
				ResourceName:      testAccNetworkingPortSecGroupAssociateV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// On import there is no configuration yet, so the managed subset is
				// empty and enforce takes its default.
				ImportStateVerifyIgnore: []string{"security_group_ids", "enforce"},
			},
		},
	})
}

// TestAccNetworkingPortSecGroupAssociateV2Enforce covers SCN-121-01: the
// configured set becomes the port's whole list.
func TestAccNetworkingPortSecGroupAssociateV2Enforce(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortSecGroupAssociateV2Enforce(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					// Exactly the configured group, so the one the port carried
					// before has been removed.
					resource.TestCheckResourceAttr(testAccNetworkingPortSecGroupAssociateV2ResourceName, "security_group_ids.#", "1"),
					testAccCheckNetworkingPortSecGroupAssociateV2PortHas(1),
				),
			},
			{
				Config:   testAccNetworkingPortSecGroupAssociateV2Enforce(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortSecGroupAssociateV2SwitchToEnforce covers SCN-122-03: the
// switch itself forces the port to match the configuration.
func TestAccNetworkingPortSecGroupAssociateV2SwitchToEnforce(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				// Additive: the port keeps its pre-existing group as well.
				Config: testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				Check:  testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast(2),
			},
			{
				// enforce=true with the same set: the port drops everything else,
				// and the resource is updated rather than replaced.
				Config: testAccNetworkingPortSecGroupAssociateV2Enforce(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortSecGroupAssociateV2IDIsPort(),
					testAccCheckNetworkingPortSecGroupAssociateV2PortHas(1),
				),
			},
		},
	})
}

// TestAccNetworkingPortSecGroupAssociateV2ExternalPortDeletion covers REQ-123 and
// REQ-124: the association cannot outlive its port.
func TestAccNetworkingPortSecGroupAssociateV2ExternalPortDeletion(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				Check:  testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast(2),
			},
			{
				Config:             testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				Check:              testAccCheckNetworkingPortSecGroupAssociateV2DeletePortOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				// The plan re-creates the port and the association rather than
				// reporting an error.
				Config: testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region),
				Check:  testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast(2),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingPortSecGroupAssociateV2IDIsPort() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingPortSecGroupAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingPortSecGroupAssociateV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["port_id"] {
			return fmt.Errorf("resource id %s is not the port id %s",
				rs.Primary.ID, rs.Primary.Attributes["port_id"])
		}

		return nil
	}
}

func testAccCheckNetworkingPortSecGroupAssociateV2PortHas(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groups, err := testAccNetworkingPortSecGroupAssociateV2Groups(s)
		if err != nil {
			return err
		}

		if len(groups) != expected {
			return fmt.Errorf("the port has %d security groups, expected %d", len(groups), expected)
		}

		return nil
	}
}

// testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast is used in additive
// mode, where the count also depends on what the platform put on the port.
func testAccCheckNetworkingPortSecGroupAssociateV2PortHasAtLeast(minimum int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groups, err := testAccNetworkingPortSecGroupAssociateV2Groups(s)
		if err != nil {
			return err
		}

		if len(groups) < minimum {
			return fmt.Errorf("the port has %d security groups, expected at least %d",
				len(groups), minimum)
		}

		return nil
	}
}

func testAccCheckNetworkingPortSecGroupAssociateV2DeletePortOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingPortSecGroupAssociateV2ResourceName)
		if err != nil {
			return err
		}

		return port.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccNetworkingPortSecGroupAssociateV2Groups(s *terraform.State) ([]string, error) {
	handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingPortSecGroupAssociateV2ResourceName)
	if err != nil {
		return nil, err
	}

	found, err := port.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.SecurityGroups, nil
}

// testAccCheckNetworkingPortSecGroupAssociateV2Destroy asserts the association
// released what it owned. The port itself is not this resource's to delete, so a
// port that survives with other groups is the expected outcome.
func testAccCheckNetworkingPortSecGroupAssociateV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_port_secgroup_associate_v2" {
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

		found, err := port.Get(context.Background(), client, rs.Primary.ID)
		if err != nil {
			if networkingIsNotFound(err) {
				continue
			}

			return err
		}

		// Whatever mode it was in, the group this resource managed must be gone.
		managed := rs.Primary.Attributes["security_group_ids.0"]
		if managed == "" {
			continue
		}
		for _, group := range found.SecurityGroups {
			if group == managed {
				return fmt.Errorf("security group %s is still associated with port %s",
					managed, rs.Primary.ID)
			}
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingPortSecGroupAssociateV2Base is a port that already carries one
// security group, plus a second group for the association to add or enforce.
func testAccNetworkingPortSecGroupAssociateV2Base(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}

resource "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "secgroup-existing"
}

resource "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "secgroup-associated"
}

resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%[1]s"
  network_id         = selectel_networking_network_v2.network_tf_acc_test_1.id
  security_group_ids = [selectel_networking_secgroup_v2.secgroup_tf_acc_test_1.id]

  fixed_ip {
    subnet_id = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  }
}
`, region)
}

func testAccNetworkingPortSecGroupAssociateV2Additive(projectID, region string) string {
	return testAccNetworkingPortSecGroupAssociateV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_secgroup_associate_v2" "associate_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%s"
  port_id            = selectel_networking_port_v2.port_tf_acc_test_1.id
  security_group_ids = [selectel_networking_secgroup_v2.secgroup_tf_acc_test_2.id]
}
`, region)
}

func testAccNetworkingPortSecGroupAssociateV2Enforce(projectID, region string) string {
	return testAccNetworkingPortSecGroupAssociateV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_secgroup_associate_v2" "associate_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%s"
  port_id            = selectel_networking_port_v2.port_tf_acc_test_1.id
  security_group_ids = [selectel_networking_secgroup_v2.secgroup_tf_acc_test_2.id]
  enforce            = true
}
`, region)
}
