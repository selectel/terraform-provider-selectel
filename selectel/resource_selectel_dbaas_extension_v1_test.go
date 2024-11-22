package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSExtensionV1Basic(t *testing.T) {
	var (
		dbaasExtension dbaas.Extension
		project        projects.Project
	)

	const extensionName = "hstore"

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	userName := RandomWithPrefix("tf_acc_user")
	userPassword := acctest.RandomWithPrefix("tf-acc-pass")
	databaseName := RandomWithPrefix("tf_acc_db")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSExtensionV1Basic(projectName, datastoreName, userName, userPassword, databaseName, extensionName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSExtensionV1Exists("selectel_dbaas_extension_v1.extension_tf_acc_test_1", &dbaasExtension),
					resource.TestCheckResourceAttrSet("selectel_dbaas_extension_v1.extension_tf_acc_test_1", "available_extension_id"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_extension_v1.extension_tf_acc_test_1", "datastore_id"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_extension_v1.extension_tf_acc_test_1", "database_id"),
				),
			},
		},
	})
}

func testAccCheckDBaaSExtensionV1Exists(n string, dbaasExtension *dbaas.Extension) resource.TestCheckFunc {
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

		extension, err := dbaasClient.Extension(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if extension.ID != rs.Primary.ID {
			return errors.New("extension not found")
		}

		*dbaasExtension = extension

		return nil
	}
}

func testAccDBaaSExtensionV1Basic(projectName, datastoreName, userName, userPassword, databaseName, extensionName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "postgresql"
    version = "12"
  }
}

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 4096
    disk = 32
  }
}

resource "selectel_dbaas_user_v1" "user_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  password = "%s"
}

resource "selectel_dbaas_database_v1" "database_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  owner_id = "${selectel_dbaas_user_v1.user_tf_acc_test_1.id}"
}

data "selectel_dbaas_available_extension_v1" "ae" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    name = "%s"
  }
}

resource "selectel_dbaas_extension_v1" "extension_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  available_extension_id = "${data.selectel_dbaas_available_extension_v1.ae.available_extensions[0].id}"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  database_id = "${selectel_dbaas_database_v1.database_tf_acc_test_1.id}"
}`, projectName, datastoreName, nodeCount, userName, userPassword, databaseName, extensionName)
}
