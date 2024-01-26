package selectel

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

func resourceDomainsZoneV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsZoneV2Create,
		ReadContext:   resourceDomainsZoneV2Read,
		DeleteContext: resourceDomainsZoneV2Delete,
		UpdateContext: resourceDomainsZoneV2Update,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainsZoneV2ImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
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
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceDomainsZoneV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneName := d.Get("name").(string)
	createOpts := domainsV2.Zone{
		Name: zoneName,
	}

	log.Println(msgCreate(objectZone, zoneName))

	zone, err := client.CreateZone(ctx, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectZone, err))
	}
	// Update comment after creating
	// because set comment in creating request not supporting
	if v, ok := d.GetOk("comment"); ok {
		comment := v.(string)
		err = client.UpdateZoneComment(ctx, zone.ID, comment)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, zone.ID, err))
		}
	}
	// Update disabled after creating
	// because set disabled in creating request not supporting
	if v, ok := d.GetOk("disabled"); ok {
		disabled := v.(bool)
		err = client.UpdateZoneState(ctx, zone.ID, disabled)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, zone.ID, err))
		}
	}

	err = setZoneToResourceData(d, zone)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectZone, err))
	}

	return nil
}

func resourceDomainsZoneV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneName := d.Get("name").(string)

	log.Println(msgGet(objectZone, zoneName))
	zone, err := getZoneByName(ctx, client, zoneName)
	if err != nil {
		return diag.FromErr(errGettingObject(objectZone, zoneName, err))
	}

	err = setZoneToResourceData(d, zone)
	if err != nil {
		return diag.FromErr(errGettingObject(objectZone, zoneName, err))
	}

	return nil
}

func resourceDomainsZoneV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return nil, err
	}

	zoneName := d.Id()

	log.Println(msgImport(objectZone, zoneName))

	zone, err := getZoneByName(ctx, client, zoneName)
	if err != nil {
		return nil, err
	}

	err = setZoneToResourceData(d, zone)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDomainsZoneV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Id()

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectZone, zoneID, err))
	}

	if d.HasChange("comment") {
		comment := ""
		if v, ok := d.GetOk("comment"); ok {
			comment = v.(string)
		}
		log.Println(msgUpdate(objectZone, zoneID, comment))

		err = client.UpdateZoneComment(ctx, zoneID, comment)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, zoneID, err))
		}
	}

	if d.HasChange("disabled") {
		disabled := false
		if v, ok := d.GetOk("disabled"); ok {
			disabled = v.(bool)
		}
		log.Println(msgUpdate(objectZone, zoneID, disabled))

		err = client.UpdateZoneState(ctx, zoneID, disabled)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, zoneID, err))
		}
	}

	return nil
}

func resourceDomainsZoneV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneID := d.Id()

	log.Println(msgDelete(objectZone, zoneID))

	err = client.DeleteZone(ctx, zoneID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectZone, zoneID, err))
	}

	return nil
}

func setZoneToResourceData(d *schema.ResourceData, zone *domainsV2.Zone) error {
	d.SetId(zone.ID)
	d.Set("name", zone.Name)
	d.Set("comment", zone.Comment)
	d.Set("created_at", zone.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", zone.UpdatedAt.Format(time.RFC3339))
	d.Set("delegation_checked_at", zone.DelegationCheckedAt.Format(time.RFC3339))
	d.Set("last_check_status", zone.LastCheckStatus)
	d.Set("last_delegated_at", zone.LastDelegatedAt.Format(time.RFC3339))
	d.Set("project_id", strings.ReplaceAll(zone.ProjectID, "-", ""))
	d.Set("disabled", zone.Disabled)

	return nil
}
