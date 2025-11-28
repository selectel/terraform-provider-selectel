package selectel

import (
	"context"
	"fmt"
	"log"

	privatedns "git.selectel.org/bykov.e/private-dns-go/pkg/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePrivateDNSServiceV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateDNSServiceV1Create,
		ReadContext:   resourcePrivateDNSServiceV1Read,
		DeleteContext: resourcePrivateDNSServiceV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePrivateDNSServiceV1ImportState,
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
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"high_availability": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourcePrivateDNSServiceV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	opts := &privatedns.ServiceCreateDTO{NetworkID: d.Get("network_id").(string)}
	log.Print(msgCreate(objectPrivateDNSService, opts))

	service, err := client.CreateService(ctx, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectPrivateDNSService, err))
	}

	d.SetId(service.ID)
	fillPrivateDNSServiceV1Data(service, d)

	return nil
}

func resourcePrivateDNSServiceV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectPrivateDNSService, d.Id()))

	service, err := client.GetService(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectPrivateDNSService, d.Id(), err))
	}

	fillPrivateDNSServiceV1Data(service, d)

	return nil
}

func resourcePrivateDNSServiceV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getPrivateDNSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}
	log.Print(msgDelete(objectPrivateDNSService, d.Id()))
	err := client.DeleteService(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectPrivateDNSService, d.Id(), err))
	}

	return nil
}

func resourcePrivateDNSServiceV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func fillPrivateDNSServiceV1Data(service *privatedns.ServiceDetails, d *schema.ResourceData) {
	d.Set("network_id", service.NetworkID)
	d.Set("high_availability", service.HighAvailability)
	addresses := []any{}
	for _, addr := range service.Addresses {
		addresses = append(addresses, map[string]string{
			"address": addr.Address,
			"cidr":    addr.CIDR,
		})
	}
	d.Set("addresses", addresses)
}
