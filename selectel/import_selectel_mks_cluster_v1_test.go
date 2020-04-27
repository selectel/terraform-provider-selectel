package selectel

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccMKSClusterV1ImportBasic(t *testing.T) {
	resourceName := "selectel_mks_cluster_v1.cluster_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := "1.16.8"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSClusterV1Basic(projectName, clusterName, kubeVersion),
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return fmt.Errorf("not found: %s", resourceName)
					}
					if rs.Primary.ID == "" {
						return errors.New("no ID is set")
					}

					var projectID, region string
					if v, ok := rs.Primary.Attributes["project_id"]; ok {
						projectID = v
					}
					if v, ok := rs.Primary.Attributes["region"]; ok {
						region = v
					}

					if err := os.Setenv("SEL_PROJECT_ID", projectID); err != nil {
						return fmt.Errorf("error setting SEL_PROJECT_ID: %s", err)
					}
					if err := os.Setenv("SEL_REGION", region); err != nil {
						return fmt.Errorf("error setting SEL_REGION: %s", err)
					}

					return nil
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
