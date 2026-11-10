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

func TestAccDBaaSRolesV1Basic(t *testing.T) {
	var dbaasRoles []dbaas.Role

	projectID := os.Getenv("INFRA_PROJECT_ID")
	datastoreTypeEngine := "postgresql"
	datastoreTypeVersion := "16"
	parameterName := "pg_read_all_stats"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSRolesV1Basic(projectID, datastoreTypeEngine, datastoreTypeVersion, parameterName),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSRolesV1Exists("data.selectel_dbaas_roles_v1.roles_tf_acc_test_1", &dbaasRoles),
					resource.TestCheckResourceAttr("data.selectel_dbaas_roles_v1.roles_tf_acc_test_1", "roles.0.name", parameterName),
				),
			},
		},
	})
}

func testAccDBaaSRolesV1Exists(n string, dbaasRoles *[]dbaas.Role) resource.TestCheckFunc {
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

func testAccDBaaSRolesV1Basic(projectID, engine, version, name string) string {
	return fmt.Sprintf(`
		locals {
		  project_id = "%s"
		}

		data "selectel_dbaas_datastore_type_v1" "dt" {
			project_id = local.project_id
			region     = "ru-3"
			filter {
				engine = "%s"
				version = "%s"
			}
		}

		data "selectel_dbaas_roles_v1" "roles_tf_acc_test_1" {
			project_id = local.project_id
			region     = "ru-3"
			filter {
				datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
				name = "%s"
			}
		}

		output "config" {
			value = data.selectel_dbaas_roles_v1.roles_tf_acc_test_1.roles[0]
		}
	`, projectID, engine, version, name)
}
