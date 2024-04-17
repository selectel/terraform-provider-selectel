package selectel

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomainsRRSetV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainsRRSetV2Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"records": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDomainsRRSetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	rrsetName := d.Get("name").(string)
	zoneID := d.Get("zone_id").(string)
	rrsetType := d.Get("type").(string)

	zoneIDWithRRSetNameAndType := fmt.Sprintf("zone_id: %s, rrset_name: %s, rrset_type: %s", zoneID, rrsetName, rrsetType)
	log.Println(msgGet(objectRRSet, zoneIDWithRRSetNameAndType))

	rrset, err := getRRSetByNameAndType(ctx, client, zoneID, rrsetName, rrsetType)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRRSetToResourceData(d, rrset)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
