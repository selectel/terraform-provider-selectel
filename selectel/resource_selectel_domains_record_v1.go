package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/domains-go/pkg/v1/record"
)

func resourceDomainsRecordV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsRecordV1Create,
		ReadContext:   resourceDomainsRecordV1Read,
		UpdateContext: resourceDomainsRecordV1Update,
		DeleteContext: resourceDomainsRecordV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					TypeRecordA,
					TypeRecordAAAA,
					TypeRecordTXT,
					TypeRecordCNAME,
					TypeRecordNS,
					TypeRecordSOA,
					TypeRecordMX,
					TypeRecordSRV,
				}, false),
			},
			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(60, 604800),
			},
			"content": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"priority", "port", "target"},
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"priority", "weight", "target"},
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"target": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"port", "priority", "weight"},
			},
		},
	}
}

func resourceDomainsRecordV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(int)
	selMutexKV.Lock(strconv.Itoa(domainID))
	defer selMutexKV.Unlock(strconv.Itoa(domainID))

	config := meta.(*Config)
	client := config.domainsV1Client()

	createOpts := &record.CreateOpts{
		Name:     d.Get("name").(string),
		Type:     record.Type(d.Get("type").(string)),
		TTL:      d.Get("ttl").(int),
		Content:  d.Get("content").(string),
		Email:    d.Get("email").(string),
		Priority: getIntPtrOrNil(d.Get("priority")),
		Weight:   getIntPtrOrNil(d.Get("weight")),
		Port:     getIntPtrOrNil(d.Get("port")),
		Target:   d.Get("target").(string),
	}

	log.Print(msgCreate(objectRecord, createOpts))
	recordObj, _, err := record.Create(ctx, client, domainID, createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRecord, err))
	}

	d.SetId(strconv.Itoa(recordObj.ID))

	// The ID must be a combination of the domain and record ID
	// since domain ID is required to retrieve a domain record.
	id := fmt.Sprintf("%d/%d", domainID, recordObj.ID)
	d.SetId(id)

	return resourceDomainsRecordV1Read(ctx, d, meta)
}

func resourceDomainsRecordV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.domainsV1Client()

	domainID, recordID, err := domainsV1ParseDomainRecordIDsPair(d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectRecord, d.Id(), err))
	}

	log.Print(msgGet(objectRecord, d.Id()))

	recordObj, resp, err := record.Get(ctx, client, domainID, recordID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(errGettingObject(objectRecord, d.Id(), err))
	}

	d.Set("name", recordObj.Name)
	d.Set("type", recordObj.Type)
	d.Set("ttl", recordObj.TTL)
	d.Set("content", recordObj.Content)
	d.Set("email", recordObj.Email)
	d.Set("priority", recordObj.Priority)
	d.Set("weight", recordObj.Weight)
	d.Set("port", recordObj.Port)
	d.Set("target", recordObj.Target)

	return nil
}

func resourceDomainsRecordV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID, recordID, err := domainsV1ParseDomainRecordIDsPair(d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectRecord, d.Id(), err))
	}
	selMutexKV.Lock(strconv.Itoa(domainID))
	defer selMutexKV.Unlock(strconv.Itoa(domainID))

	config := meta.(*Config)
	client := config.domainsV1Client()

	if d.HasChanges("name", "content", "email", "ttl", "priority", "weight", "port", "target") {
		updateOpts := &record.UpdateOpts{
			Name:     d.Get("name").(string),
			Type:     record.Type(d.Get("type").(string)),
			TTL:      d.Get("ttl").(int),
			Content:  d.Get("content").(string),
			Email:    d.Get("email").(string),
			Priority: getIntPtrOrNil(d.Get("priority")),
			Weight:   getIntPtrOrNil(d.Get("weight")),
			Port:     getIntPtrOrNil(d.Get("port")),
			Target:   d.Get("target").(string),
		}
		_, _, err = record.Update(ctx, client, domainID, recordID, updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectRecord, d.Id(), err))
		}
	}

	return resourceDomainsRecordV1Read(ctx, d, meta)
}

func resourceDomainsRecordV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID, recordID, err := domainsV1ParseDomainRecordIDsPair(d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectRecord, d.Id(), err))
	}
	selMutexKV.Lock(strconv.Itoa(domainID))
	defer selMutexKV.Unlock(strconv.Itoa(domainID))

	config := meta.(*Config)
	client := config.domainsV1Client()

	log.Print(msgDelete(objectRecord, d.Id()))

	_, err = record.Delete(ctx, client, domainID, recordID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRecord, d.Id(), err))
	}

	return nil
}
