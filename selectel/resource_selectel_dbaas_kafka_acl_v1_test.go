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
	"github.com/selectel/go-selvpcclient/v3/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSKafkaACLV1Basic(t *testing.T) {
	var (
		dbaasACL dbaas.ACL
		project  projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	userName := RandomWithPrefix("tf_acc_user")
	userPassword := acctest.RandomWithPrefix("tf-acc-pass")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSKafkaACLV1Basic(projectName, datastoreName, userName, userPassword, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSKafkaACLV1Exists("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", &dbaasACL),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "pattern_type", "prefixed"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "pattern", "topic"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "allow_read", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "allow_write", "false"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
			{
				Config: testAccDBaaSKafkaACLV1Update(projectName, datastoreName, userName, userPassword, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSKafkaACLV1Exists("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", &dbaasACL),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "pattern_type", "prefixed"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "pattern", "topic"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "allow_read", "false"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "allow_write", "true"),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_acl_v1.acl_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
		},
	})
}

func testAccCheckDBaaSKafkaACLV1Exists(n string, dbaasACL *dbaas.ACL) resource.TestCheckFunc {
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

		acl, err := dbaasClient.ACL(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if acl.ID != rs.Primary.ID {
			return errors.New("acl not found")
		}

		*dbaasACL = acl

		return nil
	}
}

func testAccDBaaSKafkaACLV1Basic(projectName, datastoreName, userName, userPassword string, nodeCount int) string {
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
}

resource "selectel_dbaas_user_v1" "user_tf_acc_test_1" {
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	region = "ru-3"
	datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
	name = "%s"
	password = "%s"
}

resource "selectel_dbaas_kafka_acl_v1" "acl_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
  user_id = "${selectel_dbaas_user_v1.user_tf_acc_test_1.id}"
  pattern = "topic"
  pattern_type = "prefixed"
  allow_read = true
  allow_write = false
}`, projectName, datastoreName, nodeCount, userName, userPassword)
}

func testAccDBaaSKafkaACLV1Update(projectName, datastoreName, userName, userPassword string, nodeCount int) string {
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
}

resource "selectel_dbaas_user_v1" "user_tf_acc_test_1" {
	project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	region = "ru-3"
	datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
	name = "%s"
	password = "%s"
}

resource "selectel_dbaas_kafka_acl_v1" "acl_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
  user_id = "${selectel_dbaas_user_v1.user_tf_acc_test_1.id}"
  pattern = "topic"
  pattern_type = "prefixed"
  allow_read = false
  allow_write = true
}`, projectName, datastoreName, nodeCount, userName, userPassword)
}
