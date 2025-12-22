package selectel

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-provider-openstack/terraform-provider-openstack/openstack"
)

var (
	testAccProviders              map[string]func() (*schema.Provider, error)
	testAccProvidersWithOpenStack map[string]func() (*schema.Provider, error)
	testAccProvider               *schema.Provider
	// global router TestAcc env variables.
	globalRouterDedicatedNetworkVLAN  = os.Getenv("GLOBAL_ROUTER_DEICATED_NETWORK_VLAN")
	globalRouterDedicatedRegion       = os.Getenv("GLOBAL_ROUTER_DEDICATED_REGION")
	globalRouterVPCRegion             = os.Getenv("GLOBAL_ROUTER_CLOUD_REGION")
	globalRouterVPCProjectID          = os.Getenv("GLOBAL_ROUTER_CLOUD_PROJECT_ID")
	globalRouterSubnetCidr            = os.Getenv("GLOBAL_ROUTER_SUBNET_CIDR")
	globalRouterSubnetGateway         = os.Getenv("GLOBAL_ROUTER_SUBNET_GATEWAY")
	globalRouterSubnetServiceAddress1 = os.Getenv("GLOBAL_ROUTER_SUBNET_SERVICE_ADDR1")
	globalRouterSubnetServiceAddress2 = os.Getenv("GLOBAL_ROUTER_SUBNET_SERVICE_ADDR2")
	globalRouterStaticRouteCidr       = os.Getenv("GLOBAL_ROUTER_STATIC_ROUTE_CIDR")
	globalRouterNextHop               = os.Getenv("GLOBAL_ROUTER_STATIC_ROUTE_NEXT_HOP")
)

func init() {
	testAccProvider = Provider("test")
	testAccProviders = map[string]func() (*schema.Provider, error){
		"selectel": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
	testAccProvidersWithOpenStack = map[string]func() (*schema.Provider, error){
		"selectel": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
		"openstack": func() (*schema.Provider, error) {
			return openstack.Provider(), nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider("test").InternalValidate(); err != nil {
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

func testAccGlobalRouterDedicatedNetworkPreCheck(t *testing.T) {
	if globalRouterDedicatedRegion == "" {
		t.Skip("GLOBAL_ROUTER_DEDICATED_REGION must be set for acceptance tests of Global Router Dedicated network")
	}
	if globalRouterDedicatedNetworkVLAN == "" {
		t.Skip("GLOBAL_ROUTER_DEICATED_NETWORK_VLAN must be set for acceptance tests of Global Router Dedicated network")
	}
}

func testAccGlobalRouterVPCNetworkPreCheck(t *testing.T) {
	if globalRouterVPCRegion == "" {
		t.Skip("GLOBAL_ROUTER_CLOUD_REGION must be set for acceptance tests of Global Router VPC network")
	}
	if globalRouterVPCProjectID == "" {
		t.Skip("GLOBAL_ROUTER_CLOUD_PROJECT_ID must be set for acceptance tests of Global Router VPC network")
	}
}

func testAccGlobalRouterSubnetPreCheck(t *testing.T) {
	if globalRouterSubnetCidr == "" {
		t.Skip("GLOBAL_ROUTER_SUBNET_CIDR must be set for acceptance tests of Global Router subnet")
	}
	if globalRouterSubnetGateway == "" {
		t.Skip("GLOBAL_ROUTER_SUBNET_GATEWAY must be set for acceptance tests of Global Router subnet")
	}
	if globalRouterSubnetServiceAddress1 == "" {
		t.Skip("GLOBAL_ROUTER_SUBNET_SERVICE_ADDR1 must be set for acceptance tests of Global Router subnet")
	}
	if globalRouterSubnetServiceAddress2 == "" {
		t.Skip("GLOBAL_ROUTER_SUBNET_SERVICE_ADDR2 must be set for acceptance tests of Global Router subnet")
	}
}

func testAccGlobalRouterStaticRoutePreCheck(t *testing.T) {
	if globalRouterStaticRouteCidr == "" {
		t.Skip("GLOBAL_ROUTER_STATIC_ROUTE_CIDR must be set for acceptance tests of Global Router static router in VPC subnet")
	}
	if globalRouterNextHop == "" {
		t.Skip("GLOBAL_ROUTER_STATIC_ROUTE_NEXT_HOP must be set for acceptance tests of Global Router static router in VPC subnet")
	}
}
