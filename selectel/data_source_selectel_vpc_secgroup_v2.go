package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
)

var vpcSecGroupDataSourceDocs = dataSourceDocs{Name: "security group"}

func dataSourceVPCSecGroupV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCSecGroupV2Read,
		Description: "Provides a security group. " +
			securityGroupsMoreInfo + "\n\n" +
			upgradeGuideNote,
		Schema: map[string]*schema.Schema{
			"id":         vpcSecGroupDataSourceDocs.idSchema(),
			"project_id": vpcSecGroupDataSourceDocs.projectIDSchema(),
			"region":     vpcSecGroupDataSourceDocs.regionSchema(),
			"secgroup_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the security group. " +
					secGroupIDFromControlPanel,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the security group.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the security group.",
			},
			"stateful": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the security group is stateful.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the security group.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the security group.",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if another project has been granted access to this group.",
			},
			"security_group_rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of rules contained in the security group.",
			},
		},
	}
}

func dataSourceVPCSecGroupV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"secgroup_id": "id",
		"name":        "name",
		"description": "description",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(convert.ToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := securitygroup.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(vpcSecGroupV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			vpcSecGroupV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), vpcSecGroupV2Object))
	}

	result := found[0]

	d.SetId(result.ID)
	d.Set("secgroup_id", result.ID)
	flattenVPCSecGroupV2(d, &result)

	log.Printf("[DEBUG] found VPC security group %s", result.ID)

	return nil
}
