package selvpc

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
)

func resourceResellSubnetV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellSubnetV2Create,
		Read:   resourceResellSubnetV2Read,
		Delete: resourceResellSubnetV2Delete,
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
			"prefix_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      29,
				ValidateFunc: validation.IntBetween(24, 29),
			},
			"ip_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ipv4",
				ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
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

func resourceResellSubnetV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID := d.Get("project_id").(string)
	opts := subnets.SubnetOpts{
		Subnets: []subnets.SubnetOpt{
			{
				Region:       d.Get("region").(string),
				Quantity:     1,
				Type:         selvpcclient.IPVersion(d.Get("ip_version").(string)),
				PrefixLength: d.Get("prefix_length").(int),
			},
		},
	}

	log.Printf("[DEBUG] Creating subnet with options: %v\n", opts)
	subnetsResponse, _, err := subnets.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return err
	}

	if len(subnetsResponse) != 1 {
		return errors.New("can't get subnets from the response")
	}

	d.SetId(strconv.Itoa(subnetsResponse[0].ID))

	return resourceResellSubnetV2Read(d, meta)
}

func resourceResellSubnetV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Getting subnet %s", d.Id())
	subnet, response, err := subnets.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("cidr", subnet.CIDR)
	d.Set("network_id", subnet.NetworkID)
	d.Set("subnet_id", subnet.SubnetID)
	d.Set("project_id", subnet.ProjectID)
	d.Set("region", subnet.Region)
	d.Set("status", subnet.Status)

	associatedServers := serversMapsFromStructs(subnet.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Printf("[DEBUG] servers: %s", err)
	}

	return nil
}

func resourceResellSubnetV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Deleting subnet %s\n", d.Id())
	response, err := subnets.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	return nil
}
