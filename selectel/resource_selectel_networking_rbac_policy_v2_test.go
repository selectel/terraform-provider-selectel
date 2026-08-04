package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/rbacpolicy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingRBACPolicyV2ResourceName = "selectel_networking_rbac_policy_v2.rbac_tf_acc_test_1"

// networkingRBACPolicyV2Targets is where the policy grants access to. A grant
// needs real second and third tenants, so by default the configuration creates two
// empty projects itself and needs nothing from the environment.
//
// SELECTEL_RBAC_TARGET_PROJECT_ID and SELECTEL_RBAC_TARGET_PROJECT_ID_ALTERNATIVE
// point the tests at existing projects instead. That escape hatch exists because
// the Resell project API is not always reliable: a 500 on reading or deleting a
// freshly created project both fails the test and leaves an orphaned project
// behind, which pointing at two long-lived projects avoids.
type networkingRBACPolicyV2Targets struct {
	// prefix names the projects the configuration creates; empty means it creates
	// none and the targets below are literal identifiers.
	prefix string
	first  string
	second string
}

func networkingRBACPolicyV2TestTargets() networkingRBACPolicyV2Targets {
	first := os.Getenv("SELECTEL_RBAC_TARGET_PROJECT_ID")
	second := os.Getenv("SELECTEL_RBAC_TARGET_PROJECT_ID_ALTERNATIVE")
	if first != "" && second != "" {
		return networkingRBACPolicyV2Targets{
			first:  strconv.Quote(first),
			second: strconv.Quote(second),
		}
	}

	return networkingRBACPolicyV2Targets{
		prefix: acctest.RandomWithPrefix("tf-acc-rbac-target"),
		first:  testAccNetworkingRBACPolicyV2FirstTarget + ".id",
		second: testAccNetworkingRBACPolicyV2SecondTarget + ".id",
	}
}

// check asserts that target_tenant points at the given target, whether that is a
// project the configuration created or an identifier taken from the environment.
func (targets networkingRBACPolicyV2Targets) check(target string) resource.TestCheckFunc {
	if targets.prefix != "" {
		return resource.TestCheckResourceAttrPair(
			testAccNetworkingRBACPolicyV2ResourceName, "target_tenant",
			strings.TrimSuffix(target, ".id"), "id",
		)
	}

	value, err := strconv.Unquote(target)
	if err != nil {
		value = target
	}

	return resource.TestCheckResourceAttr(
		testAccNetworkingRBACPolicyV2ResourceName, "target_tenant", value)
}

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingRBACPolicyV2Schema covers REQ-125 and REQ-126: what defines the
// policy replaces it, and target_tenant is the one attribute updated in place.
func TestNetworkingRBACPolicyV2Schema(t *testing.T) {
	s := resourceNetworkingRBACPolicyV2().Schema

	assert.Len(t, s, 6, "the policy is fully described by its six arguments")

	t.Run("every attribute is required", func(t *testing.T) {
		for name, attr := range s {
			assert.True(t, attr.Required, name+" must be required")
			assert.False(t, attr.Computed, name+" must not be computed")
		}
	})

	// REQ-126: exactly these define the policy and its scope.
	t.Run("only the scope attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "action", "object_type", "object_id",
		}, forceNew)
	})

	t.Run("target_tenant is updated in place", func(t *testing.T) {
		require.Contains(t, s, "target_tenant")
		assert.False(t, s["target_tenant"].ForceNew)
	})
}

// TestNetworkingRBACPolicyV2ActionValidation covers REQ-127: the action set is
// closed and checked before the API is called.
func TestNetworkingRBACPolicyV2ActionValidation(t *testing.T) {
	validate := resourceNetworkingRBACPolicyV2().Schema["action"].ValidateFunc
	require.NotNil(t, validate)

	for _, action := range []string{"access_as_external", "access_as_shared"} {
		t.Run("accepts "+action, func(t *testing.T) {
			_, errs := validate(action, "action")
			assert.Empty(t, errs)
		})
	}

	for _, action := range []string{"", "access_as_public", "ACCESS_AS_SHARED", "shared"} {
		t.Run("rejects "+action, func(t *testing.T) {
			_, errs := validate(action, "action")
			assert.NotEmpty(t, errs)
		})
	}
}

// TestNetworkingRBACPolicyV2ObjectTypeValidation covers REQ-127 and SCN-127-03:
// the set is the seven types upstream Neutron validates, including the two this
// migration adds no resource for.
func TestNetworkingRBACPolicyV2ObjectTypeValidation(t *testing.T) {
	validate := resourceNetworkingRBACPolicyV2().Schema["object_type"].ValidateFunc
	require.NotNil(t, validate)

	// All seven, not just the five that get their own resource here: a policy can
	// reference an existing object this provider does not manage.
	for _, objectType := range []string{
		"network", "qos_policy", "security_group", "subnetpool",
		"address_scope", "address_group", "bgpvpn",
	} {
		t.Run("accepts "+objectType, func(t *testing.T) {
			_, errs := validate(objectType, "object_type")
			assert.Empty(t, errs)
		})
	}

	for _, objectType := range []string{"", "subnet", "port", "router", "floatingip", "Network"} {
		t.Run("rejects "+objectType, func(t *testing.T) {
			_, errs := validate(objectType, "object_type")
			assert.NotEmpty(t, errs)
		})
	}
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingRBACPolicyV2Basic covers SCN-125-01, SCN-125-02, SCN-129-03
// and SCN-130-01.
func TestAccNetworkingRBACPolicyV2Basic(t *testing.T) {
	var found rbacpolicy.Policy
	projectID, region := networkingV2TestEnv()
	targets := networkingRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRBACPolicyV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "action", "access_as_shared"),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "object_type", "network"),
					targets.check(targets.first),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRBACPolicyV2ResourceName, "object_id",
						"selectel_networking_network_v2.network_tf_acc_test_1", "id",
					),
				),
			},
			{
				// SCN-125-02: the follow-up plan is empty.
				Config:   testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				PlanOnly: true,
			},
			{
				// SCN-129-03: import by the policy identifier alone.
				ResourceName:      testAccNetworkingRBACPolicyV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2Update covers SCN-126-01 and SCN-126-02: the
// target project changes in place, and the scope attributes replace the policy.
func TestAccNetworkingRBACPolicyV2Update(t *testing.T) {
	var found rbacpolicy.Policy
	projectID, region := networkingV2TestEnv()
	targets := networkingRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:  testAccCheckNetworkingRBACPolicyV2Exists(&found),
			},
			{
				// SCN-126-01: only target_tenant changes, so the policy keeps its id.
				Config: testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.second),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRBACPolicyV2SameID(&found),
					targets.check(targets.second),
				),
			},
			{
				// SCN-126-02: a different object_id replaces the policy.
				Config: testAccNetworkingRBACPolicyV2SecondNetwork(projectID, region, targets.prefix, targets.second),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRBACPolicyV2Replaced(&found),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRBACPolicyV2ResourceName, "object_id",
						"selectel_networking_network_v2.network_tf_acc_test_2", "id",
					),
				),
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2LocalValidation covers SCN-127-01: an unsupported
// action or object_type is rejected before any API call.
func TestAccNetworkingRBACPolicyV2LocalValidation(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingRBACPolicyV2Invalid(projectID, region, "access_as_public", "network"),
				ExpectError: regexp.MustCompile(`expected action to be one of`),
			},
			{
				Config:      testAccNetworkingRBACPolicyV2Invalid(projectID, region, "access_as_shared", "subnet"),
				ExpectError: regexp.MustCompile(`expected object_type to be one of`),
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2UnsupportedCombination covers SCN-127-02: both
// values are individually valid, but the API refuses the pair.
func TestAccNetworkingRBACPolicyV2UnsupportedCombination(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	targets := networkingRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				// access_as_external does not apply to a security group.
				Config: testAccNetworkingRBACPolicyV2Combination(
					projectID, region, targets.prefix, targets.first,
					"access_as_external", "security_group"),
				ExpectError: regexp.MustCompile(`(?i)error creating|invalid|not supported|bad request`),
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2ExternalDeletion covers SCN-129-01 and SCN-130-02.
func TestAccNetworkingRBACPolicyV2ExternalDeletion(t *testing.T) {
	var found rbacpolicy.Policy
	projectID, region := networkingV2TestEnv()
	targets := networkingRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:  testAccCheckNetworkingRBACPolicyV2Exists(&found),
			},
			{
				Config:             testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:              testAccCheckNetworkingRBACPolicyV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				// The plan re-creates it rather than erroring out.
				Config: testAccNetworkingRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:  testAccCheckNetworkingRBACPolicyV2Exists(&found),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingRBACPolicyV2Exists(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRBACPolicyV2ResourceName)
		if err != nil {
			return err
		}

		result, err := rbacpolicy.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingRBACPolicyV2SameID(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRBACPolicyV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRBACPolicyV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking rbac policy was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingRBACPolicyV2Replaced(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRBACPolicyV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRBACPolicyV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("networking rbac policy %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckNetworkingRBACPolicyV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRBACPolicyV2ResourceName)
		if err != nil {
			return err
		}

		return rbacpolicy.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingRBACPolicyV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_rbac_policy_v2" {
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

		if _, err := rbacpolicy.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking rbac policy still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingRBACPolicyV2Prerequisites is the network the policy grants
// access to plus the two projects it grants it to. A grant needs real second and
// third tenants, and an empty project costs nothing to create, so the test makes
// its own instead of depending on projects somebody prepared by hand.
func testAccNetworkingRBACPolicyV2Prerequisites(projectID, region, targetPrefix string) string {
	config := testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "network-tf-acc-test-1"
}
`, region)

	if targetPrefix == "" {
		return config
	}

	return config + fmt.Sprintf(`
resource "selectel_vpc_project_v2" "rbac_target_1" {
  name = "%[1]s-1"
}

resource "selectel_vpc_project_v2" "rbac_target_2" {
  name = "%[1]s-2"
}
`, targetPrefix)
}

// The target argument of the configurations below is the address of one of the
// two projects created above, so target_tenant is a reference rather than an
// identifier carried in from the environment.
const (
	testAccNetworkingRBACPolicyV2FirstTarget  = "selectel_vpc_project_v2.rbac_target_1"
	testAccNetworkingRBACPolicyV2SecondTarget = "selectel_vpc_project_v2.rbac_target_2"
)

func testAccNetworkingRBACPolicyV2Basic(projectID, region, targetPrefix, target string) string {
	return testAccNetworkingRBACPolicyV2Prerequisites(projectID, region, targetPrefix) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  target_tenant = %[2]s
}
`, region, target)
}

// testAccNetworkingRBACPolicyV2SecondNetwork points the policy at another
// network, which is a ForceNew change of object_id.
func testAccNetworkingRBACPolicyV2SecondNetwork(projectID, region, targetPrefix, target string) string {
	return testAccNetworkingRBACPolicyV2Prerequisites(projectID, region, targetPrefix) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-2"
}

resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_2.id
  target_tenant = %[2]s
}
`, region, target)
}

// testAccNetworkingRBACPolicyV2Invalid uses a literal object_id: the schema
// validation must fail before anything is resolved or called.
func testAccNetworkingRBACPolicyV2Invalid(projectID, region, action, objectType string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  action        = "%[2]s"
  object_type   = "%[3]s"
  object_id     = "1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809"
  target_tenant = "b4c5d6e7f80912a35e6f70811a2b3c4d"
}
`, region, action, objectType)
}

func testAccNetworkingRBACPolicyV2Combination(projectID, region, targetPrefix, target, action, objectType string) string {
	return testAccNetworkingRBACPolicyV2Prerequisites(projectID, region, targetPrefix) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%[1]s"
  action        = "%[2]s"
  object_type   = "%[3]s"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  target_tenant = %[4]s
}
`, region, action, objectType, target)
}
