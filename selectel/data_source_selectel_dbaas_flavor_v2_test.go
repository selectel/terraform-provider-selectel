package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dbaas_v2 "github.com/selectel/dbaas-go/v2/common"
)

func TestAccDBaaSV2FlavorsBasic(t *testing.T) {
	var dbaasFlavors []dbaas_v2.FlavorResponse

	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSV2FlavorsBasic(projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSV2FlavorsExists("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.fl_size"),
				),
			},
			{
				Config: testAccDBaaSV2FlavorsClickHouseFlavor(projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSV2FlavorsExists("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.fl_size"),
					resource.TestCheckTypeSetElemAttrPair("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.datastore_type_ids.*", "data.selectel_dbaas_datastore_type_v2.dt", "datastore_types.0.id"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.allowed_roles.#", "2"),
				),
			},
		},
	})
}

func testAccDBaaSV2FlavorsExists(n string, dbaasFlavors *[]dbaas_v2.FlavorResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSV2Client(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		response, err := dbaasClient.Flavor.GetFlavorList(ctx)
		if err != nil {
			return err
		}

		*dbaasFlavors = response.Flavors

		return nil
	}
}

func testAccDBaaSV2FlavorsBasic(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_flavor_v2" "flavor_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-1"
}
`, projectID)
}

func testAccDBaaSV2FlavorsClickHouseFlavor(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = local.project_id
  region     = "ru-1"
  filter {
    engine = "clickhouse"
  }
}

data "selectel_dbaas_flavor_v2" "flavor_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-1"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
	allowed_role = "DATA"
  }
}
`, projectID)
}
