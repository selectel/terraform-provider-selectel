package selectel

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/domains-go/pkg/v1/domain"
)

func dataSourceDomainsDomainV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainsDomainV1Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainsDomainV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getDomainsClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	domainName := d.Get("name").(string)

	log.Print(msgGet(objectDomain, domainName))

	domainObj, _, err := domain.GetByName(ctx, client, domainName)
	if err != nil {
		return diag.FromErr(errGettingObject(objectDomain, domainName, err))
	}

	d.SetId(strconv.Itoa(domainObj.ID))
	d.Set("name", domainObj.Name)
	d.Set("user_id", domainObj.UserID)

	return nil
}
