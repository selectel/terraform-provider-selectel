package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/addressscope"
)

func dataSourceNetworkingAddressScopeV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingAddressScopeV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			// Selectable here even though the resource cannot set it: searching
			// for the shared scopes an admin created is a legitimate read.
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkingAddressScopeV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	if v, ok := d.GetOk("name"); ok {
		filters["name"] = []string{v.(string)}
	}
	if v, ok := d.GetOk("ip_version"); ok {
		filters["ip_version"] = []string{strconv.Itoa(v.(int))}
	}
	if v, ok := d.GetOkExists("shared"); ok { //nolint:staticcheck // GetOkExists distinguishes an explicit false from an unset bool
		filters["shared"] = []string{strconv.FormatBool(v.(bool))}
	}

	found, err := addressscope.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingAddressScopeV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingAddressScopeV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), networkingAddressScopeV2Object))
	}

	result := found[0]

	// Every attribute comes from the found object, not from the selectors that
	// were passed in.
	d.SetId(result.ID)
	flattenNetworkingAddressScopeV2(d, &result)

	log.Printf("[DEBUG] found networking address scope %s", result.ID)

	return nil
}
