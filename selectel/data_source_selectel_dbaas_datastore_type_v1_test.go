package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
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

		var projectID, endpoint string
		if id, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = id
		}
		if region, ok := rs.Primary.Attributes["region"]; ok {
			endpoint = getDBaaSV1Endpoint(region)
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		tokenOpts := tokens.TokenOpts{
			ProjectID: projectID,
		}
		token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
		if err != nil {
			return errCreatingObject(objectToken, err)
		}

		dbaasClient, err := dbaas.NewDBAASClient(token.ID, endpoint)
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
  auto_quotas = true
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
