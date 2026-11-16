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
			var steps []resource.TestStep
			for _, step := range []struct {
				config  string
				pin     int
				version string
			}{
				{original, 0, "1"},
				{rotated, 0, "2"},
				{rotated, 1, "1"},
				{original, 1, "1"},
				{original, 3, "3"},
				{original, 0, "3"},
			} {
				config := strings.Replace(step.config, "project_id =", fmt.Sprintf("current_version = %d\n project_id =", step.pin), 1)
				steps = append(steps, resource.TestStep{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(address, "version", step.version),
						resource.TestCheckResourceAttr(address, "current_version", fmt.Sprint(step.pin)),
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
