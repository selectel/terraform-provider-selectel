package selvpc

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/crossregionsubnets"
)

func resourceResellCrossRegionSubnetV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellCrossRegionSubnetV2Create,
		Read:   resourceResellCrossRegionSubnetV2Read,
		Delete: resourceResellCrossRegionSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Set:      hashRegions,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vlan_id": {
				Type:     schema.TypeInt,
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
			"subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashSubnets,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Computed: true,
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
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vtep_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceResellCrossRegionSubnetV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	regionsOpts, err := expandResellV2CrossRegionOpts(d.Get("regions").(*schema.Set))
	if err != nil {
		return errParseCrossRegionSubnetV2Regions(err)
	}
	opts := crossregionsubnets.CrossRegionSubnetOpts{
		CrossRegionSubnets: []crossregionsubnets.CrossRegionSubnetOpt{
			{
				Quantity: 1,
				Regions:  regionsOpts,
				CIDR:     d.Get("cidr").(string),
			},
		},
	}
	projectID := d.Get("project_id").(string)

	log.Print(msgCreate(objectCrossRegionSubnet, opts))
	crossRegionSubnetsResponse, _, err := crossregionsubnets.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return errCreatingObject(objectCrossRegionSubnet, err)
	}
	if len(crossRegionSubnetsResponse) != 1 {
		return errReadFromResponse(objectCrossRegionSubnet)
	}

	d.SetId(strconv.Itoa(crossRegionSubnetsResponse[0].ID))

	return resourceResellCrossRegionSubnetV2Read(d, meta)
}

func resourceResellCrossRegionSubnetV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgGet(objectCrossRegionSubnet, d.Id()))
	crossRegionSubnet, response, err := crossregionsubnets.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errGettingObject(objectCrossRegionSubnet, d.Id(), err)
	}

	d.Set("cidr", crossRegionSubnet.CIDR)
	d.Set("status", crossRegionSubnet.Status)
	d.Set("vlan_id", crossRegionSubnet.VLANID)

	associatedServers := serversMapsFromStructs(crossRegionSubnet.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	associatedSubnets := subnetsMapsFromStructs(crossRegionSubnet.Subnets)
	if err := d.Set("subnets", associatedSubnets); err != nil {
		log.Print(errSettingComplexAttr("subnets", err))
	}

	associatedRegions := regionsMapsFromSubnetsStructs(crossRegionSubnet.Subnets)
	if err := d.Set("regions", associatedRegions); err != nil {
		log.Print(errSettingComplexAttr("regions", err))
	}

	associatedProjectID, err := projectIDFromSubnetsMaps(associatedSubnets)
	if err != nil {
		log.Print(errParseCrossRegionSubnetV2ProjectID(err))
	}
	if err := d.Set("project_id", associatedProjectID); err != nil {
		log.Print(errSettingComplexAttr("project_id", err))
	}

	return nil
}

func resourceResellCrossRegionSubnetV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgDelete(objectCrossRegionSubnet, d.Id()))
	response, err := crossregionsubnets.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject(objectCrossRegionSubnet, d.Id(), err)
	}

	return nil
}

func projectIDFromSubnetsMaps(associatedSubnets []map[string]interface{}) (interface{}, error) {
	if len(associatedSubnets) == 0 {
		return nil, errors.New("got empty associated subnets")
	}

	var associatedProjectID interface{}
	if projectID, ok := associatedSubnets[0]["project_id"]; ok {
		associatedProjectID = projectID
	} else {
		return nil, errors.New("no project ID key in associated subnet map")
	}

	return associatedProjectID, nil
}
