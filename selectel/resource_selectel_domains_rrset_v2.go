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

func resourceDomainsRrsetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsRrsetV2Create,
		ReadContext:   resourceDomainsRrsetV2Read,
		UpdateContext: resourceDomainsRrsetV2Update,
		DeleteContext: resourceDomainsRrsetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainsRrsetV2ImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
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

func resourceDomainsRrsetV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Get("zone_id").(string)
	selMutexKV.Lock(zoneID)
	defer selMutexKV.Unlock(zoneID)

	client, err := getDomainsV2Client(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	recordType := domainsV2.RecordType(d.Get("type").(string))
	recordsSet := d.Get("records").(*schema.Set)
	records := generateRecordsFromSet(recordsSet)
	createOpts := &domainsV2.RRSet{
		Name:     d.Get("name").(string),
		Type:     recordType,
		TTL:      d.Get("ttl").(int),
		ZoneUUID: zoneID,
		Records:  records,
	}

	if comment := d.Get("comment"); comment != nil {
		createOpts.Comment = comment.(string)
	}
	if managedBy := d.Get("managed_by"); managedBy != nil {
		createOpts.ManagedBy = managedBy.(string)
	}

	rrset, err := client.CreateRRSet(ctx, zoneID, createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRrset, err))
	}

	err = setRrsetToResourceData(d, rrset)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRrset, err))
	}

	return nil
}

func resourceDomainsRrsetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsV2Client(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	rrsetID := d.Id()
	zoneID := d.Get("zone_id").(string)
	zoneIDWithRrsetID := fmt.Sprintf("zone_id: %s, rrset_id: %s", zoneID, rrsetID)

	log.Print(msgGet(objectRrset, zoneIDWithRrsetID))

	rrset, err := client.GetRRSet(ctx, zoneID, rrsetID)
	if err != nil {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectRrset, zoneIDWithRrsetID, err))
	}

	err = setRrsetToResourceData(d, rrset)
	if err != nil {
		return diag.FromErr(errGettingObject(objectRrset, zoneIDWithRrsetID, err))
	}

	return nil
}

func resourceDomainsRrsetV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, err := getDomainsV2Client(meta)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, errors.New("id must include three parts: zone_name/rrset_name/rrset_type")
	}

	zoneName := parts[0]
	rrsetName := parts[1]
	rrsetType := parts[2]

	log.Print(msgImport(objectRrset, fmt.Sprintf("%s/%s/%s", zoneName, rrsetName, rrsetType)))

	zone, err := getZoneByName(ctx, client, zoneName)
	if err != nil {
		return nil, err
	}

	rrset, err := getRrsetByNameAndType(ctx, client, zone.UUID, rrsetName, rrsetType)
	if err != nil {
		return nil, err
	}

	err = setRrsetToResourceData(d, rrset)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDomainsRrsetV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rrsetID := d.Id()
	zoneID := d.Get("zone_id").(string)

	selMutexKV.Lock(zoneID)
	defer selMutexKV.Unlock(zoneID)

	client, err := getDomainsV2Client(meta)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectRrset, rrsetID, err))
	}

	if d.HasChanges("ttl", "comment", "records") {
		recordsSet := d.Get("records").(*schema.Set)
		records := generateRecordsFromSet(recordsSet)

		updateOpts := &domainsV2.RRSet{
			Name:      d.Get("name").(string),
			Type:      domainsV2.RecordType(d.Get("type").(string)),
			TTL:       d.Get("ttl").(int),
			ZoneUUID:  zoneID,
			Comment:   d.Get("comment").(string),
			ManagedBy: d.Get("managed_by").(string),
			Records:   records,
		}
		err = client.UpdateRRSet(ctx, zoneID, rrsetID, updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectRrset, rrsetID, err))
		}
	}

	return resourceDomainsRrsetV2Read(ctx, d, meta)
}

func resourceDomainsRrsetV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zoneID := d.Get("zone_id").(string)
	rrsetID := d.Id()
	selMutexKV.Lock(zoneID)
	defer selMutexKV.Unlock(zoneID)

	client, err := getDomainsV2Client(meta)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRrset, rrsetID, err))
	}

	log.Print(msgDelete(objectRrset, fmt.Sprintf("zone_id: %s, rrset_id: %s", zoneID, rrsetID)))

	err = client.DeleteRRSet(ctx, zoneID, rrsetID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRrset, rrsetID, err))
	}

	return nil
}

func setRrsetToResourceData(d *schema.ResourceData, rrset *domainsV2.RRSet) error {
	d.SetId(rrset.UUID)
	d.Set("name", rrset.Name)
	d.Set("comment", rrset.Comment)
	d.Set("managed_by", rrset.ManagedBy)
	d.Set("ttl", rrset.TTL)
	d.Set("type", rrset.Type)
	d.Set("zone_id", rrset.ZoneUUID)
	d.Set("records", generateSetFromRecords(rrset.Records))

	return nil
}

// generateSetFromRecords - generate terraform TypeList from records in rrset.
func generateSetFromRecords(records []domainsV2.RecordItem) []interface{} {
	recordsAsList := []interface{}{}
	for _, record := range records {
		recordsAsList = append(recordsAsList, map[string]interface{}{
			"content":  record.Content,
			"disabled": record.Disabled,
		})
	}

	return recordsAsList
}

// generateRecordsFromSet - generate records for Rrset from terraform TypeList.
func generateRecordsFromSet(recordsSet *schema.Set) []domainsV2.RecordItem {
	records := []domainsV2.RecordItem{}
	for _, recordItem := range recordsSet.List() {
		if record, isOk := recordItem.(map[string]interface{}); isOk {
			records = append(records, domainsV2.RecordItem{
				Content:  record["content"].(string),
				Disabled: record["disabled"].(bool),
			})
		}
	}

	return records
}