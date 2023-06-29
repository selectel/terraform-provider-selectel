package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v1 "github.com/selectel/craas-go/pkg"
	"github.com/selectel/craas-go/pkg/v1/token"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/tokens"
)

func TestAccCRaaSTokenV1Basic(t *testing.T) {
	var (
		project    projects.Project
		craasToken token.Token
	)

	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCRaaSTokenV1Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckCRaaSTokenV1Exists("selectel_craas_token_v1.token_tf_acc_test_1", &craasToken),
					resource.TestCheckResourceAttr("selectel_craas_token_v1.token_tf_acc_test_1", "token_ttl", "1y"),
					resource.TestCheckResourceAttr("selectel_craas_token_v1.token_tf_acc_test_1", "username", "token"),
				),
			},
		},
	})
}

func testAccCheckCRaaSTokenV1Exists(n string, craasToken *token.Token) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		var projectID string
		if id, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = id
		}

		var tokenID string
		if t, ok := rs.Primary.Attributes["token"]; ok {
			tokenID = t
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		selTokenOpts := tokens.TokenOpts{
			ProjectID: projectID,
		}
		selToken, _, err := tokens.Create(ctx, resellV2Client, selTokenOpts)
		if err != nil {
			return errCreatingObject(objectToken, err)
		}

		craasClient := v1.NewCRaaSClientV1(selToken.ID, craasV1Endpoint)
		foundToken, _, err := token.Get(ctx, craasClient, tokenID)
		if err != nil {
			return err
		}

		if foundToken.Token != tokenID {
			return errors.New("token not found")
		}

		*craasToken = *foundToken

		return nil
	}
}

func testAccCRaaSTokenV1Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_craas_token_v1" "token_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
}
`, projectName)
}
