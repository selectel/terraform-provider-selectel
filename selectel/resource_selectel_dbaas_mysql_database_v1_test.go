package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSMySQLDatabaseV1Basic(t *testing.T) {
	var (
		dbaasDatabase dbaas.Database
		project       projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	databaseName := RandomWithPrefix("tf_acc_db")
	nodeCount := 1
	datastoreTypeEngine := mySQLDatastoreType

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSMySQLDatabaseV1Basic(projectName, datastoreName, datastoreTypeEngine, databaseName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatabaseV1Exists("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", &dbaasDatabase),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", "name", databaseName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
		},
	})
}

func TestAccDBaaSMySQLNativeDatabaseV1Basic(t *testing.T) {
	var (
		dbaasDatabase dbaas.Database
		project       projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	databaseName := RandomWithPrefix("tf_acc_db")
	nodeCount := 1
	datastoreTypeEngine := mySQLNativeDatastoreType

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSMySQLDatabaseV1Basic(projectName, datastoreName, datastoreTypeEngine, databaseName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatabaseV1Exists("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", &dbaasDatabase),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", "name", databaseName),
					resource.TestCheckResourceAttr("selectel_dbaas_mysql_database_v1.database_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
		},
	})
}

func testAccDBaaSMySQLDatabaseV1Basic(projectName, datastoreName, datastoreTypeEngine, databaseName string, nodeCount int) string {
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
    engine = "%s"
    version = "8"
  }
}

resource "selectel_dbaas_mysql_datastore_v1" "datastore_tf_acc_test_1" {
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

resource "selectel_dbaas_mysql_database_v1" "database_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_mysql_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
}`, projectName, datastoreTypeEngine, datastoreName, nodeCount, databaseName)
}
