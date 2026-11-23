package selectel

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSecretsManagerVersioning(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			projectName := acctest.RandomWithPrefix("tf-acc")
			name := acctest.RandomWithPrefix("tf-acc")
			var original, rotated, address string
			if kind == "certificate" {
				original = testAccSecretsManagerCertificateV1BasicConfig(projectName, name)
				rotated = testAccSecretsManagerCertificateV1UpdateConfig(projectName, name)
				address = "selectel_secretsmanager_certificate_v1.certificate_tf_acc_test_1"
			} else {
				original = testAccSecretsManagerSecretV1BasicConfig(projectName, name, "versioning", "original-value")
				rotated = testAccSecretsManagerSecretV1BasicConfig(projectName, name, "versioning", "rotated-value")
				address = "selectel_secretsmanager_secret_v1.secret_tf_acc_test_1"
			}
			steps := make([]resource.TestStep, 0, 4)
			for _, step := range []struct {
				config   string
				activate bool
				version  string
			}{
				{original, false, "1"},
				{rotated, false, "1"},
				{rotated, true, "1"},
				{original, true, "3"},
			} {
				config := strings.Replace(step.config, "project_id =", fmt.Sprintf("is_current = %t\n project_id =", step.activate), 1)
				steps = append(steps, resource.TestStep{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(address, "version", step.version),
						resource.TestCheckResourceAttr(address, "is_current", fmt.Sprint(step.activate)),
					),
				})
			}
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccSelectelPreCheck(t) },
				ProviderFactories: testAccProviders,
				CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
				Steps:             steps,
			})
		})
	}
}
