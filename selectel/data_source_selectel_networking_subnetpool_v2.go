package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnetpool"
)

func dataSourceNetworkingSubnetPoolV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingSubnetPoolV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnetpool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"default_quota": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"default_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"min_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"max_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"address_scope_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkingSubnetPoolV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"subnetpool_id":    "id",
		"name":             "name",
		"description":      "description",
		"address_scope_id": "address_scope_id",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	for schemaKey, queryKey := range map[string]string{
		"default_quota":     "default_quota",
		"default_prefixlen": "default_prefixlen",
		"min_prefixlen":     "min_prefixlen",
		"max_prefixlen":     "max_prefixlen",
		"ip_version":        "ip_version",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{strconv.Itoa(v.(int))}
		}
	}
	for _, schemaKey := range []string{"shared", "is_default"} {
		if v, ok := d.GetOkExists(schemaKey); ok {
			filters[schemaKey] = []string{strconv.FormatBool(v.(bool))}
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(expandToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := subnetpool.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingSubnetPoolV2Object, err))
	}

	switch len(found) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingSubnetPoolV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(found), networkingSubnetPoolV2Object))
	}

	result := found[0]

	d.SetId(result.ID)
	d.Set("subnetpool_id", result.ID)
	flattenNetworkingSubnetPoolV2(d, &result)

	log.Printf("[DEBUG] found networking subnetpool %s", result.ID)

	return nil
}
