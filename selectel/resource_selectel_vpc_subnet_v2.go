package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/clients"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/subnets"
)

func resourceVPCSubnetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCSubnetV2Create,
		ReadContext:   resourceVPCSubnetV2Read,
		DeleteContext: resourceVPCSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCSubnetV2ImportState,
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

func resourceVPCSubnetV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for subnet object: %w", err))
	}

	region := d.Get("region").(string)
	err = validateRegion(selvpcClient, clients.ResellServiceType, region)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	opts := subnets.SubnetOpts{
		Subnets: []subnets.SubnetOpt{
			{
				Region:       region,
				Quantity:     1,
				Type:         selvpcclient.IPVersion(d.Get("ip_version").(string)),
				PrefixLength: d.Get("prefix_length").(int),
			},
		},
	}

	log.Print(msgCreate(objectSubnet, opts))
	subnetsResponse, _, err := subnets.Create(selvpcClient, projectID, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectSubnet, err))
	}
	if len(subnetsResponse) != 1 {
		return diag.FromErr(errReadFromResponse(objectSubnet))
	}

	d.SetId(strconv.Itoa(subnetsResponse[0].ID))

	return resourceVPCSubnetV2Read(ctx, d, meta)
}

func resourceVPCSubnetV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for subnet object: %w", err))
	}

	log.Print(msgGet(objectSubnet, d.Id()))
	subnet, response, err := subnets.Get(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectSubnet, d.Id(), err))
	}

	d.Set("cidr", subnet.CIDR)
	d.Set("network_id", subnet.NetworkID)
	d.Set("subnet_id", subnet.SubnetID)
	d.Set("project_id", subnet.ProjectID)
	d.Set("region", subnet.Region)
	d.Set("status", subnet.Status)

	prefixLength, err := getPrefixLengthFromCIDR(subnet.CIDR)
	if err != nil {
		log.Print(errParsingPrefixLength(objectSubnet, d.Id(), err))
	} else {
		d.Set("prefix_length", prefixLength)
	}

	d.Set("ip_version", getIPVersionFromPrefixLength(prefixLength))

	associatedServers := serversMapsFromStructs(subnet.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	return nil
}

func resourceVPCSubnetV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for subnet object: %w", err))
	}

	log.Print(msgDelete(objectSubnet, d.Id()))
	response, err := subnets.Delete(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectSubnet, d.Id(), err))
	}

	return nil
}

func resourceVPCSubnetV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, fmt.Errorf("INFRA_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
