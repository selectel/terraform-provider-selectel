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

func TestAccDBaaSV2DatastoreTypesBasic(t *testing.T) {
	var dbaasDatastoreTypes []dbaas_v2.DatastoreTypeResponse

	projectID := os.Getenv("INFRA_PROJECT_ID")
	datastoreTypeEngine := "clickhouse"
	datastoreTypeVersion := "26.3.12.3"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSV2DatastoreTypesBasic(projectID, datastoreTypeEngine, datastoreTypeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSV2DatastoreTypesExists("data.selectel_dbaas_datastore_type_v2.datastore_type_tf_acc_test_1", &dbaasDatastoreTypes),
					resource.TestCheckResourceAttr("data.selectel_dbaas_datastore_type_v2.datastore_type_tf_acc_test_1", "datastore_types.0.engine", datastoreTypeEngine),
					resource.TestCheckResourceAttr("data.selectel_dbaas_datastore_type_v2.datastore_type_tf_acc_test_1", "datastore_types.0.version", datastoreTypeVersion),
				),
			},
		},
	})
}

func testAccDBaaSV2DatastoreTypesExists(n string, dbaasDatastoreTypes *[]dbaas_v2.DatastoreTypeResponse) resource.TestCheckFunc {
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

		datastoreTypes, err := dbaasClient.DatastoreType.GetDatastoreTypeList(ctx)
		if err != nil {
			return err
		}

		*dbaasDatastoreTypes = datastoreTypes.DatastoreTypes

		return nil
	}
}

func testAccDBaaSV2DatastoreTypesBasic(projectID, engine, version string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_datastore_type_v2" "datastore_type_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-1"
  filter {
    engine = "%s"
    version = "%s"
  }
}
`, projectID, engine, version)
}
