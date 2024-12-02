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

func TestAccDBaaSKafkaTopicV1Basic(t *testing.T) {
	var (
		dbaasTopic dbaas.Topic
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	topicName := RandomWithPrefix("tf_acc_topic")
	topicPartitions := 1
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSKafkaTopicV1Basic(projectName, datastoreName, topicName, strconv.Itoa(topicPartitions), nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSKafkaTopicV1Exists("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", &dbaasTopic),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "name", topicName),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "partitions", strconv.Itoa(topicPartitions)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
			{
				Config: testAccDBaaSKafkaTopicV1Update(projectName, datastoreName, topicName, strconv.Itoa(topicPartitions+1), nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSKafkaTopicV1Exists("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", &dbaasTopic),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "name", topicName),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "partitions", strconv.Itoa(topicPartitions+1)),
					resource.TestCheckResourceAttr("selectel_dbaas_kafka_topic_v1.topic_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
		},
	})
}

func testAccCheckDBaaSKafkaTopicV1Exists(n string, dbaasTopic *dbaas.Topic) resource.TestCheckFunc {
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

		topic, err := dbaasClient.Topic(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if topic.ID != rs.Primary.ID {
			return errors.New("topic not found")
		}

		*dbaasTopic = topic

		return nil
	}
}

func testAccDBaaSKafkaTopicV1Basic(projectName, datastoreName, topicName, topicPartitions string, nodeCount int) string {
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

resource "selectel_dbaas_kafka_topic_v1" "topic_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  partitions = "%s"
}`, projectName, datastoreName, nodeCount, topicName, topicPartitions)
}

func testAccDBaaSKafkaTopicV1Update(projectName, datastoreName, topicName, topicPartitions string, nodeCount int) string {
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

resource "selectel_dbaas_kafka_topic_v1" "topic_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_kafka_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  partitions = "%s"
}`, projectName, datastoreName, nodeCount, topicName, topicPartitions)
}
