package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDedicatedConfigurationV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	configurationName := "EL50-SSD"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedConfigurationV1Basic(projectName, configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDedicatedConfigurationV1Exists("data.selectel_dedicated_configuration_v1.server_configuration_tf_acc_test_1", configurationName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_configuration_v1.server_configuration_tf_acc_test_1", "configurations.0.name", configurationName),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_configuration_v1.server_configuration_tf_acc_test_1", "configurations.0.config_name"),
				),
			},
		},
	})
}

func testAccDedicatedConfigurationV1Exists(
	n string, serverName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestDedicatedAPIClient(rs, testAccProvider)

		serversFromAPI, _, err := dsClient.Servers(ctx)
		if err != nil {
			return err
		}

		var srvFromAPI *dedicated.Server
		for _, srv := range serversFromAPI {
			if srv.Name == serverName {
				srvFromAPI = &srv
			}
		}

		if srvFromAPI == nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		return nil
	}
}

func testAccDedicatedConfigurationV1Basic(projectName, configurationName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dedicated_configuration_v1" "server_configuration_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

  deep_filter = "{\"name\": \"%s\"}"
}
`, projectName, configurationName)
}

func TestFilterDedicatedConfigurations(t *testing.T) {
	type testCase struct {
		name     string
		list     []dedicated.Server
		filter   *dedicatedConfigurationsFilter
		expected int // number of expected results
	}

	// Mock servers for testing
	servers := []dedicated.Server{
		{
			ID:         "server-1",
			Name:       "EL50-SSD",
			ConfigName: "el50-ssd-config",
		},
		{
			ID:         "server-2",
			Name:       "EL100-HDD",
			ConfigName: "el100-hdd-config",
		},
		{
			ID:         "server-3",
			Name:       "EL50-HDD",
			ConfigName: "el50-hdd-config",
		},
	}

	testCases := []testCase{
		{
			name: "Deep filter by name",
			list: servers,
			filter: &dedicatedConfigurationsFilter{
				deepFilter: map[string]any{"name": "EL50-SSD"},
			},
			expected: 1,
		},
		{
			name: "Name filter",
			list: servers,
			filter: &dedicatedConfigurationsFilter{
				name: "EL100-HDD",
			},
			expected: 1,
		},
		{
			name:     "Empty filter returns none (because Name != empty string)",
			list:     servers,
			filter:   &dedicatedConfigurationsFilter{},
			expected: 0,
		},
		{
			name: "Non-matching filter",
			list: servers,
			filter: &dedicatedConfigurationsFilter{
				name: "NONEXISTENT",
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterDedicatedConfigurations(tc.list, tc.filter)
			if len(result) != tc.expected {
				t.Errorf("Expected %d results, got %d", tc.expected, len(result))
			}
		})
	}
}

func TestExpandDedicatedConfigurationsSearchFilter(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func() *schema.ResourceData
		expected *dedicatedConfigurationsFilter
		hasError bool
	}{
		{
			name: "Valid deep filter",
			setupFn: func() *schema.ResourceData {
				d := dataSourceDedicatedConfigurationV1().TestResourceData()
				d.Set("deep_filter", `{"name": "test-config"}`)
				return d
			},
			expected: &dedicatedConfigurationsFilter{
				deepFilter: map[string]any{"name": "test-config"},
				name:       "",
				locationID: "",
			},
			hasError: false,
		},
		{
			name: "Valid name filter",
			setupFn: func() *schema.ResourceData {
				d := dataSourceDedicatedConfigurationV1().TestResourceData()
				// Create a proper filter set with the correct schema
				filterSchema := &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				}
				filterSet := schema.NewSet(schema.HashResource(filterSchema), []interface{}{})
				filterSet.Add(map[string]interface{}{
					"name":        "test-name",
					"location_id": "",
				})
				d.Set("filter", filterSet)

				return d
			},
			expected: &dedicatedConfigurationsFilter{
				deepFilter: map[string]any{},
				name:       "test-name",
				locationID: "",
			},
			hasError: false,
		},
		{
			name: "Invalid deep filter JSON",
			setupFn: func() *schema.ResourceData {
				d := dataSourceDedicatedConfigurationV1().TestResourceData()
				d.Set("deep_filter", `{"invalid": json}`)
				return d
			},
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupFn()

			result, err := expandDedicatedConfigurationsSearchFilter(d)

			if (err != nil) != tt.hasError {
				t.Errorf("expandDedicatedConfigurationsSearchFilter() error = %v, wantErr %v", err, tt.hasError)
				return
			}

			if !tt.hasError && result != nil {
				// Compare deepFilter
				if len(tt.expected.deepFilter) != len(result.deepFilter) {
					t.Errorf("expandDedicatedConfigurationsSearchFilter() deepFilter length mismatch: expected %v, got %v", tt.expected.deepFilter, result.deepFilter)
				}

				// Check individual keys in deepFilter
				for k, v := range tt.expected.deepFilter {
					if result.deepFilter[k] != v {
						t.Errorf("expandDedicatedConfigurationsSearchFilter() deepFilter mismatch for key %s: expected %v, got %v", k, v, result.deepFilter[k])
					}
				}

				if result.name != tt.expected.name {
					t.Errorf("expandDedicatedConfigurationsSearchFilter() name = %v, want %v", result.name, tt.expected.name)
				}

				if result.locationID != tt.expected.locationID {
					t.Errorf("expandDedicatedConfigurationsSearchFilter() locationID = %v, want %v", result.locationID, tt.expected.locationID)
				}
			}
		})
	}
}
