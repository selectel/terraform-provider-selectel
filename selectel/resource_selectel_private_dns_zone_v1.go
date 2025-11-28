package selectel

import (
	"context"
	"fmt"
	"log"

	privatedns "git.selectel.org/bykov.e/private-dns-go/pkg/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	minRecordValues = 1
	maxRecordValues = 100
)

func resourcePrivateDNSZoneV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateDNSZoneV1Create,
		ReadContext:   resourcePrivateDNSZoneV1Read,
		DeleteContext: resourcePrivateDNSZoneV1Delete,
		UpdateContext: resourcePrivateDNSZoneV1Update,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePrivateDNSZoneV1ImportState,
		},
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"serial_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"records": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								TypeRecordA,
								TypeRecordAAAA,
								TypeRecordTXT,
								TypeRecordCNAME,
								TypeRecordMX,
							}, false),
						},
						"ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: minRecordValues,
							MaxItems: maxRecordValues,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"bindings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourcePrivateDNSZoneV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	opts := &privatedns.ZoneCreateDTO{
		Name:   d.Get("domain").(string),
		Domain: d.Get("domain").(string),
	}
	if ttlVal := d.Get("ttl"); ttlVal != nil {
		ttl := ttlVal.(int)
		opts.TTL = &ttl
	}

	recordsFields := d.Get("records")
	if recordsFields != nil {
		records := recordsFields.([]any)
		opts.Records = make([]*privatedns.RecordSetDTO, 0, len(records))
		for _, rec := range records {
			opts.Records = append(opts.Records, objectMapToAddPrivateDNSRecord(rec.(map[string]any)))
		}
	}

	log.Print(msgCreate(objectPrivateDNSZone, opts))

	zone, err := client.CreateZone(ctx, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectPrivateDNSZone, err))
	}

	d.SetId(zone.ID)
	fillPrivateDNSZoneV1Data(zone, d)

	return nil
}

func resourcePrivateDNSZoneV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectPrivateDNSZone, d.Id()))

	zone, err := client.GetZone(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectPrivateDNSZone, d.Id(), err))
	}

	fillPrivateDNSZoneV1Data(zone, d)

	return nil
}

func resourcePrivateDNSZoneV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	opts := privatedns.ZoneUpdateDto{}
	if d.HasChange("ttl") {
		ttlVal := d.Get("ttl")
		if ttlVal != nil {
			ttl := ttlVal.(int)
			opts.TTL = &ttl
		}

		err := client.UpdateZone(ctx, d.Id(), &opts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectPrivateDNSZone, d.Id(), err))
		}
	}

	if d.HasChange("records") {
		recordChanges := processPrivateDNSRecordChanges(d)
		_, err := client.PutRecords(ctx, d.Id(), recordChanges)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectPrivateDNSZone, d.Id(), err))
		}
	}

	return resourcePrivateDNSZoneV1Read(ctx, d, meta)
}

func resourcePrivateDNSZoneV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}
	log.Print(msgDelete(objectPrivateDNSService, d.Id()))
	err := client.DeleteZone(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectPrivateDNSZone, d.Id(), err))
	}

	return nil
}

func resourcePrivateDNSZoneV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, fmt.Errorf("INFRA_PROJECT_ID must be set for the resource import")
	}

	if config.Region == "" {
		return nil, fmt.Errorf("INFRA_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}

func privateDNSRecordKey(v interface{}) string {
	m := v.(map[string]interface{})
	return m["type"].(string) + " " + m["domain"].(string)
}

func fillPrivateDNSZoneV1Data(zone *privatedns.ZoneDetails, d *schema.ResourceData) {
	d.Set("domain", zone.Domain)
	d.Set("serial_number", zone.SerialNumber)
	d.Set("ttl", zone.TTL)

	bindings := []any{}
	for _, binding := range zone.Bindings {
		bindings = append(bindings, map[string]any{
			"resource_id":   binding.ResourceID,
			"resource_type": binding.ResourceType,
		})
	}
	d.Set("bindings", bindings)

	records := []any{}
	for _, record := range zone.Records {
		records = append(records, map[string]any{
			"domain": record.Domain,
			"type":   record.Type,
			"ttl":    record.TTL,
			"values": record.Values,
		})
	}
	d.Set("records", records)
}

func processPrivateDNSRecordChanges(d *schema.ResourceData) *privatedns.PutRecordsDTO {
	oldVal, newVal := d.GetChange("records")
	oldRecords := oldVal.([]any)
	newRecords := newVal.([]any)

	existed := map[string]map[string]any{}
	for _, rec := range oldRecords {
		recObject := rec.(map[string]any)
		existed[privateDNSRecordKey(recObject)] = recObject
	}

	dto := &privatedns.PutRecordsDTO{
		Set: make([]*privatedns.RecordSetDTO, 0),
	}

	for _, rec := range newRecords {
		recObject := rec.(map[string]any)
		dto.Set = append(dto.Set, objectMapToAddPrivateDNSRecord(recObject))
		delete(existed, privateDNSRecordKey(rec))
	}

	if len(existed) > 0 {
		dto.Delete = make([]*privatedns.RecordDeleteDTO, 0, len(existed))
		for _, rec := range existed {
			dto.Delete = append(dto.Delete, &privatedns.RecordDeleteDTO{
				Type:   rec["type"].(string),
				Domain: rec["domain"].(string),
			})
		}
	}

	return dto
}

func objectMapToAddPrivateDNSRecord(recMap map[string]any) *privatedns.RecordSetDTO {
	record := &privatedns.RecordSetDTO{
		Type:   recMap["type"].(string),
		Domain: recMap["domain"].(string),
	}

	ttlField := recMap["ttl"]
	if ttlField != nil {
		ttl := ttlField.(int)
		record.TTL = &ttl
	}

	valuesField := recMap["values"].([]any)
	record.Values = make([]string, 0, len(valuesField))
	for _, value := range valuesField {
		record.Values = append(record.Values, value.(string))
	}

	return record
}
