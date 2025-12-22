package selectel

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

func dataSourceGlobalRouterZoneV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRouterZoneV1Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"visible_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_create": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_update": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_delete": {
				Type:     schema.TypeBool,
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
			"options": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceGlobalRouterZoneV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}
	zoneName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterZone, zoneName, errors.New("'name' should have type string")),
		)
	}
	service, ok := d.Get("service").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterZone, zoneName, errors.New("'service' should have type string")),
		)
	}

	zone, err := getZoneByParams(ctx, client, zoneName, service)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterZone, zoneName, err))
	}

	err = setGRZoneToResourceData(d, zone)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterZone, zoneName, err))
	}

	return nil
}

func setGRZoneToResourceData(d *schema.ResourceData, zone *globalrouter.Zone) error {
	d.SetId(zone.ID)
	d.Set("name", zone.Name)
	d.Set("created_at", zone.CreatedAt)
	d.Set("updated_at", zone.UpdatedAt)
	d.Set("visible_name", zone.VisibleName)
	d.Set("service", zone.Service)
	d.Set("enable", zone.Enable)
	d.Set("allow_create", zone.AllowCreate)
	d.Set("allow_update", zone.AllowUpdate)
	d.Set("allow_delete", zone.AllowDelete)
	d.Set("options", zone.Options)
	d.Set("groups", flattenZoneGroupsV1(zone.ZoneGroups))

	return nil
}
