package selectel

import (
	"context"
	"errors"
	"log"

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
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("SEL_PROJECT_ID must be set for the resource import")
	}
	d.Set("project_id", config.ProjectID)

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return nil, err
	}

	// use zone name instead of zone id for importing zone.
	// example: terraform import domains_zone_v2.<resource_name> <zone_name>
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
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectZone, d.Id(), err))
	}

	if d.HasChange("comment") {
		comment := ""
		if v, ok := d.GetOk("comment"); ok {
			comment = v.(string)
		}
		log.Println(msgUpdate(objectZone, d.Id(), comment))

		err = client.UpdateZoneComment(ctx, d.Id(), comment)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, d.Id(), err))
		}
	}

	if d.HasChange("disabled") {
		disabled := false
		if v, ok := d.GetOk("disabled"); ok {
			disabled = v.(bool)
		}
		log.Println(msgUpdate(objectZone, d.Id(), disabled))

		err = client.UpdateZoneState(ctx, d.Id(), disabled)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectZone, d.Id(), err))
		}
	}

	return nil
}

func resourceDomainsZoneV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println(msgDelete(objectZone, d.Id()))

	err = client.DeleteZone(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectZone, d.Id(), err))
	}

	return nil
}
