package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-provider-openstack/terraform-provider-openstack/openstack"
)

var (
	testAccProviders              map[string]func() (*schema.Provider, error)
	testAccProvidersWithOpenStack map[string]func() (*schema.Provider, error)
	testAccProvider               *schema.Provider
	// testAccProtoV6ProviderFactories serves the provider through the same mux as
	// main.go, plus the selectel_mux_test data source.
	testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
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
	testAccProvider = testAccCapTimeouts(Provider("test"))
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
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"selectel": func() (tfprotov6.ProviderServer, error) {
			sdk := testAccCapTimeouts(Provider("test"))
			server, err := muxProviderServer(context.Background(), sdk,
				testFrameworkProvider{&frameworkProvider{sdk: sdk, version: "test"}})
			if err != nil {
				return nil, err
			}

			return server(), nil
		},
	}
}

// testAccCapTimeouts sets the operation timeouts of the SDKv2 resources to at
// most SELECTEL_ACC_OPERATION_TIMEOUT, for example "30m". Go cannot stop one
// test of a running package, so the cap is what ends a test that waits for a
// stuck object: the operation fails after the cap instead of the resource
// default, which can be an hour, and the next tests run. Resources without
// their own timeouts keep the SDK default of 20 minutes.
func testAccCapTimeouts(p *schema.Provider) *schema.Provider {
	v := os.Getenv("SELECTEL_ACC_OPERATION_TIMEOUT")
	if v == "" {
		return p
	}
	limit, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("SELECTEL_ACC_OPERATION_TIMEOUT: %s", err))
	}

	capped := func(d *time.Duration) *time.Duration {
		if d == nil || *d <= limit {
			return d
		}

		return &limit
	}
	for _, r := range p.ResourcesMap {
		if t := r.Timeouts; t != nil {
			t.Create, t.Read, t.Update = capped(t.Create), capped(t.Read), capped(t.Update)
			t.Delete, t.Default = capped(t.Delete), capped(t.Default)
		}
	}

	return p
}

func TestProvider(t *testing.T) {
	if err := Provider("test").InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderVPCRegistration(t *testing.T) {
	provider := Provider("test")

	for _, name := range []string{
		"selectel_vpc_floatingip_v2",
		"selectel_vpc_network_v2",
		"selectel_vpc_network_subnet_v2",
		"selectel_vpc_network_subnet_route_v2",
		"selectel_vpc_port_v2",
		"selectel_vpc_port_forwarding_v2",
		"selectel_vpc_port_secgroup_associate_v2",
		"selectel_vpc_floatingip_associate_v2",
		"selectel_vpc_router_v2",
		"selectel_vpc_router_interface_v2",
		"selectel_vpc_router_route_v2",
		"selectel_vpc_router_routes_v2",
		"selectel_vpc_rbac_policy_v2",
		"selectel_vpc_secgroup_v2",
		"selectel_vpc_secgroup_rule_v2",
		"selectel_vpc_subnet_v2",
		"selectel_vpc_public_subnet_v2",
	} {
		if _, ok := provider.ResourcesMap[name]; !ok {
			t.Errorf("expected managed resource %s to be registered", name)
		}
	}

	deprecatedSubnet := provider.ResourcesMap["selectel_vpc_subnet_v2"]
	if deprecatedSubnet.DeprecationMessage == "" {
		t.Error("expected selectel_vpc_subnet_v2 to be marked as deprecated")
	}

	for _, name := range []string{
		"selectel_vpc_network_v2",
		"selectel_vpc_network_subnet_v2",
		"selectel_vpc_port_v2",
		"selectel_vpc_router_v2",
		"selectel_vpc_secgroup_v2",
	} {
		if _, ok := provider.DataSourcesMap[name]; !ok {
			t.Errorf("expected data source %s to be registered", name)
		}
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

func testAccSelectelPreCheckWithProjectIDAndRegion(t *testing.T) {
	testAccSelectelPreCheckWithProjectID(t)
	if v := os.Getenv("INFRA_REGION"); v == "" {
		t.Fatal("INFRA_REGION must be set for acceptance tests")
	}
}

func testAccSelectelPreCheckWithAuth(t *testing.T) {
	testAccSelectelPreCheck(t)

	if v := os.Getenv("OS_AUTH_URL"); v == "" {
		t.Fatal("OS_AUTH_URL must be set for acceptance tests")
	}

	if v := os.Getenv("OS_REGION_NAME"); v == "" {
		t.Fatal("OS_REGION_NAME must be set for acceptance tests")
	}
}

// testAccCheckSelectelImportEnv points INFRA_PROJECT_ID and INFRA_REGION at the
// project and region of the resource, so the next import step finds it: the
// importers take them from the provider configuration. t.Setenv restores the
// variables when the test ends, so later tests keep their own values.
func testAccCheckSelectelImportEnv(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		if v, ok := rs.Primary.Attributes["project_id"]; ok {
			t.Setenv("INFRA_PROJECT_ID", v)
		}
		if v, ok := rs.Primary.Attributes["region"]; ok {
			t.Setenv("INFRA_REGION", v)
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

// testAccSkipUnlessEnv skips a test that an ordinary account cannot run,
// unless the variable is set. The reason tells what the account needs.
func testAccSkipUnlessEnv(t *testing.T, name, reason string) {
	t.Helper()
	if os.Getenv(name) == "" {
		t.Skipf("%s must be set to run this test: %s", name, reason)
	}
}

func testAccDedicatedServersPreCheck(t *testing.T) {
	t.Helper()
	testAccSkipUnlessEnv(t, "SELECTEL_ACC_DEDICATED_SERVERS",
		"the test orders a real dedicated server or looks for a server created by hand")
}

func testAccDedicatedNetworkPreCheck(t *testing.T) {
	t.Helper()
	testAccSkipUnlessEnv(t, "SELECTEL_ACC_DEDICATED_NETWORK",
		"the test needs VLAN 2989 of a dedicated network in SPB-5")
}

func testAccDomainsV1PreCheck(t *testing.T) {
	t.Helper()
	testAccSkipUnlessEnv(t, "SELECTEL_ACC_DOMAINS_V1",
		"Domains API v1 answers 403 to accounts without access to it")
}

func testAccIAMUsersPreCheck(t *testing.T) {
	t.Helper()
	testAccSkipUnlessEnv(t, "SELECTEL_ACC_IAM_USERS",
		"the test creates users with example.com emails, and some accounts accept internal emails only")
}

func testAccMKSMultiAZPreCheck(t *testing.T) {
	t.Helper()
	testAccSkipUnlessEnv(t, "SELECTEL_ACC_MKS_MULTI_AZ",
		"the project needs the mks_cluster_high_availability_multi_az quota")
}

func testAccDBaaSV2PreCheck(t *testing.T) {
	testAccSelectelPreCheckWithProjectID(t)
	if v := os.Getenv("INFRA_REGION"); v == "" {
		t.Fatal("INFRA_REGION must be set for acceptance tests")
	}
}
