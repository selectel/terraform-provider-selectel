package selvpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/floatingips"
)

func resourceResellFloatingIPV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellFloatingIPV2Create,
		Read:   resourceResellFloatingIPV2Read,
		Delete: resourceResellFloatingIPV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"floating_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"fixed_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"servers": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceResellFloatingIPV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)
	opts := floatingips.FloatingIPOpts{
		FloatingIPs: []floatingips.FloatingIPOpt{
			{
				Region:   region,
				Quantity: 1,
			},
		},
	}

	floatingIPs, _, err := floatingips.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return err
	}

	if len(floatingIPs) != 1 {
		return fmt.Errorf("can't get floating ip from the response")
	}

	d.SetId(floatingIPs[0].ID)

	return resourceResellFloatingIPV2Read(d, meta)
}

func resourceResellFloatingIPV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	floatingIP, _, err := floatingips.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	d.Set("fixed_ip_address", floatingIP.FixedIPAddress)
	d.Set("floating_ip_address", floatingIP.FloatingIPAddress)
	d.Set("port_id", floatingIP.PortID)
	d.Set("project_id", floatingIP.ProjectID)
	d.Set("status", floatingIP.Status)
	// Convert servers to a list of maps.
	associatedServers := make([]map[string]interface{}, len(floatingIP.Servers))
	for i, server := range floatingIP.Servers {
		associatedServers[i] = map[string]interface{}{
			"id":     server.ID,
			"name":   server.Name,
			"status": server.Status,
		}
	}
	d.Set("servers", associatedServers)

	return nil
}

func resourceResellFloatingIPV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	_, err := floatingips.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	return nil
}
