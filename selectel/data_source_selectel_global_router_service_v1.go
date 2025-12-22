package selectel

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

func dataSourceGlobalRouterServiceV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRouterServiceV1Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"extension": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGlobalRouterServiceV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}
	serviceName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterService, serviceName, errors.New("'name' should have type string")),
		)
	}

	service, err := getServiceByParams(ctx, client, serviceName)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterService, serviceName, err))
	}

	err = setGRServiceToResourceData(d, service)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterService, serviceName, err))
	}

	return nil
}

func setGRServiceToResourceData(d *schema.ResourceData, service *globalrouter.Service) error {
	d.SetId(service.ID)
	d.Set("name", service.Name)
	d.Set("created_at", service.CreatedAt)
	d.Set("extension", service.Extension)

	return nil
}
