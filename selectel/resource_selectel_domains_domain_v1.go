package selectel

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/domains-go/pkg/v1/domain"
)

func resourceDomainsDomainV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsDomainV1Create,
		ReadContext:   resourceDomainsDomainV1Read,
		DeleteContext: resourceDomainsDomainV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceDomainsDomainV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createOpts := &domain.CreateOpts{
		Name: d.Get("name").(string),
	}

	log.Print(msgCreate(objectDomain, createOpts))
	domainObj, _, err := domain.Create(ctx, client, createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDomain, err))
	}

	d.SetId(strconv.Itoa(domainObj.ID))

	return resourceDomainsDomainV1Read(ctx, d, meta)
}

func resourceDomainsDomainV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print(msgGet(objectDomain, d.Id()))

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(errParseDomainsDomainV1ID(d.Id()))
	}

	domainObj, resp, err := domain.GetByID(ctx, client, domainID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errGettingObject(objectDomain, d.Id(), err))
	}

	d.Set("name", domainObj.Name)
	d.Set("user_id", domainObj.UserID)

	return nil
}

func resourceDomainsDomainV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print(msgDelete(objectDomain, d.Id()))

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(errParseDomainsDomainV1ID(d.Id()))
	}

	_, err = domain.Delete(ctx, client, domainID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDomain, d.Id(), err))
	}

	return nil
}
