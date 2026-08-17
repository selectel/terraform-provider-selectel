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
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
	"github.com/stretchr/testify/assert"
)

const (
	testAccVPCSecGroupV2ResourceName   = "selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1"
	testAccVPCSecGroupV2DataSourceName = "data.selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1"
)

func TestVPCSecGroupV2Flatten(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceVPCSecGroupV2().Schema, map[string]interface{}{})

	flattenVPCSecGroupV2(d, &securitygroup.SecurityGroup{
		ID:          "sg-id",
		Name:        "web",
		Description: "web traffic",
		Stateful:    false,
		Shared:      true,
		Tags:        []string{"production"},
		SecurityGroupRules: []securitygrouprule.Rule{
			{ID: "rule-egress-v4"},
			{ID: "rule-egress-v6"},
			{ID: "rule-allpass-v4"},
		},
	})

	assert.Equal(t, "web", d.Get("name"))
	assert.Equal(t, "web traffic", d.Get("description"))
	assert.Equal(t, false, d.Get("stateful"))
	assert.Equal(t, true, d.Get("shared"))
	assert.Equal(t, []interface{}{"rule-egress-v4", "rule-egress-v6", "rule-allpass-v4"},
		d.Get("security_group_rules"))
}

func TestAccVPCSecGroupV2Basic(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupV2Exists(&found),
					resource.TestCheckResourceAttrSet(testAccVPCSecGroupV2ResourceName, "stateful"),
					resource.TestCheckResourceAttrSet(testAccVPCSecGroupV2ResourceName, "shared"),
					testAccCheckVPCSecGroupV2RuleCountAtLeast(1),
				),
			},
			{
				Config:   testAccVPCSecGroupV2Basic(projectID, region, "secgroup-1"),
				PlanOnly: true,
			},
			{
				ResourceName:            testAccVPCSecGroupV2ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_default_rules"},
			},
		},
	})
}

func TestAccVPCSecGroupV2DeleteDefaultRules(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupV2DeleteDefaultRules(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "security_group_rules.#", "0"),
					testAccCheckVPCSecGroupV2RuleCount(0),
				),
			},
			{
				Config:   testAccVPCSecGroupV2DeleteDefaultRules(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCSecGroupV2Update(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupV2Described(projectID, region, "secgroup-1", "web traffic", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config: testAccVPCSecGroupV2Described(projectID, region, "secgroup-1-renamed", "changed", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecGroupV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "name", "secgroup-1-renamed"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "description", "changed"),
					resource.TestCheckResourceAttr(testAccVPCSecGroupV2ResourceName, "stateful", "false"),
				),
			},
			{
				Config: testAccVPCSecGroupV2DeleteDefaultRules(projectID, region),
				Check:  testAccCheckVPCSecGroupV2Replaced(&found),
			},
		},
	})
}

func TestAccVPCSecGroupV2DataSource(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	secGroupName := acctest.RandomWithPrefix("tf-acc-secgroup")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecGroupV2DataSource(projectID, region, secGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "secgroup_id",
						testAccVPCSecGroupV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "shared",
						testAccVPCSecGroupV2ResourceName, "shared",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCSecGroupV2DataSourceName, "security_group_rules.#",
						testAccVPCSecGroupV2ResourceName, "security_group_rules.#",
					),
				),
			},
			{
				Config:      testAccVPCSecGroupV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no VPC security group found`),
			},
		},
	})
}

func testAccCheckVPCSecGroupV2Exists(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		result, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckVPCSecGroupV2SameID(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCSecGroupV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCSecGroupV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC security group was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCSecGroupV2Replaced(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCSecGroupV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCSecGroupV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("VPC security group %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckVPCSecGroupV2RuleCount(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		found, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if len(found.SecurityGroupRules) != expected {
			return fmt.Errorf("the group has %d rules, expected %d",
				len(found.SecurityGroupRules), expected)
		}

		return nil
	}
}

func testAccCheckVPCSecGroupV2RuleCountAtLeast(minimum int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		found, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if len(found.SecurityGroupRules) < minimum {
			return fmt.Errorf("the group has %d rules, expected at least %d",
				len(found.SecurityGroupRules), minimum)
		}

		return nil
	}
}

func testAccCheckVPCSecGroupV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_secgroup_v2" {
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

		if _, err := securitygroup.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC security group still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCSecGroupV2Basic(projectID, region, name string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "%s"
}
`, region, name)
}

func testAccVPCSecGroupV2Described(projectID, region, name, description string, stateful bool) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  name        = "%s"
  description = "%s"
  stateful    = %t
  tags        = ["tf-acc", "VPC"]
}
`, region, name, description, stateful)
}

func testAccVPCSecGroupV2DeleteDefaultRules(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  name                 = "secgroup-tf-acc-test-1"
  delete_default_rules = true
}
`, region)
}

func testAccVPCSecGroupV2DataSource(projectID, region, name string) string {
	return testAccVPCSecGroupV2Basic(projectID, region, name) + fmt.Sprintf(`
data "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.name
}
`, region)
}

func testAccVPCSecGroupV2DataSourceNoMatch(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
data "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-secgroup"
}
`, region)
}
