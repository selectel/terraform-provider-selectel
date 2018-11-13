package selvpc

import (
	"context"
	"errors"
	"log"
	"net/http"

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

func resourceResellFloatingIPV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID := d.Get("project_id").(string)
	opts := floatingips.FloatingIPOpts{
		FloatingIPs: []floatingips.FloatingIPOpt{
			{
				Region:   d.Get("region").(string),
				Quantity: 1,
			},
		},
	}

	log.Printf("[DEBUG] Creating floating ip with options: %v\n", opts)
	floatingIPs, _, err := floatingips.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return err
	}

	if len(floatingIPs) != 1 {
		return errors.New("can't get floating ip from the response")
	}

	d.SetId(floatingIPs[0].ID)

	return resourceResellFloatingIPV2Read(d, meta)
}

func resourceResellFloatingIPV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Getting floating ip %s", d.Id())
	floatingIP, response, err := floatingips.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("fixed_ip_address", floatingIP.FixedIPAddress)
	d.Set("floating_ip_address", floatingIP.FloatingIPAddress)
	d.Set("port_id", floatingIP.PortID)
	d.Set("project_id", floatingIP.ProjectID)
	d.Set("region", floatingIP.Region)
	d.Set("status", floatingIP.Status)

	associatedServers := serversMapsFromStructs(floatingIP.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Printf("[DEBUG] servers: %s", err)
	}

	return nil
}

func resourceResellFloatingIPV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Deleting floating ip %s\n", d.Id())
	response, err := floatingips.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	return nil
}
