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

func TestAccDBaaSV2DatastoreTypesBasic(t *testing.T) {
	var (
		dbaasDatastoreTypes []dbaas_v2.DatastoreTypeResponse
		project             projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreTypeEngine := "clickhouse"
	datastoreTypeVersion := "26.3.12.3"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSV2DatastoreTypesBasic(projectName, datastoreTypeEngine, datastoreTypeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testAccDBaaSV2DatastoreTypesBasic(projectName, engine, version string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v2" "datastore_type_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  filter {
    engine = "%s"
    version = "%s"
  }
}
`, projectName, engine, version)
}
