package selectel

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/selectel/domains-go/pkg/v1/domain"
)

func resourceDomainsDomainV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainsDomainV1Create,
		Read:   resourceDomainsDomainV1Read,
		Delete: resourceDomainsDomainV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceDomainsDomainV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	client := config.domainsV1Client()

	createOpts := &domain.CreateOpts{
		Name: d.Get("name").(string),
	}

	log.Print(msgCreate(objectDomain, createOpts))
	domainObj, _, err := domain.Create(ctx, client, createOpts)
	if err != nil {
		return errCreatingObject(objectDomain, err)
	}

	d.SetId(strconv.Itoa(domainObj.ID))

	return resourceDomainsDomainV1Read(d, meta)
}

func resourceDomainsDomainV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	client := config.domainsV1Client()

	log.Print(msgGet(objectDomain, d.Id()))

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return errParseDomainsDomainV1ID(d.Id())
	}

	domainObj, resp, err := domain.GetByID(ctx, client, domainID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return errGettingObject(objectDomain, d.Id(), err)
	}

	d.Set("name", domainObj.Name)
	d.Set("user_id", domainObj.UserID)

	return nil
}

func resourceDomainsDomainV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	client := config.domainsV1Client()

	log.Print(msgDelete(objectDomain, d.Id()))

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return errParseDomainsDomainV1ID(d.Id())
	}

	_, err = domain.Delete(ctx, client, domainID)
	if err != nil {
		return errDeletingObject(objectDomain, d.Id(), err)
	}

	return nil
}
