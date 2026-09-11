package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dbaas_v2 "github.com/selectel/dbaas-go/v2/common"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSV2FlavorsBasic(t *testing.T) {
	var (
		dbaasFlavors []dbaas_v2.FlavorResponse
		project      projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSV2FlavorsBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSV2FlavorsExists("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.fl_size"),
				),
			},
			{
				Config: testAccDBaaSV2FlavorsClickHouseFlavor(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSV2FlavorsExists("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", &dbaasFlavors),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.ram"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.disk"),
					resource.TestCheckResourceAttrSet("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.fl_size"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_flavor_v2.flavor_tf_acc_test_1", "flavors.0.datastore_type_ids.#", "2"),
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

func testAccDBaaSV2FlavorsBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_flavor_v2" "flavor_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
}
`, projectName)
}

func testAccDBaaSV2FlavorsClickHouseFlavor(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  filter {
    engine = "clickhouse"
  }
}

data "selectel_dbaas_flavor_v2" "flavor_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
	allowed_role = "DATA"
  }
}
`, projectName)
}
