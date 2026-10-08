package selectel

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	dbaas_v2_os "github.com/selectel/dbaas-go/v2/opensearch"
)

const resourceDBaaSOpensearchDatastoreV2Name = "selectel_dbaas_opensearch_datastore_v2.datastore_tf_acc_test_1"

func testAccCheckDBaaSV2OpensearchDatastoreDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_dbaas_opensearch_datastore_v2" {
			continue
		}
		ctx := context.Background()
		client, err := newTestDBaaSV2Client(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.Opensearch.GetDatastore(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(
				"opensearch datastore %s still exists after destroy",
				rs.Primary.ID,
			)
		}
	}

	return nil
}

func testAccCheckDBaaSV2OpensearchDatastoreExists(n string, dbaasDatastore *dbaas_v2_os.DatastoreResponse) resource.TestCheckFunc {
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

		datastore, err := dbaasClient.Opensearch.GetDatastore(ctx, rs.Primary.ID)
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

func TestAccDBaaSOpensearchDatastoreV2Basic(t *testing.T) {
	var dbaasDatastore dbaas_v2_os.DatastoreResponse
	var dataOneNodeGroupID string

	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	datastorePassword := "Iu2YgYlk!ORz" //nolint:gosec // used in acceptance tests only.
	datastoreSG := ""
	dataOneName := "data1"
	dataOneNameRenamed := "data1-renamed"
	dataOneNodeCount := 1
	dataOneFlavor := dbaas_v2_os.FlavorForNodeGroupRequest{
		Type:     dbaas_v2_common.FlavorTypeFlexible,
		VCPUs:    2,
		RAM:      8192,
		Disk:     32,
		DiskType: dbaas_v2_common.FlavorDiskLocal,
	}
	dataOneHasPublicIps := false
	managersBlock := `
	node_group {
	  key        = "managers"
	  name       = "managers"
	  role       = "MANAGER"
	  node_count = 3

	  flavor {
	    id    = "${data.selectel_dbaas_flavor_v2.manager_flavor.flavors[0].id}"
	    type  = "FIXED"
	  }
	}
	`
	dashboardBlock := ""
	updatedDatastoreName := acctest.RandomWithPrefix("tf-acc-ds-updated")
	updatedDatastorePassword := "Iu2YgYlk!ORzUpd" //nolint:gosec // used in acceptance tests only.

	updatedDataOneNodeCountTwo := 2
	updatedDataOneFlavor := dbaas_v2_os.FlavorForNodeGroupRequest{
		Type:     dbaas_v2_common.FlavorTypeFlexible,
		VCPUs:    4,
		RAM:      8192,
		Disk:     35,
		DiskType: dbaas_v2_common.FlavorDiskLocal,
	}
	updatedDatastoreSG := "${openstack_networking_secgroup_v2.ds_sg.id}"
	updatedDataOneHasPublicIps := true
	updatedDashboardBlock := `
	node_group {
	  key        = "dashboard"
	  name       = "dashboard"
	  role       = "DASHBOARD"
	  node_count = 1

	  flavor {
	    type  = "FLEXIBLE"
		vcpus     = 2
		ram       = 4096
		disk      = 64
		disk_type = "LOCAL"
	  }
	}
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			// Need to create a network by openstack operator beacause 'selectel_vpc_subnet_v2' creates a network with 'public' tag.
			// Opensearch api denies 'You can not use subnet {subnet_id} because it is a part of the external network'.
			testAccDBaaSV2PreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckDBaaSV2OpensearchDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(datastoreName, datastorePassword, datastoreSG, managersBlock, dataOneName, dataOneNodeCount, dataOneFlavor, dataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBaaSV2OpensearchDatastoreExists(resourceDBaaSOpensearchDatastoreV2Name, &dbaasDatastore),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", datastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "region", dbaasRegion),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "status", string(dbaas_v2_common.DatastoreStatusActive)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "state", string(dbaas_v2_common.DatastoreStateRunning)),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.#", "2"),
					// The 'node_group' attribute is a TypeList, elements are stored
					// in the configuration order: managers, data1.
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.key", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.name", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.role", "MANAGER"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.name", dataOneName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.node_count", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.type", string(dataOneFlavor.Type)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.vcpus", strconv.Itoa(dataOneFlavor.VCPUs)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.ram", strconv.Itoa(dataOneFlavor.RAM)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.disk", strconv.Itoa(dataOneFlavor.Disk)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.disk_type", string(dataOneFlavor.DiskType)),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "security_groups.#", "0"),
				),
			},
			// Update datastore name
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, datastorePassword, datastoreSG, managersBlock, dataOneName, dataOneNodeCount, dataOneFlavor, dataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),
				),
			},
			// Update datastore password
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, datastoreSG, managersBlock, dataOneName, dataOneNodeCount, dataOneFlavor, dataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),
				),
			},
			// Update datastore security groups
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneName, dataOneNodeCount, dataOneFlavor, dataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "security_groups.#", "1"),
					// The 'security_groups' attribute is a TypeSet, so items are addressed
					// by hash instead of index. Check that the security group resource ID
					// is present in the set.
					resource.TestCheckTypeSetElemAttrPair(resourceDBaaSOpensearchDatastoreV2Name, "security_groups.*", "openstack_networking_secgroup_v2.ds_sg", "id"),
				),
			},
			// Update data1 add public ips
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneName, dataOneNodeCount, dataOneFlavor, updatedDataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.has_public_ips", "true"),
				),
			},
			// Resize data1 by flavor
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneName, dataOneNodeCount, updatedDataOneFlavor, updatedDataOneHasPublicIps, dashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.node_count", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.type", string(updatedDataOneFlavor.Type)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.vcpus", strconv.Itoa(updatedDataOneFlavor.VCPUs)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.ram", strconv.Itoa(updatedDataOneFlavor.RAM)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.disk", strconv.Itoa(updatedDataOneFlavor.Disk)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.disk_type", string(updatedDataOneFlavor.DiskType)),
				),
			},
			// Add dashboard node group
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneName, dataOneNodeCount, updatedDataOneFlavor, updatedDataOneHasPublicIps, updatedDashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.#", "3"),
					// Configuration order: managers, data1, dashboard.
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.key", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.name", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.role", "MANAGER"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.name", dataOneName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.node_count", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.type", string(updatedDataOneFlavor.Type)),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.key", "dashboard"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.name", "dashboard"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.role", "DASHBOARD"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.node_count", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.flavor.0.type", "FLEXIBLE"),
				),
			},
			// Update node count for data1 (add node)
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneName, updatedDataOneNodeCountTwo, updatedDataOneFlavor, updatedDataOneHasPublicIps, updatedDashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBaaSV2OpensearchDatastoreExists(resourceDBaaSOpensearchDatastoreV2Name, &dbaasDatastore),
					// Capture the current data1 node group ID to verify
					// later that renaming is in-place and does not recreate
					// the node group.
					func(_ *terraform.State) error {
						for _, nodeGroup := range dbaasDatastore.NodeGroups {
							if nodeGroup.Name != dataOneName {
								continue
							}
							dataOneNodeGroupID = nodeGroup.ID

							return nil
						}

						return fmt.Errorf("node group with name %q not found in datastore", dataOneName)
					},

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "name", updatedDatastoreName),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.#", "3"),
					// Configuration order: managers, data1, dashboard.
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.key", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.name", "managers"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.role", "MANAGER"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.node_count", "3"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.0.flavor.0.type", "FIXED"),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.name", dataOneName),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.node_count", "2"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.type", string(updatedDataOneFlavor.Type)),

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.key", "dashboard"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.name", "dashboard"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.role", "DASHBOARD"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.node_count", "1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.2.flavor.0.type", "FLEXIBLE"),
				),
			},
			// Rename data1 node group in-place, its key is not changed
			{
				Config: testAccDBaaSOpensearchDatastoreV2Basic(updatedDatastoreName, updatedDatastorePassword, updatedDatastoreSG, managersBlock, dataOneNameRenamed, updatedDataOneNodeCountTwo, updatedDataOneFlavor, updatedDataOneHasPublicIps, updatedDashboardBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBaaSV2OpensearchDatastoreExists(resourceDBaaSOpensearchDatastoreV2Name, &dbaasDatastore),
					// The renamed node group must keep its ID.
					func(_ *terraform.State) error {
						for _, nodeGroup := range dbaasDatastore.NodeGroups {
							if nodeGroup.ID != dataOneNodeGroupID {
								continue
							}
							if nodeGroup.Name != dataOneNameRenamed {
								return fmt.Errorf("node group %s has name %q, expected %q",
									nodeGroup.ID, nodeGroup.Name, dataOneNameRenamed)
							}

							return nil
						}

						return fmt.Errorf("node group with ID %s not found after rename, it was probably recreated",
							dataOneNodeGroupID)
					},

					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.#", "3"),
					// The renamed group stays at index 1: in-place rename does
					// not change the key or the position.
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.key", "data1"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.name", dataOneNameRenamed),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.role", "DATA"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.node_count", strconv.Itoa(updatedDataOneNodeCountTwo)),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.has_public_ips", "true"),
					resource.TestCheckResourceAttr(resourceDBaaSOpensearchDatastoreV2Name, "node_group.1.flavor.0.type", string(updatedDataOneFlavor.Type)),
				),
			},
		},
	})
}

// testAccDBaaSOpensearchDatastoreV2Basic is a simple cluster with one data node group and manager node group.
func testAccDBaaSOpensearchDatastoreV2Basic(datastoreName, datastorePassword, datastoreSG, managersBlock, dataOneName string, dataOneNodeCount int, dataOneFlavor dbaas_v2_os.FlavorForNodeGroupRequest, dataOneHasPublicIps bool, dashboardBlock string) string {
	securityGroupsBlock := ""
	if datastoreSG != "" {
		securityGroupsBlock = fmt.Sprintf("security_groups = [\"%s\"]", datastoreSG)
	}

	HasPublicIPsBlock := ""
	if dataOneHasPublicIps {
		HasPublicIPsBlock = "has_public_ips = true"
	}

	return fmt.Sprintf(`
locals {
  project_id = "%s"
  region_name = "%s"
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
    engine = "opensearch"
    version = "2.18"

  }
}

data "selectel_dbaas_flavor_v2" "manager_flavor" {
  project_id = local.project_id
  region = local.region_name
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
	allowed_role = "MANAGER"
  }
}

data "selectel_dbaas_flavor_v2" "dashboard_flavor" {
  project_id = local.project_id
  region = local.region_name
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
	allowed_role = "DASHBOARD"
  }
}

resource "selectel_dbaas_opensearch_datastore_v2" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = local.project_id
  region = local.region_name
  type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
  subnet_id = "${openstack_networking_subnet_v2.ds_subnet.id}"
  password = "%s"
  %s // security_groups

  // managers
  %s

  node_group {
    key        = "data1"
    name       = "%s"
    role       = "DATA"
    node_count = "%d"
    %s // has_public_ips

    flavor {
      type  = "%s"
      vcpus = "%d"
      ram   = "%d"
      disk  = "%d"
      disk_type = "%s"
    }
  }
  
  // dashboard
  %s
}`, dbaasProjectID, dbaasRegion, datastoreName, datastorePassword, securityGroupsBlock, managersBlock, dataOneName, dataOneNodeCount, HasPublicIPsBlock, dataOneFlavor.Type, dataOneFlavor.VCPUs, dataOneFlavor.RAM, dataOneFlavor.Disk, dataOneFlavor.DiskType, dashboardBlock)
}
