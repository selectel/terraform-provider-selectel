package selectel

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

func dataSourceGlobalRouterZoneGroupV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRouterZoneGroupV1Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
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
		},
	}
}

func dataSourceGlobalRouterZoneGroupV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}
	zoneGroupName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterZoneGroup, zoneGroupName, errors.New("'name' should have type string")),
		)
	}

	zoneGroup, err := getZoneGroupByParams(ctx, client, zoneGroupName)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterZoneGroup, zoneGroupName, err))
	}

	err = setGRZoneGroupToResourceData(d, zoneGroup)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterZoneGroup, zoneGroupName, err))
	}

	return nil
}

func setGRZoneGroupToResourceData(d *schema.ResourceData, zoneGroup *globalrouter.ZoneGroup) error {
	d.SetId(zoneGroup.ID)
	d.Set("name", zoneGroup.Name)
	d.Set("description", zoneGroup.Description)
	d.Set("created_at", zoneGroup.CreatedAt)
	d.Set("updated_at", zoneGroup.UpdatedAt)

	return nil
}
