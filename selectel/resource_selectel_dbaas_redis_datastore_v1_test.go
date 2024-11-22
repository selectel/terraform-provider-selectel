package selectel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSRedisDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1
	resizeNodeCount := 2

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSRedisDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "volatile-lru"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSRedisDatastoreV1UpdateConfig(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSRedisDatastoreV1UpdatePassword(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSRedisDatastoreV1Resize(projectName, datastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func testAccDBaaSRedisDatastoreV1Basic(projectName, datastoreName string, nodeCount int) string {
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
    engine = "redis"
    version = "6"
  }
}

data "selectel_dbaas_flavor_v1" "flavor" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  }
}

resource "selectel_dbaas_redis_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor_id = "${data.selectel_dbaas_flavor_v1.flavor.flavors[0].id}"
  config = {
    maxmemory-policy = "volatile-lru"
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSRedisDatastoreV1UpdateConfig(projectName, datastoreName string, nodeCount int) string {
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
    engine = "redis"
    version = "6"
  }
}

data "selectel_dbaas_flavor_v1" "flavor" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  }
}

resource "selectel_dbaas_redis_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor_id = "${data.selectel_dbaas_flavor_v1.flavor.flavors[0].id}"
  config = {
    maxmemory-policy = "noeviction"
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSRedisDatastoreV1UpdatePassword(projectName, datastoreName string, nodeCount int) string {
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
    engine = "redis"
    version = "6"
  }
}

data "selectel_dbaas_flavor_v1" "flavor" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  }
}

resource "selectel_dbaas_redis_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor_id = "${data.selectel_dbaas_flavor_v1.flavor.flavors[0].id}"
  config = {
    maxmemory-policy = "noeviction"
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6123"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSRedisDatastoreV1Resize(projectName, datastoreName string, nodeCount int) string {
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
    engine = "redis"
    version = "6"
  }
}

data "selectel_dbaas_flavor_v1" "flavor" {
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	region     = "ru-3"
	filter {
	  datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
	}
  }

resource "selectel_dbaas_redis_datastore_v1" "datastore_tf_acc_test_1" {
name = "%s"
project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
region = "ru-3"
type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
node_count = "%d"
flavor_id = "${data.selectel_dbaas_flavor_v1.flavor.flavors[0].id}"
config = {
	maxmemory-policy = "noeviction"
}
redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6123"
}`, projectName, datastoreName, nodeCount)
}
