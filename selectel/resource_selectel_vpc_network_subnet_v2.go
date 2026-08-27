package selectel

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
)

const (
	vpcNetworkSubnetV2Object = "VPC network subnet"

	vpcNetworkSubnetV2IPVersion4 = 4
	vpcNetworkSubnetV2IPVersion6 = 6
)

var vpcNetworkSubnetV2IPv6Modes = []string{"slaac", "dhcpv6-stateful", "dhcpv6-stateless"}

func resourceVPCNetworkSubnetV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkSubnetV2Create,
		ReadContext:   resourceVPCNetworkSubnetV2Read,
		UpdateContext: resourceVPCNetworkSubnetV2Update,
		DeleteContext: resourceVPCNetworkSubnetV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkSubnetV2ImportState,
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
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateVPCNetworkSubnetV2CIDR,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"allocation_pool": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
					},
				},
			},
			"gateway_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"no_gateway"},
				ValidateFunc:  validation.IsIPAddress,
			},
			"no_gateway": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"gateway_ip"},
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  vpcNetworkSubnetV2IPVersion4,
				ValidateFunc: validation.IntInSlice([]int{
					vpcNetworkSubnetV2IPVersion4,
					vpcNetworkSubnetV2IPVersion6,
				}),
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dns_nameservers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
			},
			"ipv6_address_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcNetworkSubnetV2IPv6Modes, false),
			},
			"ipv6_ra_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcNetworkSubnetV2IPv6Modes, false),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"host_routes": vpcComputedRoutesSchema(),
		},
	}
}

func resourceVPCNetworkSubnetV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	enableDHCP := d.Get("enable_dhcp").(bool)

	request := subnet.CreateRequest{
		NetworkID:  d.Get("network_id").(string),
		IPVersion:  d.Get("ip_version").(int),
		EnableDHCP: &enableDHCP,
	}
	if v, ok := d.GetOk("cidr"); ok {
		request.CIDR = new(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = new(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if v, ok := d.GetOk("allocation_pool"); ok {
		pools := expandVPCNetworkSubnetV2AllocationPools(v.(*schema.Set).List())
		request.AllocationPools = &pools
	}
	if v, ok := d.GetOk("dns_nameservers"); ok {
		nameservers := convert.ToStringSlice(v.([]any))
		request.DNSNameservers = &nameservers
	}
	if v, ok := d.GetOk("ipv6_address_mode"); ok {
		request.IPv6AddressMode = new(v.(string))
	}
	if v, ok := d.GetOk("ipv6_ra_mode"); ok {
		request.IPv6RAMode = new(v.(string))
	}

	if d.Get("no_gateway").(bool) {
		request.GatewayIP = vpc.Null[string]()
	} else if v, ok := d.GetOk("gateway_ip"); ok {
		request.GatewayIP = vpc.Value(v.(string))
	}

	log.Print("[DEBUG] creating VPC network subnet")
	created, err := subnet.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcNetworkSubnetV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := convert.ToStringSlice(tags.(*schema.Set).List())
		if _, err := subnet.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkSubnetV2Object, created.ID, err))
		}
	}

	return resourceVPCNetworkSubnetV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC network subnet %s", d.Id())
	found, err := subnet.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcNetworkSubnetV2Object)
	}

	flattenVPCNetworkSubnetV2(d, found)

	return nil
}

func resourceVPCNetworkSubnetV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     subnet.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = new(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = new(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("enable_dhcp") {
		enableDHCP := d.Get("enable_dhcp").(bool)
		request.EnableDHCP = &enableDHCP
		needsUpdate = true
	}
	if d.HasChange("allocation_pool") {
		pools := expandVPCNetworkSubnetV2AllocationPools(d.Get("allocation_pool").(*schema.Set).List())
		request.AllocationPools = &pools
		needsUpdate = true
	}
	if d.HasChange("dns_nameservers") {
		nameservers := convert.ToStringSlice(d.Get("dns_nameservers").([]any))
		request.DNSNameservers = &nameservers
		needsUpdate = true
	}
	if d.HasChange("gateway_ip") {
		request.GatewayIP = nil
		if v, ok := d.GetOk("gateway_ip"); ok {
			request.GatewayIP = vpc.Value(v.(string))
		}
		needsUpdate = true
	}
	if d.HasChange("no_gateway") && d.Get("no_gateway").(bool) {
		request.GatewayIP = vpc.Null[string]()
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating VPC network subnet %s", d.Id())
		if _, err := subnet.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkSubnetV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandVPCUpdateTagsV2(d)
		if _, err := subnet.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkSubnetV2Object, d.Id(), err))
		}
	}

	return resourceVPCNetworkSubnetV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC network subnet %s", d.Id())
	if err := subnet.Delete(ctx, client, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcNetworkSubnetV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCNetworkSubnetV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func flattenVPCNetworkSubnetV2(d *schema.ResourceData, found *subnet.Subnet) {
	flattenVPCNetworkSubnetV2Common(d, found)
	d.Set("no_gateway", found.GatewayIP == nil)
}

func flattenVPCNetworkSubnetV2Common(d *schema.ResourceData, found *subnet.Subnet) {
	d.Set("network_id", found.NetworkID)
	d.Set("cidr", found.CIDR)
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("ip_version", found.IPVersion)
	d.Set("gateway_ip", found.GatewayIP)
	d.Set("enable_dhcp", found.EnableDHCP)
	d.Set("ipv6_address_mode", found.IPv6AddressMode)
	d.Set("ipv6_ra_mode", found.IPv6RAMode)
	d.Set("blocked", found.Blocked)

	if err := d.Set("allocation_pool", flattenVPCNetworkSubnetV2AllocationPools(found.AllocationPools)); err != nil {
		log.Print(errSettingComplexAttr("allocation_pool", err))
	}
	if err := d.Set("dns_nameservers", found.DNSNameservers); err != nil {
		log.Print(errSettingComplexAttr("dns_nameservers", err))
	}
	if err := d.Set("service_types", found.ServiceTypes); err != nil {
		log.Print(errSettingComplexAttr("service_types", err))
	}
	if err := d.Set("host_routes", flattenVPCNetworkSubnetV2HostRoutes(found.HostRoutes)); err != nil {
		log.Print(errSettingComplexAttr("host_routes", err))
	}
	flattenVPCTagsV2(d, found.Tags)
}

func validateVPCNetworkSubnetV2CIDR(v any, k string) ([]string, []error) {
	value, ok := v.(string)
	if !ok || value == "" {
		return nil, nil
	}

	ip, ipNet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, []error{fmt.Errorf("%q is not a valid CIDR: %w", k, err)}
	}

	if !ip.Equal(ipNet.IP) {
		return nil, []error{fmt.Errorf(
			"%q must be the network address of its prefix: got %s, expected %s", k, value, ipNet.String())}
	}

	return nil, nil
}

func expandVPCNetworkSubnetV2AllocationPools(rawPools []any) []subnet.AllocationPool {
	pools := make([]subnet.AllocationPool, 0, len(rawPools))

	for _, rawPool := range rawPools {
		poolMap, ok := rawPool.(map[string]any)
		if !ok {
			continue
		}

		pools = append(pools, subnet.AllocationPool{
			Start: poolMap["start"].(string),
			End:   poolMap["end"].(string),
		})
	}

	return pools
}

func flattenVPCNetworkSubnetV2AllocationPools(pools []subnet.AllocationPool) []any {
	result := make([]any, 0, len(pools))

	for _, pool := range pools {
		result = append(result, map[string]any{
			"start": pool.Start,
			"end":   pool.End,
		})
	}

	return result
}

func flattenVPCNetworkSubnetV2HostRoutes(routes []subnet.HostRoute) []any {
	result := make([]any, 0, len(routes))

	for _, route := range routes {
		result = append(result, map[string]any{
			"destination_cidr": route.Destination,
			"next_hop":         route.NextHop,
		})
	}

	return result
}
