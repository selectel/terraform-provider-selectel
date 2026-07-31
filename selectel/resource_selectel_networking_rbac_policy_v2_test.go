package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/rbacpolicy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingRBACPolicyV2ResourceName = "selectel_networking_rbac_policy_v2.rbac_tf_acc_test_1"

// networkingRBACPolicyV2TestTargetProject is the project the policy grants access
// to. It has to be a second, real project, so the test skips when none is
// configured rather than sharing the network with itself.
func networkingRBACPolicyV2TestTargetProject(t *testing.T) string {
	t.Helper()

	projectID := os.Getenv("SELECTEL_RBAC_TARGET_PROJECT_ID")
	if projectID == "" {
		t.Skip("SELECTEL_RBAC_TARGET_PROJECT_ID must be set to the id of a second project")
	}

	return projectID
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
	projectName := acctest.RandomWithPrefix("tf-acc")
	targetProjectID := networkingRBACPolicyV2TestTargetProject(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRBACPolicyV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "action", "access_as_shared"),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "object_type", "network"),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "target_tenant", targetProjectID),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRBACPolicyV2ResourceName, "object_id",
						"selectel_networking_network_v2.network_tf_acc_test_1", "id",
					),
				),
			},
			{
				// SCN-125-02: the follow-up plan is empty.
				Config:   testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
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
	projectName := acctest.RandomWithPrefix("tf-acc")
	targetProjectID := networkingRBACPolicyV2TestTargetProject(t)

	secondTargetProjectID := os.Getenv("SELECTEL_RBAC_TARGET_PROJECT_ID_ALTERNATIVE")
	if secondTargetProjectID == "" {
		t.Skip("SELECTEL_RBAC_TARGET_PROJECT_ID_ALTERNATIVE must name a third project to move the grant to")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
				Check:  testAccCheckNetworkingRBACPolicyV2Exists(&found),
			},
			{
				// SCN-126-01: only target_tenant changes, so the policy keeps its id.
				Config: testAccNetworkingRBACPolicyV2Basic(projectName, secondTargetProjectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRBACPolicyV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRBACPolicyV2ResourceName, "target_tenant", secondTargetProjectID),
				),
			},
			{
				// SCN-126-02: a different object_id replaces the policy.
				Config: testAccNetworkingRBACPolicyV2SecondNetwork(projectName, secondTargetProjectID),
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
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingRBACPolicyV2Invalid(projectName, "access_as_public", "network"),
				ExpectError: regexp.MustCompile(`expected action to be one of`),
			},
			{
				Config:      testAccNetworkingRBACPolicyV2Invalid(projectName, "access_as_shared", "subnet"),
				ExpectError: regexp.MustCompile(`expected object_type to be one of`),
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2UnsupportedCombination covers SCN-127-02: both
// values are individually valid, but the API refuses the pair.
func TestAccNetworkingRBACPolicyV2UnsupportedCombination(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	targetProjectID := networkingRBACPolicyV2TestTargetProject(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				// access_as_external does not apply to a security group.
				Config: testAccNetworkingRBACPolicyV2Combination(
					projectName, targetProjectID, "access_as_external", "security_group"),
				ExpectError: regexp.MustCompile(`(?i)error creating|invalid|not supported|bad request`),
			},
		},
	})
}

// TestAccNetworkingRBACPolicyV2ExternalDeletion covers SCN-129-01 and SCN-130-02.
func TestAccNetworkingRBACPolicyV2ExternalDeletion(t *testing.T) {
	var found rbacpolicy.Policy
	projectName := acctest.RandomWithPrefix("tf-acc")
	targetProjectID := networkingRBACPolicyV2TestTargetProject(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
				Check:  testAccCheckNetworkingRBACPolicyV2Exists(&found),
			},
			{
				Config:             testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
				Check:              testAccCheckNetworkingRBACPolicyV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				// The plan re-creates it rather than erroring out.
				Config: testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID),
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
			return nil //nolint:nilerr // the project is gone together with the policy
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

// testAccNetworkingRBACPolicyV2Network is the shared network the policy grants
// access to.
func testAccNetworkingRBACPolicyV2Network(projectName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  name       = "network-tf-acc-test-1"
}
`, networkingV2TestRegion())
}

func testAccNetworkingRBACPolicyV2Basic(projectName, targetProjectID string) string {
	return testAccNetworkingRBACPolicyV2Network(projectName) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region        = "%s"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  target_tenant = "%s"
}
`, networkingV2TestRegion(), targetProjectID)
}

// testAccNetworkingRBACPolicyV2SecondNetwork points the policy at another
// network, which is a ForceNew change of object_id.
func testAccNetworkingRBACPolicyV2SecondNetwork(projectName, targetProjectID string) string {
	return testAccNetworkingRBACPolicyV2Network(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_2" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%[1]s"
  name       = "network-tf-acc-test-2"
}

resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region        = "%[1]s"
  action        = "access_as_shared"
  object_type   = "network"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_2.id
  target_tenant = "%[2]s"
}
`, networkingV2TestRegion(), targetProjectID)
}

// testAccNetworkingRBACPolicyV2Invalid uses a literal object_id: the schema
// validation must fail before anything is resolved or called.
func testAccNetworkingRBACPolicyV2Invalid(projectName, action, objectType string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region        = "%[1]s"
  action        = "%[2]s"
  object_type   = "%[3]s"
  object_id     = "1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809"
  target_tenant = "b4c5d6e7f80912a35e6f70811a2b3c4d"
}
`, networkingV2TestRegion(), action, objectType)
}

func testAccNetworkingRBACPolicyV2Combination(projectName, targetProjectID, action, objectType string) string {
	return testAccNetworkingRBACPolicyV2Network(projectName) + fmt.Sprintf(`
resource "selectel_networking_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id    = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region        = "%[1]s"
  action        = "%[2]s"
  object_type   = "%[3]s"
  object_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  target_tenant = "%[4]s"
}
`, networkingV2TestRegion(), action, objectType, targetProjectID)
}
