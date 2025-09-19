package selectel

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicatedservers"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestAccDedicatedServersServerV1Basic(t *testing.T) {
	var (
		project projects.Project

		projectName = acctest.RandomWithPrefix("tf-acc")

		osName                        = "Ubuntu"
		osVersion, updatedOSVersion   = "2404", "2204"
		locationName                  = "MSK-2"
		cfgName                       = "CL25-NVMe"
		pricePlanName                 = "1 день"
		osHostName, updatedOSHostName = "hostname", "hostname1"
		osPassword, updatedOSPassword = "Passw0rd!", "Passw0rd!1"
		userData, updatedUserData     = "#!/bin/bash", "#!/bin/sh"
		sshKey                        = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCOIWeVNMRC7Y9jeBoG5GP3irOf/u5EbuHYixuZEmtHDtmtlnmzdcBEnpPY5OlFhjSySlUC1clCIShMXgWBfdnvk7Dbp5hgCP3Lh9pS/b8e3kxstIiGF9d7IX04DfVTOF424LlMAFbHNsrmX+uU3lizO20DljFIJNML0OdUO7eKg0XOK1OWVQlSzvZbFj39oYTSqCtoI92czQf4DdJ+0IF1/ZNewE6xPohfnZp5cl82UjYs8vxmcaiifVf7kUyQe/ilv/fZYpt59KCJBJDrTU/ko9hNxCVXrCOw7pPOQayoQ2Vir3M1AnhDMunoxFBocndgffNXVQYtA/3TXLVB7feb"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			// create case
			{
				Config: testAccDedicatedServersServerV1(projectName, osName, updatedOSVersion, locationName, cfgName, pricePlanName, osHostName, sshKey, osPassword, userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDedicatedServersServerV1Exists("selectel_dedicated_servers_server_v1.server_tf_acc_test_1"),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "price_plan_name", pricePlanName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_host_name", osHostName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "user_data", userData),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_password", osPassword),
				),
			},
			// update cases
			{
				Config: testAccDedicatedServersServerV1(projectName, osName, osVersion, locationName, cfgName, pricePlanName, updatedOSHostName, sshKey, osPassword, userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDedicatedServersServerV1Exists("selectel_dedicated_servers_server_v1.server_tf_acc_test_1"),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "price_plan_name", pricePlanName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_host_name", updatedOSHostName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "user_data", userData),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_password", osPassword),
				),
			},
			{
				Config: testAccDedicatedServersServerV1(projectName, osName, updatedOSVersion, locationName, cfgName, pricePlanName, updatedOSHostName, sshKey, updatedOSPassword, userData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDedicatedServersServerV1Exists("selectel_dedicated_servers_server_v1.server_tf_acc_test_1"),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "price_plan_name", pricePlanName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_host_name", updatedOSHostName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "user_data", userData),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_password", updatedOSPassword),
				),
			},
			{
				Config: testAccDedicatedServersServerV1(projectName, osName, osVersion, locationName, cfgName, pricePlanName, updatedOSHostName, sshKey, updatedOSPassword, updatedUserData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDedicatedServersServerV1Exists("selectel_dedicated_servers_server_v1.server_tf_acc_test_1"),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "price_plan_name", pricePlanName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_host_name", updatedOSHostName),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "user_data", updatedUserData),
					resource.TestCheckResourceAttr("selectel_dedicated_servers_server_v1.server_tf_acc_test_1", "os_password", updatedOSPassword),
				),
			},
		},
	})
}

func testAccCheckDedicatedServersServerV1Exists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		cl := newTestDedicatedServersAPIClient(rs, testAccProvider)

		res, _, err := cl.ResourceDetails(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if res.UUID != rs.Primary.ID {
			return fmt.Errorf("resource not found %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDedicatedServersServerV1(
	projectName, osName, osVersion, locationName, cfgName, pricePlanName, osHostName, sshKey, osPassword, userData string,
) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_servers_os_v1" "os_tf_acc_test_1" {
 project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

 filter {
   name             = "%s"
   version          = "%s"
 }
}

data "selectel_dedicated_servers_location_v1" "location_tf_acc_test_1" {
 project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
 filter {
   name = "%s"
 }
}

data "selectel_dedicated_servers_configuration_v1" "server_configuration_tf_acc_test_1" {
 project_id     = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
 filter {
   name           = "%s"
 }
}

resource "selectel_dedicated_servers_server_v1" "server_tf_acc_test_1" {
 project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

 configuration_id = "${data.selectel_dedicated_servers_configuration_v1.server_configuration_tf_acc_test_1.configurations.0.id}"
 location_id      = "${data.selectel_dedicated_servers_location_v1.location_tf_acc_test_1.locations[0].id}"
 os_id            = "${data.selectel_dedicated_servers_os_v1.os_tf_acc_test_1.os.0.id}"
 price_plan_name  = "%s"

 os_host_name     = "%s"

 ssh_key         = "%s"

 os_password        = "%s"

 user_data = "%s"

 partitions_config {
   soft_raid_config {
     name      = "first-raid"
     level     = "raid1"
     disk_type = "SSD NVMe M.2"
   }

   disk_partitions {
     mount = "/boot"
     size  = 1
     raid  = "first-raid"
   }
   disk_partitions {
     mount = "swap"
     # size  = 12
     size_percent = 10.5
     raid         = "first-raid"
   }
   disk_partitions {
     mount = "/"
     size  = -1
     raid  = "first-raid"
   }
   disk_partitions {
     mount   = "second_folder"
     size    = 400
     raid    = "first-raid"
     fs_type = "xfs"
   }
 }
}
`, projectName, osName, osVersion, locationName, cfgName, pricePlanName, osHostName, sshKey, osPassword, userData)
}

func Test_resourceDedicatedServersServerV1CreateValidatePreconditions(t *testing.T) {
	const (
		locationID      = "loc1"
		pricePlanID     = "plan1"
		configurationID = "conf1"
		osID            = "os1"
	)

	defaultData := func() *serversDedicatedServerV1CreateData {
		return &serversDedicatedServerV1CreateData{
			server: &dedicatedservers.Server{
				Available: []*dedicatedservers.ServerAvailable{
					{
						LocationID: locationID,
						PlanCount: []*dedicatedservers.ServerAvailablePricePlan{
							{PlanUUID: pricePlanID, Count: 1},
						},
					},
				},
				PricePlanAvailable: []string{pricePlanID},
				Tags:               []string{},
			},
			os: &dedicatedservers.OperatingSystem{
				UUID:            osID,
				ScriptAllowed:   true,
				IsSSHKeyAllowed: true,
				Partitioning:    true,
				TemplateVersion: "v2",
				OSValue:         "linux",
			},
			billing: &dedicatedservers.ServiceBilling{
				HasEnoughBalance: true,
			},
			partitions: dedicatedservers.PartitionsConfig{},
		}
	}

	tests := []struct {
		name          string
		isServerChip  bool
		needUserScrip bool
		needSSHKey    bool
		needPrivateIP bool
		data          *serversDedicatedServerV1CreateData
		wantErr       string
	}{
		{
			name: "Success",
			data: defaultData(),
		},
		{
			name: "LocationNotAvailable",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.server.Available = nil
				return d
			}(),
			wantErr: "is not available for",
		},
		{
			name: "PricePlanNotAvailableForLocation",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.server.PricePlanAvailable = nil
				return d
			}(),
			wantErr: "price-plan plan1 is not available for",
		},
		{
			name: "OSNotAvailable",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.os = nil
				return d
			}(),
			wantErr: "is not available for",
		},
		{
			name:          "UserScriptNotAllowed",
			needUserScrip: true,
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.os.ScriptAllowed = false
				return d
			}(),
			wantErr: "does not allow scripts",
		},
		{
			name:       "SSHKeyNotAllowed",
			needSSHKey: true,
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.os.IsSSHKeyAllowed = false
				return d
			}(),
			wantErr: "does not allow SSH keys",
		},
		{
			name: "PartitioningNotSupported",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.os.Partitioning = false
				d.partitions = map[string]*dedicatedservers.PartitionConfigItem{"a": {}}

				return d
			}(),
			wantErr: "does not support partitions config",
		},
		{
			name: "InsufficientBalance",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.billing.HasEnoughBalance = false
				return d
			}(),
			wantErr: "insufficient balance",
		},
		{
			name:          "PrivateIPNotSupportedByServer",
			needPrivateIP: true,
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.server.Tags = []string{"10GE_Internet"}
				return d
			}(),
			wantErr: "does not support private network",
		},
		{
			name:          "PrivateIPNotSupportedByOS",
			needPrivateIP: true,
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.os.TemplateVersion = "v1"
				return d
			}(),
			wantErr: "does not support private network",
		},
		{
			name: "PartitionsValidationFails",
			data: func() *serversDedicatedServerV1CreateData {
				d := defaultData()
				d.partitions = map[string]*dedicatedservers.PartitionConfigItem{"a": {}}
				return d
			}(),
			wantErr: "failed to validate partitions config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(dedicatedservers.ServiceClient)
			client.HTTPClient = &http.Client{
				Transport: httptest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					if strings.Contains(req.URL.Path, "validate") {
						if tt.wantErr != "" && strings.Contains(tt.wantErr, "validate") {
							return httptest.NewFakeResponse(http.StatusBadRequest, `{"error": "validation failed"}`), nil
						}

						return httptest.NewFakeResponse(http.StatusOK, `{"partitions_config": {}}`), nil
					}

					return httptest.NewFakeResponse(http.StatusNotFound, `{}`), nil
				}),
			}

			err := resourceDedicatedServersServerV1CreateValidatePreconditions(
				context.Background(), client, tt.data, locationID, pricePlanID, configurationID, osID,
				tt.needUserScrip, tt.needSSHKey, tt.needPrivateIP,
			)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_resourceDedicatedServersServerV1UpdateValidatePreconditions(t *testing.T) {
	const (
		osID = "os1"
	)

	defaultOS := func() *dedicatedservers.OperatingSystem {
		return &dedicatedservers.OperatingSystem{
			UUID:            osID,
			ScriptAllowed:   true,
			IsSSHKeyAllowed: true,
			Partitioning:    true,
			TemplateVersion: "v2",
			OSValue:         "linux",
		}
	}

	tests := []struct {
		name         string
		os           *dedicatedservers.OperatingSystem
		partitions   dedicatedservers.PartitionsConfig
		needUserData bool
		needSSHKey   bool
		changes      []string
		wantErr      string
	}{
		{
			name:    "SuccessOSIDChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID},
		},
		{
			name:    "SuccessOSHostNameChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName},
		},
		{
			name:       "SuccessOSSSHKeyChanged",
			os:         defaultOS(),
			needSSHKey: true,
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSSSHKey},
		},
		{
			name:       "SuccessOSSSHKeyNameChanged",
			os:         defaultOS(),
			needSSHKey: true,
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSSSHKeyName},
		},
		{
			name:    "SuccessOSPasswordChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSPassword},
		},
		{
			name:       "SuccessOSPartitionsConfigChanged",
			os:         defaultOS(),
			partitions: map[string]*dedicatedservers.PartitionConfigItem{},
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSPartitionsConfig},
		},
		{
			name:         "SuccessOSUserDataChanged",
			os:           defaultOS(),
			needUserData: true,
			changes:      []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSUserData},
		},
		{
			name:    "NoOSConfigChanged",
			os:      defaultOS(),
			changes: []string{},
			wantErr: "can't update cause os configuration has not changed",
		},
		{
			name:    "ProjectIDChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyProjectID},
			wantErr: "can't update cause project ID has changed",
		},
		{
			name:    "LocationIDChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyLocationID},
			wantErr: "can't update cause location ID has changed",
		},
		{
			name:    "ConfigurationIDChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyConfigurationID},
			wantErr: "can't update cause configuration ID has changed",
		},
		{
			name:    "PricePlanNameChanged",
			os:      defaultOS(),
			changes: []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyPricePlanName},
			wantErr: "can't update cause price plan name has changed",
		},
		{
			name:         "UserScriptNotAllowed",
			os:           func() *dedicatedservers.OperatingSystem { o := defaultOS(); o.ScriptAllowed = false; return o }(),
			needUserData: true,
			changes:      []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSUserData},
			wantErr:      "does not allow scripts",
		},
		{
			name:       "SSHKeyNotAllowed",
			os:         func() *dedicatedservers.OperatingSystem { o := defaultOS(); o.IsSSHKeyAllowed = false; return o }(),
			needSSHKey: true,
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSSSHKey},
			wantErr:    "does not allow SSH keys",
		},
		{
			name:       "PartitionsNotSupported",
			os:         func() *dedicatedservers.OperatingSystem { o := defaultOS(); o.Partitioning = false; return o }(),
			partitions: map[string]*dedicatedservers.PartitionConfigItem{"a": {}},
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyOSPartitionsConfig},
			wantErr:    "does not support partitions config",
		},
		{
			name:       "PartitionsValidationFails",
			os:         defaultOS(),
			partitions: map[string]*dedicatedservers.PartitionConfigItem{"a": {}},
			changes:    []string{dedicatedServersServerSchemaKeyOSID, dedicatedServersServerSchemaKeyOSHostName, dedicatedServersServerSchemaKeyOSPartitionsConfig},
			wantErr:    "failed to validate partitions config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := resourceDedicatedServersServerV1()

			initMap := map[string]interface{}{}
			for _, key := range tt.changes {
				initMap[key] = "changed"
			}

			d := schema.TestResourceDataRaw(t, res.Schema, initMap)

			client := new(dedicatedservers.ServiceClient)
			client.HTTPClient = &http.Client{
				Transport: httptest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					if strings.Contains(req.URL.Path, "validate") {
						if tt.wantErr != "" && strings.Contains(tt.wantErr, "validate") {
							return httptest.NewFakeResponse(http.StatusBadRequest, `{"error": "validation failed"}`), nil
						}

						return httptest.NewFakeResponse(http.StatusOK, `{"partitions_config": {}}`), nil
					}

					return httptest.NewFakeResponse(http.StatusNotFound, `{}`), nil
				}),
			}

			err := resourceDedicatedServersServerV1UpdateValidatePreconditions(
				context.Background(), d, client, tt.os, tt.partitions, tt.needUserData, tt.needSSHKey,
			)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
