package selectel

import (
	"fmt"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/stretchr/testify/assert"
)

func TestAccDedicatedPrivateSubnetV1DataSource(t *testing.T) {
	var (
		project      projects.Project
		projectName  = acctest.RandomWithPrefix("tf-acc")
		locationName = "SPB-5"
		vlan         = "2989"
		subnet       = "10.10.10.0/24"
		ip           = "10.10.10.10"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithAuth(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#"),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.0.id"),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.0.subnet"),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.0.vlan"),
					resource.TestCheckResourceAttrSet("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.0.reserved_ip.#"),
				),
			},
			{
				Config:   testAccDedicatedPrivateSubnetV1DataSourceBasic(projectName),
				PlanOnly: true,
			},
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceWithLocationFilter(projectName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccDedicatedLocationV1Exists("data.selectel_dedicated_location_v1.location_tf_acc_test_1", locationName),
					resource.TestCheckResourceAttr("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#", "1"),
				),
			},
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceWithVLANFilter(projectName, vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#", "1"),
					resource.TestCheckResourceAttr(
						"data.selectel_dedicated_private_subnet_v1.subnet_ds",
						"subnets.0.vlan",
						vlan,
					),
				),
			},
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceWithSubnetFilter(projectName, subnet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#", "1"),
					resource.TestCheckResourceAttr(
						"data.selectel_dedicated_private_subnet_v1.subnet_ds",
						"subnets.0.subnet",
						subnet,
					),
				),
			},
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceWithIPFilter(projectName, ip),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#", "1"),
					resource.TestCheckResourceAttr(
						"data.selectel_dedicated_private_subnet_v1.subnet_ds",
						"subnets.0.subnet",
						subnet,
					),
				),
			},
			{
				Config: testAccDedicatedPrivateSubnetV1DataSourceWithIPFilter(projectName, "1.1.1.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_dedicated_private_subnet_v1.subnet_ds", "subnets.#", "0"),
				),
			},
			{
				Config:      testAccDedicatedPrivateSubnetV1DataSourceWithVLANFilter(projectName, "wrong"),
				ExpectError: regexp.MustCompile(`invalid parameters to filter on tag or vlan`),
			},
			{
				Config:      testAccDedicatedPrivateSubnetV1DataSourceWithIPFilter(projectName, "wrong"),
				ExpectError: regexp.MustCompile(`invalid`),
			},
		},
	})
}

func Test_filterDedicatedPrivateSubnets(t *testing.T) {
	unfilteredList := dedicated.Subnets{
		{
			UUID:   "1",
			Subnet: "192.168.1.0/24",
		},
		{
			UUID:   "2",
			Subnet: "10.0.0.0/16",
		},
		{
			UUID:   "3",
			Subnet: "172.16.0.0/12",
		},
	}

	type args struct {
		subnets dedicated.Subnets
		filter  dedicatedPrivateSubnetsSearchFilter
	}
	tests := []struct {
		name    string
		args    args
		want    dedicated.Subnets
		wantErr bool
	}{
		{
			name: "EmptyFilter",
			args: args{
				subnets: unfilteredList,
				filter:  dedicatedPrivateSubnetsSearchFilter{},
			},
			want: unfilteredList,
		},
		{
			name: "IncludeSubnet",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					subnet: unfilteredList[0].Subnet,
				},
			},
			want: dedicated.Subnets{
				unfilteredList[0],
			},
		},
		{
			name: "IncludeIP",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip: "10.0.0.1",
				},
			},
			want: dedicated.Subnets{
				unfilteredList[1],
			},
		},
		{
			name: "IncludeIPAndSubnet",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip:     "10.0.0.1",
					subnet: unfilteredList[1].Subnet,
				},
			},
			want: dedicated.Subnets{
				unfilteredList[1],
			},
		},
		{
			name: "IncludeIPFrom172Range",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip: "172.16.0.10",
				},
			},
			want: dedicated.Subnets{
				unfilteredList[2],
			},
		},
		{
			name: "NoMatches",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip:     "10.0.0.1",
					subnet: unfilteredList[0].Subnet,
				},
			},
			want: nil,
		},
		{
			name: "NoMatches2",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip:     "8.0.0.1",
					subnet: "8.0.0.0/24",
				},
			},
			want: nil,
		},
		{
			name: "InvalidIPErr",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPrivateSubnetsSearchFilter{
					ip: "invalid",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterDedicatedPrivateSubnets(tt.args.subnets, tt.args.filter)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equalf(t, tt.want, got, "filterDedicatedPrivateSubnets(%v, %v)", tt.args.subnets, tt.args.filter)
		})
	}
}

func testAccDedicatedPrivateSubnetV1DataSourceBasic(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
}`, projectName)
}

func testAccDedicatedPrivateSubnetV1DataSourceWithLocationFilter(projectName, locationName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  
  filter {
    location_id = "${data.selectel_dedicated_location_v1.location_tf_acc_test_1.locations[0].id}"
  }
}`, projectName, locationName)
}

func testAccDedicatedPrivateSubnetV1DataSourceWithVLANFilter(projectName, vlan string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  
  filter {
    vlan = "%s"
  }
}`, projectName, vlan)
}

func testAccDedicatedPrivateSubnetV1DataSourceWithSubnetFilter(projectName, subnet string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  
  filter {
    subnet = "%s"
  }
}`, projectName, subnet)
}

func testAccDedicatedPrivateSubnetV1DataSourceWithIPFilter(projectName, ip string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
 name        = "%s"
}

data "selectel_dedicated_private_subnet_v1" "subnet_ds" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  
  filter {
    ip = "%s"
  }
}`, projectName, ip)
}
