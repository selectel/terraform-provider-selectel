package selvpc

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/vrrpsubnets"
)

func resourceResellVRRPSubnetV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellVRRPSubnetV2Create,
		Read:   resourceResellVRRPSubnetV2Read,
		Delete: resourceResellVRRPSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"slave_region": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashSubnets,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vtep_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func resourceResellVRRPSubnetV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID := d.Get("project_id").(string)
	opts := vrrpsubnets.VRRPSubnetOpts{
		VRRPSubnets: []vrrpsubnets.VRRPSubnetOpt{
			{
				Quantity: 1,
				Regions: vrrpsubnets.VRRPRegionOpt{
					Master: d.Get("master_region").(string),
					Slave:  d.Get("slave_region").(string),
				},
				Type:         selvpcclient.IPVersion(d.Get("ip_version").(string)),
				PrefixLength: d.Get("prefix_length").(int),
			},
		},
	}

	log.Print(msgCreate(objectVRRPSubnet, opts))
	vrrpSubnetsResponse, _, err := vrrpsubnets.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return errCreatingObject(objectVRRPSubnet, err)
	}
	if len(vrrpSubnetsResponse) != 1 {
		return errReadFromResponse(objectVRRPSubnet)
	}

	d.SetId(strconv.Itoa(vrrpSubnetsResponse[0].ID))

	return resourceResellVRRPSubnetV2Read(d, meta)
}

func resourceResellVRRPSubnetV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgGet(objectVRRPSubnet, d.Id()))
	vrrpSubnet, response, err := vrrpsubnets.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errGettingObject(objectVRRPSubnet, d.Id(), err)
	}

	d.Set("project_id", vrrpSubnet.ProjectID)
	d.Set("master_region", vrrpSubnet.MasterRegion)
	d.Set("slave_region", vrrpSubnet.SlaveRegion)
	d.Set("cidr", vrrpSubnet.CIDR)
	d.Set("status", vrrpSubnet.Status)

	prefixLength, err := getPrefixLengthFromCIDR(vrrpSubnet.CIDR)
	if err != nil {
		log.Print(errParsingPrefixLength(objectVRRPSubnet, d.Id(), err))
	} else {
		d.Set("prefix_length", prefixLength)
	}

	d.Set("ip_version", getIPVersionFromPrefixLength(prefixLength))

	associatedSubnets := subnetsMapsFromStructs(vrrpSubnet.Subnets)
	if err := d.Set("subnets", associatedSubnets); err != nil {
		log.Print(errSettingComplexAttr("subnets", err))
	}

	associatedServers := serversMapsFromStructs(vrrpSubnet.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	return nil
}

func resourceResellVRRPSubnetV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgDelete(objectVRRPSubnet, d.Id()))
	response, err := vrrpsubnets.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject(objectVRRPSubnet, d.Id(), err)
	}

	return nil
}
