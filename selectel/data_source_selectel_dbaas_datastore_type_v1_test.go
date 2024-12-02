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

func TestAccDBaaSDatastoreTypesV1Basic(t *testing.T) {
	var (
		dbaasDatastoreTypes []dbaas.DatastoreType
		project             projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreTypeEngine := "postgresql"
	datastoreTypeVersion := "12"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSDatastoreTypesV1Basic(projectName, datastoreTypeEngine, datastoreTypeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testAccDBaaSDatastoreTypesV1Basic(projectName, engine, version string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v1" "datastore_type_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    engine = "%s"
    version = "%s"
  }
}
`, projectName, engine, version)
}
