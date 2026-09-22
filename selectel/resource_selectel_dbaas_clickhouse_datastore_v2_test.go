package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
)

const resourceDBaaSClickhouseDatastoreV2Name = "selectel_dbaas_clickhouse_datastore_v2.datastore_tf_acc_test_1"

var (
	dbaasRegion    = os.Getenv("INFRA_REGION")
	dbaasProjectID = os.Getenv("INFRA_PROJECT_ID")
)

func testAccCheckDBaaSV2ClickhouseDatastoreDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_dbaas_clickhouse_datastore_v2" {
			continue
		}
		ctx := context.Background()
		client, err := newTestDBaaSV2Client(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.ClickHouse.GetDatastore(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"clickhouse datastore %s still exists after destroy",
				rs.Primary.ID,
			)
		}

	}

	return nil
}

func testAccCheckDBaaSV2ClickhouseDatastoreExists(n string, dbaasDatastore *dbaas_v2_ch.DatastoreResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSV2Client(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if datastore.ID != rs.Primary.ID {
			return errors.New("datastore not found")
		}

		*dbaasDatastore = datastore

		return nil
	}
}

func TestAccDBaaSClickhouseDatastoreV2Basic(t *testing.T) {
	var dbaasDatastore dbaas_v2_ch.DatastoreResponse

	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	datastorePassword := "Iu2YgYlk!ORz"
	datastoreSG := ""
	shardOneWeight := 50
	shardOneNodeCount := 1
	shardOneFlavor := dbaas_v2_ch.FlavorForNodeGroupRequest{
		Type:     dbaas_v2_common.FlavorTypeFlexible,
		VCPUs:    2,
		RAM:      4096,
		Disk:     32,
		DiskType: dbaas_v2_common.FlavorDiskLocal,
	}
	shardOneHasPublicIps := false
	keepersBlock := ""
	allowReduceNodes := false

	updatedDatastoreName := acctest.RandomWithPrefix("tf-acc-ds-updated")
	updatedDatastorePassword := "Iu2YgYlk!ORzUpd"
	updatedShardOneWeight := 70
	updatedshardOneNodeCountTwo := 2
	updatedshardOneNodeCountOne := 1
	updatedShardOneFlavor := dbaas_v2_ch.FlavorForNodeGroupRequest{
		Type:     dbaas_v2_common.FlavorTypeFlexible,
		VCPUs:    4,
		RAM:      4096,
		Disk:     35,
		DiskType: dbaas_v2_common.FlavorDiskLocal,
	}
	updatedDatastoreSG := "${openstack_networking_secgroup_v2.ds_sg.id}"
	updatedShardOneHasPublicIps := true
	updatedKeepersBlock := `
	node_groups {
	  name       = "keepers"
	  role       = "KEEPER"
	  node_count = 3

	  flavor {
	    id    = "${data.selectel_dbaas_flavor_v2.keeper_flavor.flavors[0].id}"
	    type  = "FIXED"
	  }
	}
	`
	updAllowReduceNodes := true

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			// Need to create a network by openstack operator beacause 'selectel_vpc_subnet_v2' creates a network with 'public' tag.
			// Cickhouse api denies 'You can not use subnet {subnet_id} because it is a part of the external network'.
			testAccDBaaSV2PreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckDBaaSV2ClickhouseDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(datastoreName, datastorePassword, datastoreSG, keepersBlock, shardOneWeight, shardOneNodeCount, shardOneFlavor, shardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBaaSV2ClickhouseDatastoreExists(resourceDBaaSClickhouseDatastoreV2Name, &dbaasDatastore),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", datastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "region", dbaasRegion),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "status", string(dbaas_v2_common.DatastoreStatusActive)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "state", string(dbaas_v2_common.DatastoreStateRunning)),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.name", "shard1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.node_count", strconv.Itoa(shardOneNodeCount)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.weight", strconv.Itoa(shardOneWeight)),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.type", string(shardOneFlavor.Type)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.vcpus", strconv.Itoa(shardOneFlavor.VCPUs)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.ram", strconv.Itoa(shardOneFlavor.RAM)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.disk", strconv.Itoa(shardOneFlavor.Disk)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.disk_type", string(shardOneFlavor.DiskType)),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.#", "1"),
					resource.TestCheckResourceAttrSet(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.0"), // first item is not empty string

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "allow_reduce_nodes", strconv.FormatBool(allowReduceNodes)),
				),
			},
			// Update datastore name
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, datastorePassword, datastoreSG, keepersBlock, shardOneWeight, shardOneNodeCount, shardOneFlavor, shardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
				),
			},
			// Update datastore password
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, datastoreSG, keepersBlock, shardOneWeight, shardOneNodeCount, shardOneFlavor, shardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
				),
			},
			// Update datastore security groups
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, keepersBlock, shardOneWeight, shardOneNodeCount, shardOneFlavor, shardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.#", "1"),
					resource.TestCheckResourceAttrSet(resourceDBaaSClickhouseDatastoreV2Name, "security_groups.0"), // first item is not empty string
				),
			},
			// Update shard1 weight
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, keepersBlock, updatedShardOneWeight, shardOneNodeCount, shardOneFlavor, shardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.has_public_ips", "false"),
				),
			},
			// Update shard1 add public ips
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, keepersBlock, updatedShardOneWeight, shardOneNodeCount, shardOneFlavor, updatedShardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.has_public_ips", "true"),
				),
			},
			// Resize shard1 by flavor
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, keepersBlock, updatedShardOneWeight, shardOneNodeCount, updatedShardOneFlavor, updatedShardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.node_count", strconv.Itoa(shardOneNodeCount)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.type", string(updatedShardOneFlavor.Type)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.vcpus", strconv.Itoa(updatedShardOneFlavor.VCPUs)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.ram", strconv.Itoa(updatedShardOneFlavor.RAM)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.disk", strconv.Itoa(updatedShardOneFlavor.Disk)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.disk_type", string(updatedShardOneFlavor.DiskType)),
				),
			},
			// Add keepers node group
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, updatedKeepersBlock, updatedShardOneWeight, shardOneNodeCount, updatedShardOneFlavor, updatedShardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.name", "keepers"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.role", "KEEPER"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.name", "shard1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.node_count", strconv.Itoa(shardOneNodeCount)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.flavor.0.type", string(updatedShardOneFlavor.Type)),
				),
			},
			// Update node count for shard1 (add node)
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, updatedKeepersBlock, updatedShardOneWeight, updatedshardOneNodeCountTwo, updatedShardOneFlavor, updatedShardOneHasPublicIps, allowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.name", "keepers"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.role", "KEEPER"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.name", "shard1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.node_count", strconv.Itoa(updatedshardOneNodeCountTwo)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.flavor.0.type", string(updatedShardOneFlavor.Type)),
				),
			},
			// Update node count for shard1 (delete)
			{
				Config: testAccDBaaSClickhouseDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, updatedKeepersBlock, updatedShardOneWeight, updatedshardOneNodeCountOne, updatedShardOneFlavor, updatedShardOneHasPublicIps, updAllowReduceNodes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "name", updatedDatastoreName),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.name", "keepers"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.role", "KEEPER"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.name", "shard1"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.node_count", strconv.Itoa(updatedshardOneNodeCountOne)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.weight", strconv.Itoa(updatedShardOneWeight)),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "node_groups.1.flavor.0.type", string(updatedShardOneFlavor.Type)),

					resource.TestCheckResourceAttr(resourceDBaaSClickhouseDatastoreV2Name, "allow_reduce_nodes", strconv.FormatBool(updAllowReduceNodes)),
				),
			},
		},
	})
}

// testAccDBaaSClickhouseDatastoreV2Basic is a simple cluster with one shard.
func testAccDBaaSClickhouseDatastoreV2Basic(datastoreName, datastorePassword, datastoreSG, keepersBlock string, shardOneWeight, shardOneNodeCount int, shardOneFlavor dbaas_v2_ch.FlavorForNodeGroupRequest, shardOneHasPublicIps bool, allowReduceNodes bool) string {
	securityGroupsBlock := ""
	if datastoreSG != "" {
		securityGroupsBlock = fmt.Sprintf("security_groups = [\"%s\"]", datastoreSG)
	}

	HasPublickIPsBlock := ""
	if shardOneHasPublicIps {
		HasPublickIPsBlock = "has_public_ips = true"
	}
	return fmt.Sprintf(`
locals {
  project_id = "%s"
  region_name     = "%s"
}

provider openstack {
	tenant_id = local.project_id
}

// Need to check floating ips
data "openstack_networking_network_v2" "external_net" {
  external = true
  name = "external-network"
}

resource "openstack_networking_router_v2" "nat_router" {
  name                = "router_test"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external_net.id
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

resource "openstack_networking_router_interface_v2" "router_interface" {
  router_id = openstack_networking_router_v2.nat_router.id
  subnet_id = openstack_networking_subnet_v2.ds_subnet.id
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
  region = local.region_name
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
  password = "%s"
  %s // security_groups
  allow_reduce_nodes = %s

  // keepers
  %s

  node_groups {
    name       = "shard1" 
    role       = "DATA"
    node_count = "%d"
	weight     = "%d"
    %s // has_public_ips

    flavor {
      type  = "%s"
      vcpus = "%d"
      ram   = "%d"
      disk  = "%d"
      disk_type = "%s"
    }
  }
}`, dbaasProjectID, dbaasRegion, datastoreName, datastorePassword, securityGroupsBlock, strconv.FormatBool(allowReduceNodes), keepersBlock, shardOneNodeCount, shardOneWeight, HasPublickIPsBlock, shardOneFlavor.Type, shardOneFlavor.VCPUs, shardOneFlavor.RAM, shardOneFlavor.Disk, shardOneFlavor.DiskType)
}
