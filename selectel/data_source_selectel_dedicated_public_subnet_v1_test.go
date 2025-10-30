package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
)

func TestAccDedicatedPublicSubnetV1Basic(t *testing.T) {
	var project projects.Project

	projectName := acctest.RandomWithPrefix("tf-acc")
	locationName := "MSK-2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedPublicSubnetV1Basic(projectName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("data.selectel_dedicated_public_subnet_v1.public_subnet_tf_acc_test_1", "subnets.#", "0"),
				),
			},
		},
	})
}

func testAccDedicatedPublicSubnetV1Basic(projectName, locationName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

data "selectel_dedicated_location_v1" "location_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  filter {
    name = "%s"
  }
}

data "selectel_dedicated_public_subnet_v1" "public_subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"

  filter {
    location_id = "${data.selectel_dedicated_location_v1.location_tf_acc_test_1.locations.0.id}"
  }
}
`, projectName, locationName)
}

func Test_filterDedicatedPublicSubnets(t *testing.T) {
	unfilteredList := dedicated.Subnets{
		{
			UUID:   "1",
			Subnet: "192.168.1.0/24",
		},
		{
			UUID:   "2",
			Subnet: "10.0.0.0/16",
		},
	}

	type args struct {
		subnets dedicated.Subnets
		filter  dedicatedPublicSubnetsSearchFilter
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
				filter:  dedicatedPublicSubnetsSearchFilter{},
			},
			want: unfilteredList,
		},
		{
			name: "IncludeSubnet",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPublicSubnetsSearchFilter{
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
				filter: dedicatedPublicSubnetsSearchFilter{
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
				filter: dedicatedPublicSubnetsSearchFilter{
					ip:     "10.0.0.1",
					subnet: unfilteredList[1].Subnet,
				},
			},
			want: dedicated.Subnets{
				unfilteredList[1],
			},
		},
		{
			name: "NoMatches",
			args: args{
				subnets: unfilteredList,
				filter: dedicatedPublicSubnetsSearchFilter{
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
				filter: dedicatedPublicSubnetsSearchFilter{
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
				filter: dedicatedPublicSubnetsSearchFilter{
					ip: "invalid",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterDedicatedPublicSubnets(tt.args.subnets, tt.args.filter)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equalf(t, tt.want, got, "filterDedicatedPublicSubnets(%v, %v)", tt.args.subnets, tt.args.filter)
		})
	}
}
