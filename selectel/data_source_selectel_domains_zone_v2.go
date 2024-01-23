package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	ErrZoneNotFound       = errors.New("zone not found")
	ErrFoundMultipleZones = errors.New("found multiple zones")
)

func dataSourceDomainsZoneV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainsZoneV2Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delegation_checked_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_check_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_delegated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainsZoneV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneName := d.Get("name").(string)

	log.Println(msgGet(objectZone, zoneName))

	zone, err := getZoneByName(ctx, client, zoneName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setZoneToResourceData(d, zone)
	if err != nil {
		return diag.FromErr(errGettingObject(objectZone, zoneName, err))
	}

	return nil
}
