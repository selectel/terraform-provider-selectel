package selvpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  selvpcclient.IPv4,
				ValidateFunc: validation.StringInSlice([]string{
					string(selvpcclient.IPv4),
					string(selvpcclient.IPv6),
				}, false),
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

	prefixLength, err := resourceResellSubnetV2PrefixLengthFromCIDR(subnet.CIDR)
	if err != nil {
		log.Printf("[DEBUG] can't parse prefix length from CIDR: %s", err)
	} else {
		d.Set("prefix_length", prefixLength)
	}

	d.Set("ip_version", resourceResellSubnetV2GetIPVersionFromPrefixLength(prefixLength))

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

func resourceResellSubnetV2PrefixLengthFromCIDR(cidr string) (int, error) {
	cidrParts := strings.Split(cidr, "/")
	if len(cidrParts) != 2 {
		return 0, fmt.Errorf("got invalid CIDR: %s", cidr)
	}

	prefixLenght, err := strconv.Atoi(cidrParts[1])
	if err != nil {
		return 0, fmt.Errorf("error reading prefix length from '%s': %s", cidrParts[1], err)
	}

	return prefixLenght, nil
}

func resourceResellSubnetV2GetIPVersionFromPrefixLength(prefixLength int) string {
	// Any subnet with prefix length larger than 32 is a IPv6 protocol subnet
	// and Selectel doesn't provide any IPv6 subnets with smaller prefix lengths.
	if prefixLength > 32 {
		return string(selvpcclient.IPv6)
	}

	return string(selvpcclient.IPv4)
}
