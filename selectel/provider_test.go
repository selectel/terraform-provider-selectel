package selectel

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccProviders map[string]func() (*schema.Provider, error)
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"selectel": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccSelectelPreCheck(t *testing.T) {
	if v := os.Getenv("OS_DOMAIN_NAME"); v == "" {
		t.Fatal("OS_DOMAIN_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("OS_USERNAME"); v == "" {
		t.Fatal("OS_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("OS_PASSWORD"); v == "" {
		t.Fatal("OS_PASSWORD must be set for acceptance tests")
	}
}

func testAccSelectelPreCheckWithProjectID(t *testing.T) {
	testAccSelectelPreCheck(t)
	if v := os.Getenv("INFRA_PROJECT_ID"); v == "" {
		t.Fatal("INFRA_PROJECT_ID must be set for acceptance tests")
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

		if err := os.Setenv("INFRA_PROJECT_ID", projectID); err != nil {
			return fmt.Errorf("error setting INFRA_PROJECT_ID: %s", err)
		}
		if err := os.Setenv("INFRA_REGION", region); err != nil {
			return fmt.Errorf("error setting INFRA_REGION: %s", err)
		}

		return nil
	}
}

func testAccCheckSelectelCRaaSImportEnv(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		var projectID string

		if v, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = v
		}

		if err := os.Setenv("INFRA_PROJECT_ID", projectID); err != nil {
			return fmt.Errorf("error setting INFRA_PROJECT_ID: %s", err)
		}

		return nil
	}
}
