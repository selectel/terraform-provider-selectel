package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/secretsmanager-go/secretsmanagererrors"
)

func TestAccSecretsManagerSecretV1ImportBasic(t *testing.T) {
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resourceName := "selectel_secretsmanager_secret_v1.secret_tf_acc_test_1"

	secretKey := acctest.RandomWithPrefix("tf-acc")
	secretValue := acctest.RandomWithPrefix("tf-acc")
	secretDescription := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSecretsManagerV1SecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsManagerSecretV1WithoutProjectBasic(secretKey, secretDescription, secretValue, projectID),
				Check:  testAccCheckSelectelImportEnv(resourceName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func testAccSecretsManagerSecretV1WithoutProjectBasic(secretKey, secretDescription, secretValue, projectID string) string {
	return fmt.Sprintf(`
		resource "selectel_secretsmanager_secret_v1" "secret_tf_acc_test_1" {
				key = "%s"
				description = "%s"
				value = "%s"
				project_id = "%s"
		}`,
		secretKey,
		secretDescription,
		secretValue,
		projectID,
	)
}

func testAccCheckSecretsManagerV1SecretDestroy(s *terraform.State) error {
	smImportClient, diagErr := getSecretsManagerClientForAccImportTests(testAccProvider.Meta())
	if diagErr != nil {
		return fmt.Errorf("can't get getSecretsManagerClientForAccImportTests for secret import test")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_secretsmanager_secret_v1" {
			continue
		}

		_, key, _ := resourceSecretsManagerSecretV1ParseID(rs.Primary.ID)
		_, err := smImportClient.Secrets.Get(context.Background(), key)
		if !errors.Is(err, secretsmanagererrors.ErrBadRequestStatusText) {
			return errors.New("secret still exists")
		}
	}

	return nil
}
