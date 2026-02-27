package selectel

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDedicatedServersV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	serverName := "CL10-SSD"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedServersV1Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDedicatedServersV1Exists("data.selectel_dedicated_servers_v1.servers_tf_acc_test_1", serverName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_servers_v1.servers_tf_acc_test_1", "servers.0.name", serverName),
				),
			},
			{
				Config: testAccDedicatedServersV1WithFilter(projectName, serverName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_dedicated_servers_v1.servers_filtered_tf_acc_test_1", "servers.0.name", serverName),
				),
			},
		},
	})
}

func testAccDedicatedServersV1Exists(n, serverName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dsClient := newTestDedicatedAPIClient(rs, testAccProvider)

		serversFromAPI, _, err := dsClient.ResourcesList(ctx, "", "")
		if err != nil {
			return err
		}

		var serverFound bool
		for _, server := range serversFromAPI {
			if server.Info == serverName {
				serverFound = true
				break
			}
		}

		if !serverFound {
			return fmt.Errorf("server %s not found", serverName)
		}

		return nil
	}
}

func testAccDedicatedServersV1Basic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name = "%s"
}

data "selectel_dedicated_servers_v1" "servers_tf_acc_test_1" {
 project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
}
`, projectName)
}

func testAccDedicatedServersV1WithFilter(projectName, serverName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name = "%s"
}

data "selectel_dedicated_servers_v1" "servers_filtered_tf_acc_test_1" {
 project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

 filter {
   name             = "%s"
 }
}
`, projectName, serverName)
}

func TestFilterDedicatedServers(t *testing.T) {
	type testCase struct {
		name     string
		servers  []dedicated.ResourceDetails
		filter   dedicatedServersSearchFilter
		ips      dedicated.ReservedIPs
		expected int // number of expected results
	}

	// Mock servers for testing
	servers := []dedicated.ResourceDetails{
		{
			UUID:         "server-1",
			Info:         "test-server-1",
			ServiceUUID:  "config-1",
			LocationUUID: "location-1",
		},
		{
			UUID:         "server-2",
			Info:         "test-server-2",
			ServiceUUID:  "config-2",
			LocationUUID: "location-2",
		},
		{
			UUID:         "server-3",
			Info:         "another-server",
			ServiceUUID:  "config-3",
			LocationUUID: "location-3",
		},
	}

	// Mock reserved IPs
	ips := dedicated.ReservedIPs{
		{
			ResourceUUID: "server-1",
			IP:           net.ParseIP("192.168.1.10"),
			Subnet:       "public-subnet-1",
		},
		{
			ResourceUUID: "server-2",
			IP:           net.ParseIP("10.0.0.5"),
			Subnet:       "private-subnet-1",
		},
	}

	testCases := []testCase{
		{
			name:     "Empty filter returns all servers",
			servers:  servers,
			filter:   dedicatedServersSearchFilter{},
			ips:      ips,
			expected: len(servers), // Should return all servers when filter is empty
		},
		{
			name:    "Filter by name",
			servers: servers,
			filter: dedicatedServersSearchFilter{
				name: "test-server-1",
			},
			ips:      ips,
			expected: 1,
		},
		{
			name:    "Filter by non-existent name",
			servers: servers,
			filter: dedicatedServersSearchFilter{
				name: "non-existent",
			},
			ips:      ips,
			expected: 0,
		},
		{
			name:    "Filter by IP",
			servers: servers,
			filter: dedicatedServersSearchFilter{
				ip: "192.168.1.10",
			},
			ips:      ips,
			expected: 1,
		},
		{
			name:    "Filter by subnet",
			servers: servers,
			filter: dedicatedServersSearchFilter{
				publicSubnet: "public-subnet-1",
			},
			ips:      ips,
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filterDedicatedServers(tc.servers, tc.filter, tc.ips)
			if err != nil {
				t.Errorf("filterDedicatedServers returned error: %v", err)
				return
			}

			if len(result) != tc.expected {
				t.Errorf("Expected %d results, got %d", tc.expected, len(result))
			}
		})
	}
}

func TestDedicatedServersSearchFilterIsEmpty(t *testing.T) {
	type testCase struct {
		name     string
		filter   dedicatedServersSearchFilter
		expected bool
	}

	testCases := []testCase{
		{
			name:     "Empty filter",
			filter:   dedicatedServersSearchFilter{},
			expected: true,
		},
		{
			name: "Filter with name",
			filter: dedicatedServersSearchFilter{
				name: "test-server",
			},
			expected: false,
		},
		{
			name: "Filter with IP",
			filter: dedicatedServersSearchFilter{
				ip: "192.168.1.10",
			},
			expected: false,
		},
		{
			name: "Filter with public subnet",
			filter: dedicatedServersSearchFilter{
				publicSubnet: "public-subnet-1",
			},
			expected: false,
		},
		{
			name: "Filter with private subnet",
			filter: dedicatedServersSearchFilter{
				privateSubnet: "private-subnet-1",
			},
			expected: false,
		},
		{
			name: "Filter with all fields empty",
			filter: dedicatedServersSearchFilter{
				name:          "",
				ip:            "",
				publicSubnet:  "",
				privateSubnet: "",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.filter.IsEmpty()
			if result != tc.expected {
				t.Errorf("IsEmpty() = %v, want %v", result, tc.expected)
			}
		})
	}
}
