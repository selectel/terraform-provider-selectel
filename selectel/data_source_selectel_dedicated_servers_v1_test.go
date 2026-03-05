package selectel

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
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

func TestFilterDedicatedServers_PartialNameMatch(t *testing.T) {
	servers := []dedicated.ResourceDetails{
		{UUID: "1", Info: "web-server-01"},
		{UUID: "2", Info: "WEB-SERVER-02"},
		{UUID: "3", Info: "db-server-01"},
		{UUID: "4", Info: "app-server"},
	}

	t.Run("Partial match case-insensitive", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "web"}
		result, err := filterDedicatedServers(servers, filter, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 2, "should match 'web-server-01' and 'WEB-SERVER-02'")
	})

	t.Run("Partial match middle of string", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "server"}
		result, err := filterDedicatedServers(servers, filter, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 4, "should match all servers containing 'server'")
	})

	t.Run("Exact match", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "web-server-01"}
		result, err := filterDedicatedServers(servers, filter, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("Empty name filter", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: ""}
		result, err := filterDedicatedServers(servers, filter, nil)
		assert.NoError(t, err)
		assert.Len(t, result, len(servers), "empty filter should return all servers")
	})
}

func TestFilterDedicatedServers_CombinedFilters(t *testing.T) {
	servers := []dedicated.ResourceDetails{
		{UUID: "1", Info: "web-01", ServiceUUID: "config-1", LocationUUID: "loc-1"},
		{UUID: "2", Info: "web-02", ServiceUUID: "config-1", LocationUUID: "loc-1"},
		{UUID: "3", Info: "db-01", ServiceUUID: "config-2", LocationUUID: "loc-2"},
	}

	ips := dedicated.ReservedIPs{
		{ResourceUUID: "1", IP: net.ParseIP("192.168.1.10"), Subnet: "subnet-1"},
		{ResourceUUID: "2", IP: net.ParseIP("192.168.1.11"), Subnet: "subnet-1"},
		{ResourceUUID: "3", IP: net.ParseIP("10.0.0.5"), Subnet: "subnet-2"},
	}

	t.Run("Name and IP filter", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "web", ip: "192.168.1.10"}
		result, err := filterDedicatedServers(servers, filter, ips)
		assert.NoError(t, err)
		assert.Len(t, result, 1, "should match only web-01 with specific IP")
		assert.Equal(t, "1", result[0].UUID)
	})

	t.Run("Name and subnet filter", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "web", publicSubnet: "subnet-1"}
		result, err := filterDedicatedServers(servers, filter, ips)
		assert.NoError(t, err)
		assert.Len(t, result, 2, "should match both web servers in subnet-1")
	})

	t.Run("Private subnet filter", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{privateSubnet: "subnet-2"}
		result, err := filterDedicatedServers(servers, filter, ips)
		assert.NoError(t, err)
		assert.Len(t, result, 1, "should match only db-01 in subnet-2")
	})

	t.Run("No matching IP", func(t *testing.T) {
		filter := dedicatedServersSearchFilter{name: "web", ip: "10.10.10.10"}
		result, err := filterDedicatedServers(servers, filter, ips)
		assert.NoError(t, err)
		assert.Len(t, result, 0, "should not match any server with non-existent IP")
	})
}

func TestExpandDedicatedServersSearchFilter(t *testing.T) {
	resource := dataSourceDedicatedServersV1()
	d := resource.TestResourceData()

	filterSet := schema.NewSet(schema.HashResource(resource.Schema["filter"].Elem.(*schema.Resource)), []interface{}{})
	filterSet.Add(map[string]interface{}{
		"name":             "test-server",
		"ip":               "192.168.1.100",
		"location_id":      "loc-uuid-123",
		"configuration_id": "config-uuid-456",
		"public_subnet":    "public-subnet-1",
		"private_subnet":   "private-subnet-1",
	})
	_ = d.Set("filter", filterSet)

	filter := expandDedicatedServersSearchFilter(d)

	assert.Equal(t, "test-server", filter.name)
	assert.Equal(t, "192.168.1.100", filter.ip)
	assert.Equal(t, "loc-uuid-123", filter.locationID)
	assert.Equal(t, "config-uuid-456", filter.configurationID)
	assert.Equal(t, "public-subnet-1", filter.publicSubnet)
	assert.Equal(t, "private-subnet-1", filter.privateSubnet)
}

func TestExpandDedicatedServersSearchFilter_Empty(t *testing.T) {
	resource := dataSourceDedicatedServersV1()
	d := resource.TestResourceData()

	filter := expandDedicatedServersSearchFilter(d)

	assert.Equal(t, "", filter.name)
	assert.Equal(t, "", filter.ip)
	assert.Equal(t, "", filter.locationID)
	assert.Equal(t, "", filter.configurationID)
	assert.Equal(t, "", filter.publicSubnet)
	assert.Equal(t, "", filter.privateSubnet)
}

func TestFlattenDedicatedServers(t *testing.T) {
	servers := []dedicated.ResourceDetails{
		{
			UUID:         "server-uuid-1",
			Info:         "test-server-1",
			ServiceUUID:  "config-uuid-1",
			LocationUUID: "location-uuid-1",
		},
		{
			UUID:         "server-uuid-2",
			Info:         "test-server-2",
			ServiceUUID:  "config-uuid-2",
			LocationUUID: "location-uuid-2",
		},
	}

	reservedPublicIPs := dedicated.ReservedIPs{
		{ResourceUUID: "server-uuid-1", IP: net.ParseIP("192.168.1.10")},
		{ResourceUUID: "server-uuid-2", IP: net.ParseIP("192.168.1.11")},
	}

	reservedPrivateIPs := dedicated.ReservedIPs{
		{ResourceUUID: "server-uuid-1", IP: net.ParseIP("10.0.0.5")},
	}

	result := flattenDedicatedServers(servers, reservedPublicIPs, reservedPrivateIPs)

	assert.Len(t, result, 2)

	// Check first server
	server1 := result[0].(map[string]interface{})
	assert.Equal(t, "server-uuid-1", server1["id"])
	assert.Equal(t, "test-server-1", server1["name"])
	assert.Equal(t, "config-uuid-1", server1["configuration_id"])
	assert.Equal(t, "location-uuid-1", server1["location_id"])
	publicIPs1 := server1["reserved_public_ips"].([]string)
	assert.Contains(t, publicIPs1, "192.168.1.10")
	privateIPs1 := server1["reserved_private_ips"].([]string)
	assert.Contains(t, privateIPs1, "10.0.0.5")

	// Check second server
	server2 := result[1].(map[string]interface{})
	assert.Equal(t, "server-uuid-2", server2["id"])
	assert.Equal(t, "test-server-2", server2["name"])
	assert.Equal(t, "config-uuid-2", server2["configuration_id"])
	assert.Equal(t, "location-uuid-2", server2["location_id"])
	publicIPs2 := server2["reserved_public_ips"].([]string)
	assert.Contains(t, publicIPs2, "192.168.1.11")
	privateIPs2 := server2["reserved_private_ips"].([]string)
	assert.Empty(t, privateIPs2, "should have no private IPs")
}

func TestFlattenDedicatedServers_NoIPs(t *testing.T) {
	servers := []dedicated.ResourceDetails{
		{
			UUID:         "server-uuid-1",
			Info:         "test-server-1",
			ServiceUUID:  "config-uuid-1",
			LocationUUID: "location-uuid-1",
		},
	}

	result := flattenDedicatedServers(servers, nil, nil)

	assert.Len(t, result, 1)
	server1 := result[0].(map[string]interface{})
	assert.Empty(t, server1["reserved_public_ips"].([]string))
	assert.Empty(t, server1["reserved_private_ips"].([]string))
}
