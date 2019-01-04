package selvpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/keypairs"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/users"
	"github.com/stretchr/testify/assert"
)

func TestAccResellV2KeypairBasic(t *testing.T) {
	var (
		user    users.User
		keypair keypairs.Keypair
	)
	keypairName := acctest.RandomWithPrefix("tf-acc")
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q== example@example.org"
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2KeypairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2KeypairBasic(userName, userPassword, keypairName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2UserExists("selvpc_resell_user_v2.user_tf_acc_test_1", &user),
					testAccCheckResellV2KeypairExists("selvpc_resell_keypair_v2.keypair_tf_acc_test_1", &keypair),
					resource.TestCheckResourceAttr("selvpc_resell_keypair_v2.keypair_tf_acc_test_1", "public_key", publicKey),
					resource.TestCheckResourceAttr("selvpc_resell_keypair_v2.keypair_tf_acc_test_1", "regions.#", "2"),
					resource.TestCheckResourceAttrSet("selvpc_resell_keypair_v2.keypair_tf_acc_test_1", "user_id"),
				),
			},
		},
	})
}

func testAccCheckResellV2KeypairDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_keypair_v2" {
			continue
		}

		userID, keypairName, err := resourceResellKeypairV2ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		existingKeypairs, _, err := keypairs.List(ctx, resellV2Client)
		if err != nil {
			return errSearchingKeypair(keypairName, err)
		}

		found := false
		for _, keypair := range existingKeypairs {
			if keypair.UserID == userID && keypair.Name == keypairName {
				found = true
			}
		}

		if found {
			return fmt.Errorf("keypair '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckResellV2KeypairExists(n string, keypair *keypairs.Keypair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		userID, keypairName, err := resourceResellKeypairV2ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		existingKeypairs, _, err := keypairs.List(ctx, resellV2Client)
		if err != nil {
			return errSearchingKeypair(keypairName, err)
		}

		found := false
		foundKeypairIdx := 0
		for i, keypair := range existingKeypairs {
			if keypair.UserID == userID && keypair.Name == keypairName {
				found = true
				foundKeypairIdx = i
			}
		}

		if !found {
			return errors.New("keypair not found")
		}

		*keypair = *existingKeypairs[foundKeypairIdx]

		return nil
	}
}

func testAccResellV2KeypairBasic(userName, userPassword, keypairName, publicKey string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_user_v2" "user_tf_acc_test_1" {
  name        = "%s"
  password    = "%s"
}

resource "selvpc_resell_keypair_v2" "keypair_tf_acc_test_1" {
  name       = "%s"
  public_key = "%s"
  regions    = ["ru-1", "ru-3"]
  user_id    = "${selvpc_resell_user_v2.user_tf_acc_test_1.id}"
}`, userName, userPassword, keypairName, publicKey)
}

func TestResourceResellKeypairV2BuildID(t *testing.T) {
	expected := "db9e1958679a4d8cbd7561e8f060aa15/key1"

	actual := resourceResellKeypairV2BuildID(
		"db9e1958679a4d8cbd7561e8f060aa15",
		"key1",
	)

	assert.Equal(t, expected, actual)
}

func TestResourceResellKeypairV2ParseID(t *testing.T) {
	expectedUserID := "db9e1958679a4d8cbd7561e8f060aa15"
	expectedKeypairName := "key1"

	actualUserID, actualUserName, err := resourceResellKeypairV2ParseID(
		"db9e1958679a4d8cbd7561e8f060aa15/key1",
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, actualUserID)
	assert.Equal(t, expectedKeypairName, actualUserName)
}
