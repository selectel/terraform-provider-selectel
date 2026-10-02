package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSRolesV1Basic(t *testing.T) {
	var (
		dbaasRoles []dbaas.Roles
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreTypeEngine := "postgresql"
	datastoreTypeVersion := "12"
	parameterName := "pg_read_all_data"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSRolesV1Basic(projectName, datastoreTypeEngine, datastoreTypeVersion, parameterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSRolesV1Exists("data.selectel_dbaas_roles_v1.roles_tf_acc_test_1", &dbaasRoles),
					resource.TestCheckResourceAttr("data.selectel_dbaas_roles_v1.roles_tf_acc_test_1", "roles.0.name", parameterName),
				),
			},
		},
	})
}

func testAccDBaaSRolesV1Exists(n string, dbaasRoles *[]dbaas.Roles) resource.TestCheckFunc {
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

		roles, err := dbaasClient.Roles(ctx)
		if err != nil {
			return err
		}

		*dbaasRoles = roles

		return nil
	}
}

func testAccDBaaSRolesV1Basic(projectName, engine, version, name string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name        = "%s"
		}

		data "selectel_dbaas_datastore_type_v1" "dt" {
			project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
			region     = "ru-3"
			filter {
				engine = "%s"
				version = "%s"
			}
		}

		data "selectel_dbaas_roles_v1" "roles_tf_acc_test_1" {
			project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
			region     = "ru-3"
			filter {
				datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
				name = "%s"
			}
		}

		output "config" {
			value = data.selectel_dbaas_roles_v1.roles_tf_acc_test_1.roles[0]
		}
	`, projectName, engine, version, name)
}
