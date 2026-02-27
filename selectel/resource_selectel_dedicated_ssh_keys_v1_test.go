package selectel

import (
	"context"
	"errors"
	"fmt"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testAccDedicatedSSHKeyV1PublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q== example@example.org"

func TestAccDedicatedSSHKeysV1Basic(t *testing.T) {
	var sshKey dedicated.SSHKey

	keyName := acctest.RandomWithPrefix("tf-key")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDedicatedSSHKeysV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedSSHKeysV1Basic(keyName, testAccDedicatedSSHKeyV1PublicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedSSHKeysV1Exists(
						"selectel_dedicated_ssh_keys_v1.keypair_tf_acc_test_1",
						&sshKey,
					),
					resource.TestCheckResourceAttr(
						"selectel_dedicated_ssh_keys_v1.keypair_tf_acc_test_1",
						"name",
						keyName,
					),
					resource.TestCheckResourceAttr(
						"selectel_dedicated_ssh_keys_v1.keypair_tf_acc_test_1",
						"public_key",
						testAccDedicatedSSHKeyV1PublicKey,
					),
				),
			},
		},
	})
}

func testAccCheckDedicatedSSHKeysV1Exists(n string, sshKey *dedicated.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no SSH key ID is set")
		}

		cl := newTestDedicatedAPIClient(rs, testAccProvider)

		found, _, err := cl.GetSSHKey(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*sshKey = *found
		return nil
	}
}

func testAccCheckDedicatedSSHKeysV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_dedicated_ssh_keys_v1" {
			continue
		}

		cl := newTestDedicatedAPIClient(rs, testAccProvider)

		_, _, err := cl.GetSSHKey(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("dedicated ssh key %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccDedicatedSSHKeysV1Basic(name, publicKey string) string {
	return fmt.Sprintf(`
resource "selectel_dedicated_ssh_keys_v1" "keypair_tf_acc_test_1" {
  name       = "%s"
  public_key = "%s"
}
`, name, publicKey)
}
