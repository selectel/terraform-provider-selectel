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

func TestAccDBaaSDatastoreTypesV1Basic(t *testing.T) {
	var dbaasDatastoreTypes []dbaas.DatastoreType

	projectID := os.Getenv("INFRA_PROJECT_ID")
	datastoreTypeEngine := "postgresql"
	datastoreTypeVersion := "16"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSDatastoreTypesV1Basic(projectID, datastoreTypeEngine, datastoreTypeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSDatastoreTypesV1Exists("data.selectel_dbaas_datastore_type_v1.datastore_type_tf_acc_test_1", &dbaasDatastoreTypes),
					resource.TestCheckResourceAttr("data.selectel_dbaas_datastore_type_v1.datastore_type_tf_acc_test_1", "datastore_types.0.engine", datastoreTypeEngine),
					resource.TestCheckResourceAttr("data.selectel_dbaas_datastore_type_v1.datastore_type_tf_acc_test_1", "datastore_types.0.version", datastoreTypeVersion),
				),
			},
		},
	})
}

func testAccDBaaSDatastoreTypesV1Exists(n string, dbaasDatastoreTypes *[]dbaas.DatastoreType) resource.TestCheckFunc {
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

		datastoreTypes, err := dbaasClient.DatastoreTypes(ctx)
		if err != nil {
			return err
		}

		*dbaasDatastoreTypes = datastoreTypes

		return nil
	}
}

func testAccDBaaSDatastoreTypesV1Basic(projectID, engine, version string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_datastore_type_v1" "datastore_type_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    engine = "%s"
    version = "%s"
  }
}
`, projectID, engine, version)
}
