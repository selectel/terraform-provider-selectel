package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

var ErrRRSetNotFound = errors.New("rrset not found")

func resourceDomainsRRSetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsRRSetV2Create,
		ReadContext:   resourceDomainsRRSetV2Read,
		UpdateContext: resourceDomainsRRSetV2Update,
		DeleteContext: resourceDomainsRRSetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainsRRSetV2ImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
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
			"managed_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"records": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disabled": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceDomainsRRSetV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Get("zone_id").(string)

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	recordType := domainsV2.RecordType(d.Get("type").(string))
	recordsSet := d.Get("records").(*schema.Set)
	records := generateRecordsFromSet(recordsSet)
	createOpts := domainsV2.RRSet{
		Name:    d.Get("name").(string),
		Type:    recordType,
		TTL:     d.Get("ttl").(int),
		ZoneID:  zoneID,
		Records: records,
	}

	if comment := d.Get("comment"); comment != nil {
		createOpts.Comment = comment.(string)
	}
	if managedBy := d.Get("managed_by"); managedBy != nil {
		createOpts.ManagedBy = managedBy.(string)
	}

	rrset, err := client.CreateRRSet(ctx, zoneID, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRRSet, err))
	}

	err = setRRSetToResourceData(d, rrset)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRRSet, err))
	}

	return nil
}

func resourceDomainsRRSetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneID := d.Get("zone_id").(string)
	zoneIDWithRRSetID := fmt.Sprintf("zone_id: %s, rrset_id: %s", zoneID, d.Id())

	log.Print(msgGet(objectRRSet, zoneIDWithRRSetID))

	rrset, err := client.GetRRSet(ctx, zoneID, d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectRRSet, zoneIDWithRRSetID, err))
	}

	err = setRRSetToResourceData(d, rrset)
	if err != nil {
		return diag.FromErr(errGettingObject(objectRRSet, zoneIDWithRRSetID, err))
	}

	return nil
}

func resourceDomainsRRSetV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("SEL_PROJECT_ID must be set for the resource import")
	}
	d.Set("project_id", config.ProjectID)

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return nil, err
	}
	// concat zone_name,rrset_name,rrset_type with symbol "/" instead of rrset id for importing rrset.
	// example: terraform import domains_rrset_v2.<resource_name> <zone_name>/<rrset_name>/<rrset_type>
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, errors.New("id must include three parts: zone_name/rrset_name/rrset_type")
	}

	zoneName := parts[0]
	rrsetName := parts[1]
	rrsetType := parts[2]

	log.Print(msgImport(objectRRSet, fmt.Sprintf("%s/%s/%s", zoneName, rrsetName, rrsetType)))

	zone, err := getZoneByName(ctx, client, zoneName)
	if err != nil {
		return nil, err
	}

	rrset, err := getRRSetByNameAndType(ctx, client, zone.ID, rrsetName, rrsetType)
	if err != nil {
		return nil, err
	}

	err = setRRSetToResourceData(d, rrset)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDomainsRRSetV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Get("zone_id").(string)

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectRRSet, d.Id(), err))
	}

	if d.HasChanges("ttl", "comment", "records") {
		recordsSet := d.Get("records").(*schema.Set)
		records := generateRecordsFromSet(recordsSet)

		updateOpts := domainsV2.RRSet{
			Name:      d.Get("name").(string),
			Type:      domainsV2.RecordType(d.Get("type").(string)),
			TTL:       d.Get("ttl").(int),
			ZoneID:    zoneID,
			ManagedBy: d.Get("managed_by").(string),
			Records:   records,
		}
		if comment, ok := d.GetOk("comment"); ok {
			updateOpts.Comment = comment.(string)
		}
		err = client.UpdateRRSet(ctx, zoneID, d.Id(), &updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectRRSet, d.Id(), err))
		}
	}

	return resourceDomainsRRSetV2Read(ctx, d, meta)
}

func resourceDomainsRRSetV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Get("zone_id").(string)

	client, err := getDomainsV2Client(d, meta)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRRSet, d.Id(), err))
	}

	log.Print(msgDelete(objectRRSet, fmt.Sprintf("zone_id: %s, rrset_id: %s", zoneID, d.Id())))

	err = client.DeleteRRSet(ctx, zoneID, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRRSet, d.Id(), err))
	}

	return nil
}
