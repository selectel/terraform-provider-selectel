package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/clients"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/floatingips"
)

func resourceVPCFloatingIPV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCFloatingIPV2Create,
		ReadContext:   resourceVPCFloatingIPV2Read,
		DeleteContext: resourceVPCFloatingIPV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCFloatingIPV2ImportState,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"floating_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fixed_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"servers": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashServers,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVPCFloatingIPV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for floatingip: %w", err))
	}

	region := d.Get("region").(string)
	err = validateRegion(selvpcClient, clients.ResellServiceType, region)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	opts := floatingips.FloatingIPOpts{
		FloatingIPs: []floatingips.FloatingIPOpt{
			{
				Region:   region,
				Quantity: 1,
			},
		},
	}

	log.Print(msgCreate(objectFloatingIP, opts))
	floatingIPs, _, err := floatingips.Create(selvpcClient, projectID, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectFloatingIP, err))
	}
	if len(floatingIPs) != 1 {
		return diag.FromErr(errReadFromResponse(objectFloatingIP))
	}

	d.SetId(floatingIPs[0].ID)

	return resourceVPCFloatingIPV2Read(ctx, d, meta)
}

func resourceVPCFloatingIPV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for floatingips: %w", err))
	}

	log.Print(msgGet(objectFloatingIP, d.Id()))
	floatingIP, response, err := floatingips.Get(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectFloatingIP, d.Id(), err))
	}

	d.Set("fixed_ip_address", floatingIP.FixedIPAddress)
	d.Set("floating_ip_address", floatingIP.FloatingIPAddress)
	d.Set("port_id", floatingIP.PortID)
	d.Set("project_id", floatingIP.ProjectID)
	d.Set("region", floatingIP.Region)
	d.Set("status", floatingIP.Status)

	associatedServers := serversMapsFromStructs(floatingIP.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	return nil
}

func resourceVPCFloatingIPV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for floatingips: %w", err))
	}

	log.Print(msgDelete(objectFloatingIP, d.Id()))
	response, err := floatingips.Delete(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectFloatingIP, d.Id(), err))
	}

	return nil
}

func resourceVPCFloatingIPV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, fmt.Errorf("INFRA_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
