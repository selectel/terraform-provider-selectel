package selectel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/rbacpolicy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVPCRBACPolicyV2CustomizeDiff(t *testing.T) {
	for _, c := range []struct {
		name       string
		action     string
		objectType string
		wantErr    bool
	}{
		{name: "external access to a network", action: "access_as_external", objectType: "network"},
		{name: "shared access to a network", action: "access_as_shared", objectType: "network"},
		{name: "shared access to a security group", action: "access_as_shared", objectType: "security_group"},
		{name: "external access to a security group", action: "access_as_external", objectType: "security_group", wantErr: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := planVPCV2ResourceDiff(t, resourceVPCRBACPolicyV2(), map[string]cty.Value{
				"project_id":        cty.StringVal("project-id"),
				"region":            cty.StringVal("ru-1"),
				"action":            cty.StringVal(c.action),
				"object_type":       cty.StringVal(c.objectType),
				"object_id":         cty.StringVal("object-id"),
				"target_project_id": cty.StringVal("target-project-id"),
			})

			if !c.wantErr {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), "action")
			assert.Contains(t, err.Error(), "object_type")
		})
	}
}

const testAccVPCRBACPolicyV2ResourceName = "selectel_vpc_rbac_policy_v2.rbac_tf_acc_test_1"

type vpcRBACPolicyV2Targets struct {
	prefix string
	first  string
	second string
}

func vpcRBACPolicyV2TestTargets() vpcRBACPolicyV2Targets {
	return vpcRBACPolicyV2Targets{
		prefix: acctest.RandomWithPrefix("tf-acc-rbac-target"),
		first:  testAccVPCRBACPolicyV2FirstTarget + ".id",
		second: testAccVPCRBACPolicyV2SecondTarget + ".id",
	}
}

func (targets vpcRBACPolicyV2Targets) check(target string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrPair(
		testAccVPCRBACPolicyV2ResourceName, "target_project_id",
		strings.TrimSuffix(target, ".id"), "id",
	)
}

func TestAccVPCRBACPolicyV2Basic(t *testing.T) {
	var found rbacpolicy.Policy
	projectID, region := vpcV2TestEnv()
	targets := vpcRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRBACPolicyV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCRBACPolicyV2ResourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(testAccVPCRBACPolicyV2ResourceName, "region", region),
					resource.TestCheckResourceAttr(testAccVPCRBACPolicyV2ResourceName, "action", "access_as_shared"),
					resource.TestCheckResourceAttr(testAccVPCRBACPolicyV2ResourceName, "object_type", "network"),
					targets.check(targets.first),
					resource.TestCheckResourceAttrPair(
						testAccVPCRBACPolicyV2ResourceName, "object_id",
						"selectel_vpc_network_v2.network_tf_acc_test_1", "id",
					),
				),
			},
			{
				Config:   testAccVPCRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCRBACPolicyV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccVPCRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:              testAccCheckVPCV2DeleteOutOfBand(testAccVPCRBACPolicyV2ResourceName, rbacpolicy.Delete),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRBACPolicyV2Update(t *testing.T) {
	var found rbacpolicy.Policy
	projectID, region := vpcV2TestEnv()
	targets := vpcRBACPolicyV2TestTargets()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRBACPolicyV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRBACPolicyV2Basic(projectID, region, targets.prefix, targets.first),
				Check:  testAccCheckVPCRBACPolicyV2Exists(&found),
			},
			{
				Config: testAccVPCRBACPolicyV2Basic(projectID, region, targets.prefix, targets.second),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRBACPolicyV2SameID(&found),
					targets.check(targets.second),
				),
			},
			{
				Config: testAccVPCRBACPolicyV2SecondNetwork(projectID, region, targets.prefix, targets.second),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRBACPolicyV2Replaced(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCRBACPolicyV2ResourceName, "object_id",
						"selectel_vpc_network_v2.network_tf_acc_test_2", "id",
					),
				),
			},
		},
	})
}

func testAccCheckVPCRBACPolicyV2Exists(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRBACPolicyV2ResourceName)
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

func testAccCheckVPCRBACPolicyV2SameID(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRBACPolicyV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRBACPolicyV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC rbac policy was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCRBACPolicyV2Replaced(found *rbacpolicy.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRBACPolicyV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRBACPolicyV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("VPC rbac policy %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckVPCRBACPolicyV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_rbac_policy_v2" {
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

		if _, err := rbacpolicy.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC rbac policy still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCRBACPolicyV2Prerequisites(projectID, region, targetPrefix string) string {
	config := testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
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

const (
	testAccVPCRBACPolicyV2FirstTarget  = "selectel_vpc_project_v2.rbac_target_1"
	testAccVPCRBACPolicyV2SecondTarget = "selectel_vpc_project_v2.rbac_target_2"
)

func testAccVPCRBACPolicyV2Basic(projectID, region, targetPrefix, target string) string {
	return testAccVPCRBACPolicyV2Prerequisites(projectID, region, targetPrefix) + fmt.Sprintf(`
resource "selectel_vpc_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%[1]s"
  action            = "access_as_shared"
  object_type       = "network"
  object_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  target_project_id = %[2]s
}
`, region, target)
}

func testAccVPCRBACPolicyV2SecondNetwork(projectID, region, targetPrefix, target string) string {
	return testAccVPCRBACPolicyV2Prerequisites(projectID, region, targetPrefix) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-2"
}

resource "selectel_vpc_rbac_policy_v2" "rbac_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%[1]s"
  action            = "access_as_shared"
  object_type       = "network"
  object_id         = selectel_vpc_network_v2.network_tf_acc_test_2.id
  target_project_id = %[2]s
}
`, region, target)
}
