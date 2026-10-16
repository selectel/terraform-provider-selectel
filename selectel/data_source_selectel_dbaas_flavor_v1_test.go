package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
)

func TestAccDBaaSFlavorsV1Basic(t *testing.T) {
	var dbaasFlavors []dbaas.FlavorResponse

	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSFlavorsV1Basic(projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSFlavorsV1Exists("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.fl_size"),
				),
			},
			{
				Config: testAccDBaaSFlavorsV1RedisFlavor(projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSFlavorsV1Exists("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.fl_size"),
					resource.TestCheckTypeSetElemAttrPair("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.datastore_type_ids.*", "data.selectel_dbaas_datastore_type_v1.dt", "datastore_types.0.id"),
				),
			},
		},
	})
}

func testAccDBaaSFlavorsV1Exists(n string, dbaasFlavors *[]dbaas.FlavorResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSClient(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		flavors, err := dbaasClient.Flavors(ctx)
		if err != nil {
			return err
		}

		*dbaasFlavors = flavors

		return nil
	}
}

func testAccDBaaSFlavorsV1Basic(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_flavor_v1" "flavor_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
}
`, projectID)
}

func testAccDBaaSFlavorsV1RedisFlavor(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    engine = "redis"
  }
}

data "selectel_dbaas_flavor_v1" "flavor_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  }
}
`, projectID)
}
