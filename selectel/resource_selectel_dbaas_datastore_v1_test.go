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
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

func TestAccDBaaSDatastoreV1Basic(t *testing.T) {
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
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.ram", strconv.Itoa(2048)),
					resource.TestCheckResourceAttr("selectel_dbaas_datastore_v1.datastore_tf_acc_test_1", "flavor.0.disk", strconv.Itoa(32)),
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

		var projectID, endpoint string
		if id, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = id
		}
		if region, ok := rs.Primary.Attributes["region"]; ok {
			endpoint = getDBaaSV1Endpoint(region)
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		tokenOpts := tokens.TokenOpts{
			ProjectID: projectID,
		}
		token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
		if err != nil {
			return errCreatingObject(objectToken, err)
		}

		dbaasClient, err := dbaas.NewDBAASClient(token.ID, endpoint)
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
  auto_quotas = true
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_types_v1" "dt" {
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
  type_id = "${data.selectel_dbaas_datastore_types_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 4096
    disk = 32
  }
}`, projectName, datastoreName, nodeCount)
}
