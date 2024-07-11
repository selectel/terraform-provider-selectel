package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIAMV1GroupMembershipBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1GroupMembershipBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "group_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.0"),
				),
			},
		},
	})
}

func TestAccIAMV1GroupUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1GroupMembershipBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "group_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.0"),
				),
			},
			{
				Config: testAccIAMV1GroupMembershipUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "group_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.0"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.1"),
				),
			},
			{
				Config: testAccIAMV1GroupMembershipBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "group_id"),
					resource.TestCheckResourceAttrSet("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.0"),
					resource.TestCheckNoResourceAttr("selectel_iam_group_membership_v1.membership_tf_acc_test_1", "user_ids.1"),
				),
			},
		},
	})
}

func testAccIAMV1GroupMembershipBasic() string {
	return `
resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "test-service-user-1"
  password    = "Qazwsxedc123"
  role {
    role_name = "reader"
    scope = "account"
  }
}

resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_2" {
  name        = "test-service-user-2"
  password    = "Qazwsxedc123"
  role {
    role_name = "reader"
    scope = "account"
  }
}

resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
	name = "test-group"
	role {
	  	role_name = "reader"
	  	scope = "account"
	}
}

resource "selectel_iam_group_membership_v1" "membership_tf_acc_test_1" {
	group_id = selectel_iam_group_v1.group_tf_acc_test_1.id

	user_ids = [
		selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1.id
	]
}
`
}

func testAccIAMV1GroupMembershipUpdate() string {
	return `
resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_1" {
  name        = "test-service-user-1"
  password    = "Qazwsxedc123"
  role {
    role_name = "reader"
    scope = "account"
  }
}

resource "selectel_iam_serviceuser_v1" "serviceuser_tf_acc_test_2" {
  name        = "test-service-user-2"
  password    = "Qazwsxedc123"
  role {
    role_name = "reader"
    scope = "account"
  }
}

resource "selectel_iam_group_v1" "group_tf_acc_test_1" {
	name = "test-group"
	role {
	  	role_name = "reader"
	  	scope = "account"
	}
}

resource "selectel_iam_group_membership_v1" "membership_tf_acc_test_1" {
	group_id = selectel_iam_group_v1.group_tf_acc_test_1.id

	user_ids = [
		selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1.id,
		selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_2.id
	]
}
`
}
