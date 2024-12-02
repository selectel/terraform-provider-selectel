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
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1

	updatedDatastoreName := acctest.RandomWithPrefix("tf-acc-ds-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdateName(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdatePooler(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdateFirewall(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1Resize(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdateConfig(projectName, updatedDatastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", updatedDatastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(8192)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.mode", "session"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "pooler.0.size", strconv.Itoa(50)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(256)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(20)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func TestAccDBaaSDatastoreV1RedisBasic(t *testing.T) {
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
				Config: testAccDBaaSDatastoreV1RedisBasic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "volatile-lru"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdateRedisConfig(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1UpdateRedisPassword(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
			{
				Config: testAccDBaaSDatastoreV1RedisResize(projectName, datastoreName, resizeNodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(resizeNodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.maxmemory-policy", "noeviction"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "firewall.0.ips.#", "2"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
				),
			},
		},
	})
}

func TestAccDBaaSMultiNodeDatastoreV1Basic(t *testing.T) {
	var (
		dbaasDatastore dbaas.Datastore
		project        projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 3

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSDatastoreV1Exists("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", &dbaasDatastore),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "name", datastoreName),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "node_count", strconv.Itoa(nodeCount)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "enabled", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "status", string(dbaas.StatusActive)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.vcpus", strconv.Itoa(2)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(4096)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.xmloption", "content"),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.work_mem", strconv.Itoa(128)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.vacuum_cost_delay", strconv.Itoa(25)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "config.transform_null_equals", "true"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.master"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.MASTER"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.replica-1"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "connections.replica-2"),
				),
			},
		},
	})
}

func testAccCheckDBaaSDatastoreV1Exists(n string, dbaasDatastore *dbaas.Datastore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSClient(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		datastore, err := dbaasClient.Datastore(ctx, rs.Primary.ID)
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

func testAccDBaaSDatastoreV1Basic(projectName, datastoreName string, nodeCount int) string {
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
  config = {
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdateName(projectName, datastoreName string, nodeCount int) string {
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
  config = {
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdatePooler(projectName, datastoreName string, nodeCount int) string {
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
  config = {
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdateFirewall(projectName, datastoreName string, nodeCount int) string {
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
  config = {
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1Resize(projectName, datastoreName string, nodeCount int) string {
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
    ram = 8192
    disk = 32
  }
  config = {
    xmloption = "content"
    work_mem = 128
    vacuum_cost_delay = 25
    transform_null_equals = true
  }
  pooler {
    mode = "session"
    size = 50
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdateConfig(projectName, datastoreName string, nodeCount int) string {
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
    ram = 8192
    disk = 32
  }
  config = {
    xmloption = "content"
    work_mem = 256
    vacuum_cost_delay = 20
  }
  pooler {
    mode = "session"
    size = 50
  }
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1RedisBasic(projectName, datastoreName string, nodeCount int) string {
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

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
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
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdateRedisConfig(projectName, datastoreName string, nodeCount int) string {
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

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
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
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1UpdateRedisPassword(projectName, datastoreName string, nodeCount int) string {
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

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
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
  firewall {
    ips = [ "127.0.0.1", "127.0.0.2" ]
  }
  redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6123"
}`, projectName, datastoreName, nodeCount)
}

func testAccDBaaSDatastoreV1RedisResize(projectName, datastoreName string, nodeCount int) string {
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

  resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
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
	firewall {
	  ips = [ "127.0.0.1", "127.0.0.2" ]
	}
	redis_password = "quie7Hoh7ohTo[i0bae3Leeb4mai7ca6123"
  }`, projectName, datastoreName, nodeCount)
}
