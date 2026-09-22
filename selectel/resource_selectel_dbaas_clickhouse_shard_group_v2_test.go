package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
)

func TestAccDBaaSClickhouseShardGroupV2Basic(t *testing.T) {
	var dbaasDatastore dbaas_v2_ch.DatastoreResponse

	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	logGroup := fmt.Sprintf("s/dbaas/%s", datastoreName)
	shardNamesGr1 := "[\"shard1\", \"shard2\"]"
	shardNamesGr2 := "[\"shard1\"]"
	descGr1 := "my first for gr1"
	descGr2 := ""

	updatedShardNamesGr1 := "[\"shard2\"]"
	updatedShardNamesGr2 := "[\"shard1\", \"shard2\"]"
	updatedDescGr1 := "updated descr for gr1"
	updatedDescGr2 := "description = \"new descr for gr2\""

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccDBaaSV2PreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckDBaaSV2ClickhouseDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSClickhouseShardGroupV2Basic(datastoreName, shardNamesGr1, shardNamesGr2, descGr1, descGr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBaaSV2ClickhouseDatastoreExists(resourceDBaaSClickhouseDatastoreV2Name, &dbaasDatastore),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", datastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "region", dbaasRegion),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "status", string(dbaas_v2_common.DatastoreStatusActive)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "state", string(dbaas_v2_common.DatastoreStateRunning)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.#", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.name", "keepers"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.name", "shard1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.2.name", "shard2"),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "log_platform.0.log_group", logGroup),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.#", "1"),
					resource.TestCheckResourceAttrSet(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.0"), // first item is not empty string

					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "name", "gr1"),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "description", descGr1),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "shard_names.#", "2"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "shard_names.*", "shard1"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "shard_names.*", "shard2"),

					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "name", "gr2"),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "description", ""),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "shard_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "shard_names.*", "shard1"),
				),
			},
			// Update shard groups
			{
				Config: testAccDBaaSClickhouseShardGroupV2Basic(datastoreName, updatedShardNamesGr1, updatedShardNamesGr2, updatedDescGr1, updatedDescGr2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "name", "gr1"),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "description", updatedDescGr1),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "shard_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_1", "shard_names.*", "shard2"),

					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "name", "gr2"),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "description", "new descr for gr2"),
					resource.TestCheckResourceAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "shard_names.#", "2"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "shard_names.*", "shard1"),
					resource.TestCheckTypeSetElemAttr("selectel_dbaas_clickhouse_shard_group_v2.shard_gr_test_2", "shard_names.*", "shard2"),
				),
			},
		},
	})
}

func testAccDBaaSClickhouseShardGroupV2Basic(datastoreName, shardNamesGr1, shardNamesGr2, descGr1, descGr2 string) string {
	return fmt.Sprintf(`
locals {
  project_id      = "%s"
  region_name     = "%s"
}

provider openstack {
	tenant_id = local.project_id
}

resource "openstack_networking_secgroup_v2" "ds_sg" {
  name        = "secgroup_test"
}

resource "openstack_networking_network_v2" "ds_net" {
 	region = local.region_name
  	name = "network_test"
}

resource "openstack_networking_subnet_v2" "ds_subnet" {
  network_id = openstack_networking_network_v2.ds_net.id
  cidr       = "192.168.1.0/24"
  ip_version = 4
  enable_dhcp = false
  name = "subnet_test"
}

data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = local.project_id
  region = local.region_name
  filter {
    engine = "clickhouse"
    version = "26.3.12.3"

  }
}

data "selectel_dbaas_flavor_v2" "keeper_flavor" {
  project_id = local.project_id
  region     = local.region_name
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
	allowed_role = "KEEPER"
  }
}

resource "selectel_dbaas_clickhouse_datastore_v2" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = local.project_id
  region = local.region_name
  type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
  subnet_id = "${openstack_networking_subnet_v2.ds_subnet.id}"
  password = "Iu2YgYlk!ORz"
  security_groups = ["${openstack_networking_secgroup_v2.ds_sg.id}"]
  log_platform {
     log_group = "s/dbaas/%s" 
  }

  node_groups {
	name       = "keepers" 
	role       = "KEEPER"
	node_count = 3
	flavor {
	id    = "${data.selectel_dbaas_flavor_v2.keeper_flavor.flavors[0].id}"
	type  = "FIXED"
	}
  }

  node_groups {
    name       = "shard1" 
    role       = "DATA"
    node_count = 1
	weight     = 100
    flavor {
      type  = "FLEXIBLE"
      vcpus = 2
      ram   = 4096
      disk  = 32
      disk_type = "LOCAL"
    }
  }
  node_groups {
    name       = "shard2" 
    role       = "DATA"
    node_count = 1
	weight     = 50
    flavor {
      type  = "FLEXIBLE"
      vcpus = 2
      ram   = 4096
      disk  = 32
      disk_type = "NETWORK_ULTRA"
    }
  }
}

resource "selectel_dbaas_clickhouse_shard_group_v2" "shard_gr_test_1" {
  name = "gr1"
  project_id = local.project_id
  region = local.region_name
  datastore_id = "${selectel_dbaas_clickhouse_datastore_v2.datastore_tf_acc_test_1.id}"
  shard_names = %s
  description = "%s"
}

resource "selectel_dbaas_clickhouse_shard_group_v2" "shard_gr_test_2" {
  name = "gr2"
  project_id = local.project_id
  region = local.region_name
  datastore_id = "${selectel_dbaas_clickhouse_datastore_v2.datastore_tf_acc_test_1.id}"
  shard_names = %s
  // description
  %s
}
`, dbaasProjectID, dbaasRegion, datastoreName, datastoreName, shardNamesGr1, descGr1, shardNamesGr2, descGr2)
}
