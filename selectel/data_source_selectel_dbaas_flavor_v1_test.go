package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSFlavorsV1Basic(t *testing.T) {
	var (
		dbaasFlavors []dbaas.FlavorResponse
		project      projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSFlavorsV1Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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
				Config: testAccDBaaSFlavorsV1RedisFlavor(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSFlavorsV1Exists("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.fl_size"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_flavor_v1.flavor_tf_acc_test_1", "flavors.0.datastore_type_ids.#", "1"),
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

func testAccDBaaSFlavorsV1Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_flavor_v1" "flavor_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}
`, projectName)
}

func testAccDBaaSFlavorsV1RedisFlavor(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    engine = "redis"
  }
}

data "selectel_dbaas_flavor_v1" "flavor_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  }
}
`, projectName)
}
