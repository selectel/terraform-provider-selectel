package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSKafkaDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSKafkaDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "config.log.retention.ms", "1000"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSKafkaDatastoreV1UpdateConfig(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "config.log.retention.ms", "10000"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "config.log.retention.bytes", "1024"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSKafkaDatastoreV1Resize(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(64)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "config.log.retention.ms", "1000"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "config.log.retention.bytes", "1024"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func testAccDBaaSKafkaDatastoreV1Basic(projectName, datastoreName string, nodeCount int) string {
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
    engine = "kafka"
    version = "3.5"
  }
}

resource "selectel_dbaas_kafka_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 8192
    disk = 32
  }
  config = {
    "log.retention.ms" = 1000
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSKafkaDatastoreV1UpdateConfig(projectName, datastoreName string, nodeCount int) string {
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
    engine = "kafka"
    version = "3.5"
  }
}

resource "selectel_dbaas_kafka_datastore_v1" "datastore_tf_acc_test_1" {
	name = "%s"
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	region = "ru-3"
	type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
	subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
	node_count = "%d"
	flavor {
	  vcpus = 2
	  ram = 8192
	  disk = 32
	}
	config = {
	  "log.retention.ms" = 10000
	  "log.retention.bytes" = 1024
	}
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSKafkaDatastoreV1Resize(projectName, datastoreName string, nodeCount int) string {
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
    engine = "kafka"
    version = "3.5"
  }
}

resource "selectel_dbaas_kafka_datastore_v1" "datastore_tf_acc_test_1" {
	name = "%s"
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	region = "ru-3"
	type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
	subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
	node_count = "%d"
	flavor {
	  vcpus = 2
	  ram = 8192
	  disk = 64
	}
	config = {
	  "log.retention.ms" = 10000
	  "log.retention.bytes" = 1024
	}
}`, projectName, datastoreName, nodeCount)
}
