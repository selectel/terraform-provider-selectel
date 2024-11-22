package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccSecretsManagerSecretV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	secretKey := acctest.RandomWithPrefix("tf-acc")
	secretValue := acctest.RandomWithPrefix("tf-acc")
	secretDescription := acctest.RandomWithPrefix("tf-acc")
	newSecretDescription := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsManagerSecretV1BasicConfig(projectName, secretKey, secretDescription, secretValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "key", secretKey),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "description", secretDescription),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "value", secretValue),
				),
			},
			{
				Config: testAccSecretsManagerSecretV1UpdateConfig(projectName, secretKey, newSecretDescription, secretValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "key", secretKey),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "description", newSecretDescription),
					resource.TestCheckResourceAttr("selectel_secretsmanager_secret_v1.secret_tf_acc_test_1", "value", secretValue),
				),
			},
		},
	})
}

func testAccSecretsManagerSecretV1BasicConfig(projectName, key, description, value string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name = "%s"
		}

		resource "selectel_secretsmanager_secret_v1" "secret_tf_acc_test_1" {
		     key = "%s"
		     description = "%s"
		     value = "%s"
		     project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		}
		`,
		projectName,
		key,
		description,
		value,
	)
}

func testAccSecretsManagerSecretV1UpdateConfig(projectName, key, description, value string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name = "%s"
		}

		resource "selectel_secretsmanager_secret_v1" "secret_tf_acc_test_1" {
		     key = "%s"
		     description = "%s"
		     value = "%s"
		     project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		}
		`,
		projectName,
		key,
		description,
		value,
	)
}
