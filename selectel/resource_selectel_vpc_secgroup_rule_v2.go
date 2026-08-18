package selectel

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
)

const vpcSecGroupRuleV2Object = "VPC security group rule"

func resourceVPCSecGroupRuleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCSecGroupRuleV2Create,
		ReadContext:   resourceVPCSecGroupRuleV2Read,
		DeleteContext: resourceVPCSecGroupRuleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCSecGroupRuleV2ImportState,
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
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ethertype": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"port_range_min": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol"},
			},
			"port_range_max": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol", "port_range_min"},
			},
			"remote_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"remote_ip_prefix"},
			},
			"remote_ip_prefix": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"remote_group_id"},
				DiffSuppressFunc: suppressVPCSecGroupRuleV2PrefixDiff,
			},
		},
	}
}

func resourceVPCSecGroupRuleV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print("[DEBUG] creating VPC security group rule")
	created, err := securitygrouprule.Create(ctx, client, expandVPCSecGroupRuleV2Request(d))
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcSecGroupRuleV2Object, err))
	}

	d.SetId(created.ID)

	return resourceVPCSecGroupRuleV2Read(ctx, d, meta)
}

func resourceVPCSecGroupRuleV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC security group rule %s", d.Id())
	found, err := securitygrouprule.Get(ctx, client, d.Id())
	if err != nil {

		return vpcCheckDeleted(d, err, vpcSecGroupRuleV2Object)
	}

	d.Set("security_group_id", found.SecurityGroupID)
	d.Set("direction", found.Direction)
	d.Set("ethertype", found.EtherType)
	d.Set("description", found.Description)
	d.Set("protocol", found.Protocol)
	d.Set("port_range_min", found.PortRangeMin)
	d.Set("port_range_max", found.PortRangeMax)
	d.Set("remote_group_id", found.RemoteGroupID)
	d.Set("remote_ip_prefix", found.RemoteIPPrefix)

	return nil
}

func resourceVPCSecGroupRuleV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC security group rule %s", d.Id())
	if err := securitygrouprule.Delete(ctx, client, d.Id()); err != nil {
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {

			return nil
		}

		return diag.FromErr(errDeletingObject(vpcSecGroupRuleV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCSecGroupRuleV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func expandVPCSecGroupRuleV2Request(d *schema.ResourceData) securitygrouprule.CreateRequest {
	request := securitygrouprule.CreateRequest{
		SecurityGroupID: d.Get("security_group_id").(string),
		Direction:       d.Get("direction").(string),
		EtherType:       new(d.Get("ethertype").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if v, ok := d.GetOk("protocol"); ok {
		request.Protocol = new(v.(string))
	}
	if v, ok := d.GetOkExists("port_range_min"); ok { //nolint:staticcheck // zero is a valid ICMP type
		request.PortRangeMin = new(v.(int))
	}
	if v, ok := d.GetOkExists("port_range_max"); ok { //nolint:staticcheck // zero is a valid ICMP code
		request.PortRangeMax = new(v.(int))
	}

	switch {
	case d.Get("remote_group_id").(string) != "":
		request.RemoteGroupID = new(d.Get("remote_group_id").(string))
	case d.Get("remote_ip_prefix").(string) != "":
		request.RemoteIPPrefix = new(d.Get("remote_ip_prefix").(string))
	}

	return request
}

// Prevent CIDR drift caused by API-side normalization,
// for example 192.168.1.42/24 -> 192.168.1.0/24.
func suppressVPCSecGroupRuleV2PrefixDiff(_, stored, configured string, _ *schema.ResourceData) bool {
	if stored == configured {
		return true
	}
	if stored == "" || configured == "" {
		return false
	}

	_, storedNet, err := net.ParseCIDR(stored)
	if err != nil {

		return strings.EqualFold(stored, configured)
	}

	_, configuredNet, err := net.ParseCIDR(configured)
	if err != nil {
		return strings.EqualFold(stored, configured)
	}

	return storedNet.String() == configuredNet.String()
}
