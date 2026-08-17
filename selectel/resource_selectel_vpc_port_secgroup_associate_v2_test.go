package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
)

const testAccVPCPortSecGroupAssociateV2ResourceName = "selectel_vpc_port_secgroup_associate_v2.associate_tf_acc_test_1"

func TestAccVPCPortSecGroupAssociateV2Additive(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortSecGroupAssociateV2Additive(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCPortSecGroupAssociateV2ResourceName, "security_group_ids.#", "1"),
					testAccCheckVPCPortSecGroupAssociateV2PortHasAtLeast(2),
					testAccCheckVPCPortSecGroupAssociateV2IDIsPort(),
				),
			},
			{
				Config:   testAccVPCPortSecGroupAssociateV2Additive(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:            testAccVPCPortSecGroupAssociateV2ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids", "enforce"},
			},
		},
	})
}

func TestAccVPCPortSecGroupAssociateV2Enforce(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortSecGroupAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortSecGroupAssociateV2Enforce(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVPCPortSecGroupAssociateV2ResourceName, "security_group_ids.#", "1"),
					testAccCheckVPCPortSecGroupAssociateV2PortHas(1),
				),
			},
			{
				Config:   testAccVPCPortSecGroupAssociateV2Enforce(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVPCPortSecGroupAssociateV2IDIsPort() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCPortSecGroupAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCPortSecGroupAssociateV2ResourceName)
		}

		if rs.Primary.ID != rs.Primary.Attributes["port_id"] {
			return fmt.Errorf("resource id %s is not the port id %s",
				rs.Primary.ID, rs.Primary.Attributes["port_id"])
		}

		return nil
	}
}

func testAccCheckVPCPortSecGroupAssociateV2PortHas(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groups, err := testAccVPCPortSecGroupAssociateV2Groups(s)
		if err != nil {
			return err
		}

		if len(groups) != expected {
			return fmt.Errorf("the port has %d security groups, expected %d", len(groups), expected)
		}

		return nil
	}
}

func testAccCheckVPCPortSecGroupAssociateV2PortHasAtLeast(minimum int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		groups, err := testAccVPCPortSecGroupAssociateV2Groups(s)
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

func testAccVPCPortSecGroupAssociateV2Groups(s *terraform.State) ([]string, error) {
	handle, err := testAccVPCNetworkV2Handle(s, testAccVPCPortSecGroupAssociateV2ResourceName)
	if err != nil {
		return nil, err
	}

	found, err := port.Get(context.Background(), handle.client, handle.id)
	if err != nil {
		return nil, err
	}

	return found.SecurityGroups, nil
}

func testAccCheckVPCPortSecGroupAssociateV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_port_secgroup_associate_v2" {
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

		found, err := port.Get(context.Background(), client, rs.Primary.ID)
		if err != nil {
			if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
				continue
			}

			return err
		}

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

func testAccVPCPortSecGroupAssociateV2Base(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_vpc_network_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}

resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "secgroup-existing"
}

resource "selectel_vpc_secgroup_v2" "secgroup_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "secgroup-associated"
}

resource "selectel_vpc_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%[1]s"
  network_id         = selectel_vpc_network_v2.network_tf_acc_test_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_1.id]

  fixed_ip {
    subnet_id = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
  }
}
`, region)
}

func testAccVPCPortSecGroupAssociateV2Additive(projectID, region string) string {
	return testAccVPCPortSecGroupAssociateV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_secgroup_associate_v2" "associate_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%s"
  port_id            = selectel_vpc_port_v2.port_tf_acc_test_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2.id]
}
`, region)
}

func testAccVPCPortSecGroupAssociateV2Enforce(projectID, region string) string {
	return testAccVPCPortSecGroupAssociateV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_secgroup_associate_v2" "associate_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%s"
  port_id            = selectel_vpc_port_v2.port_tf_acc_test_1.id
  security_group_ids = [selectel_vpc_secgroup_v2.secgroup_tf_acc_test_2.id]
  enforce            = true
}
`, region)
}
