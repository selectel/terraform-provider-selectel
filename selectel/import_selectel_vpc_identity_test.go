package selectel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/require"
)

func TestAccVPCCompositeIdentityImport(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	var forwarding floatingip.PortForwarding
	cases := []struct {
		name, address, config string
		identityFields        []string
		exists, destroy       resource.TestCheckFunc
	}{
		{
			name:           "subnet_route",
			address:        testAccVPCNetworkSubnetRouteV2ResourceName,
			config:         testAccVPCNetworkSubnetRouteV2Basic(projectID, region, "10.0.0.0/8", "192.168.199.10"),
			identityFields: []string{"subnet_id", "destination_cidr", "next_hop", "project_id", "region"},
			exists:         testAccCheckVPCNetworkSubnetRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			destroy:        testAccCheckVPCNetworkSubnetRouteV2Destroy,
		},
		{
			name:           "router_route",
			address:        testAccVPCRouterRouteV2ResourceName,
			config:         testAccVPCRouterRouteV2Basic(projectID, region, "10.0.0.0/8"),
			identityFields: []string{"router_id", "destination_cidr", "next_hop", "project_id", "region"},
			exists:         testAccCheckVPCRouterRouteV2Exists("10.0.0.0/8", "192.168.199.10"),
			destroy:        testAccCheckVPCRouterRouteV2Destroy,
		},
		{
			name:           "port_forwarding",
			address:        testAccVPCPortForwardingV2ResourceName,
			config:         testAccVPCPortForwardingV2Basic(projectID, region, 8080, "tcp"),
			identityFields: []string{"id", "floatingip_id", "project_id", "region"},
			exists:         testAccCheckVPCPortForwardingV2Exists(&forwarding),
			destroy:        testAccCheckVPCPortForwardingV2Destroy,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			identityPath := filepath.Join(t.TempDir(), "identity.json")
			var original *terraform.State
			var originalID string
			var checks []resource.TestCheckFunc
			importConfig := tc.config + fmt.Sprintf(`
import {
  to       = %s
  identity = jsondecode(file(%q))
}
`, tc.address, identityPath)

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
				ProviderFactories: testAccProviders,
				CheckDestroy:      tc.destroy,
				Steps: []resource.TestStep{
					{
						Config: tc.config,
						Check: resource.ComposeTestCheckFunc(tc.exists, func(s *terraform.State) error {
							original = s
							rs := s.RootModule().Resources[tc.address].Primary
							originalID = rs.ID
							identity := make(map[string]string, len(tc.identityFields))
							for _, field := range tc.identityFields {
								value := rs.Attributes[field]
								if field == "id" {
									value = rs.ID
								}
								if value == "" {
									return fmt.Errorf("missing identity field %s on %s", field, tc.address)
								}
								identity[field] = value
							}
							for key, value := range rs.Attributes {
								checks = append(checks, resource.TestCheckResourceAttr(tc.address, key, value))
							}
							data, err := json.Marshal(identity)
							if err != nil {
								return err
							}
							return os.WriteFile(identityPath, data, 0o600)
						}),
					},
					{
						Config: testAccVPCIdentityForgetConfig(t, tc.config, tc.address),
						Check: func(s *terraform.State) error {
							if _, ok := s.RootModule().Resources[tc.address]; ok {
								return fmt.Errorf("%s was not removed from state", tc.address)
							}
							// The remote resource must still exist before importing it.
							return tc.exists(original)
						},
					},
					{
						Config: importConfig,
						Check: resource.ComposeTestCheckFunc(tc.exists, func(s *terraform.State) error {
							rs := s.RootModule().Resources[tc.address].Primary
							if rs.ID != originalID {
								return fmt.Errorf("identity import changed ID: got %s, want %s", rs.ID, originalID)
							}
							return resource.ComposeTestCheckFunc(checks...)(s)
						}),
					},
					{
						Config:   tc.config,
						PlanOnly: true,
					},
				},
			})
		})
	}
}

func testAccVPCIdentityForgetConfig(t *testing.T, config, address string) string {
	t.Helper()
	file, diagnostics := hclwrite.ParseConfig([]byte(config), "test.tf", hcl.InitialPos)
	require.False(t, diagnostics.HasErrors(), diagnostics.Error())
	labels := strings.SplitN(address, ".", 2)
	found := false
	for _, block := range file.Body().Blocks() {
		if block.Type() == "resource" && len(block.Labels()) == 2 &&
			block.Labels()[0] == labels[0] && block.Labels()[1] == labels[1] {
			file.Body().RemoveBlock(block)
			found = true
			break
		}
	}
	require.True(t, found, "resource %s must exist in the fixture", address)
	return string(file.Bytes()) + fmt.Sprintf(`
removed {
  from = %s
  lifecycle {
    destroy = false
  }
}
`, address)
}
