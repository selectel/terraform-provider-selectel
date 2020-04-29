package selectel

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"selectel": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(_ *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccSelectelPreCheck(t *testing.T) {
	if v := os.Getenv("SEL_TOKEN"); v == "" {
		t.Fatal("SEL_TOKEN must be set for acceptance tests")
	}
}

func testAccCheckSelectelImportEnv(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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
	}
}
