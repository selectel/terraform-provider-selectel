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
)

const (
	networkingSubnetV2IPVersion4 = 4
	networkingSubnetV2IPVersion6 = 6
)

// networkingSubnetV2IPv6Modes are the IPv6 address and router advertisement
// modes Neutron accepts.
var networkingSubnetV2IPv6Modes = []string{"slaac", "dhcpv6-stateful", "dhcpv6-stateless"}

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
			// The subnet always carries its own CIDR: subnet pools are not part of
			// this provider, so there is no other source for one.
			"cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateNetworkingSubnetV2CIDR,
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
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"gateway_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"no_gateway"},
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
				Default:  networkingSubnetV2IPVersion4,
				ValidateFunc: validation.IntInSlice([]int{
					networkingSubnetV2IPVersion4,
					networkingSubnetV2IPVersion6,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv6_address_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkingSubnetV2IPv6Modes, false),
			},
			"ipv6_ra_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkingSubnetV2IPv6Modes, false),
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
			// service_types and blocked are writable only with admin
			// privileges, so they are read-only here.
			"service_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceVPCNetworkSubnetV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Everything that can be rejected without talking to the API is rejected
	// first, so a bad configuration never changes the remote state.
	if err := validateNetworkingSubnetV2Config(d); err != nil {
		return diag.FromErr(err)
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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
		request.CIDR = vpcString(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = vpcString(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOk("allocation_pool"); ok {
		pools := expandNetworkingSubnetV2AllocationPools(v.(*schema.Set).List())
		request.AllocationPools = &pools
	}
	if v, ok := d.GetOk("dns_nameservers"); ok {
		nameservers := expandToStringSlice(v.([]interface{}))
		request.DNSNameservers = &nameservers
	}
	if v, ok := d.GetOk("ipv6_address_mode"); ok {
		request.IPv6AddressMode = vpcString(v.(string))
	}
	if v, ok := d.GetOk("ipv6_ra_mode"); ok {
		request.IPv6RAMode = vpcString(v.(string))
	}

	// A nil gateway lets the API pick one, an explicit JSON null means the
	// subnet is created without a gateway.
	if d.Get("no_gateway").(bool) {
		request.GatewayIP = vpc.Null[string]()
	} else if v, ok := d.GetOk("gateway_ip"); ok {
		request.GatewayIP = vpc.Value(v.(string))
	}

	log.Print("[DEBUG] creating networking subnet")
	created, err := subnet.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingSubnetV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := subnet.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetV2Object, created.ID, err))
		}
	}

	return resourceVPCNetworkSubnetV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking subnet %s", d.Id())
	found, err := subnet.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingSubnetV2Object)
	}

	flattenNetworkingSubnetV2(d, found)

	return nil
}

func resourceVPCNetworkSubnetV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := validateNetworkingSubnetV2Config(d); err != nil {
		return diag.FromErr(err)
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     subnet.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = vpcString(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("enable_dhcp") {
		enableDHCP := d.Get("enable_dhcp").(bool)
		request.EnableDHCP = &enableDHCP
		needsUpdate = true
	}
	if d.HasChange("allocation_pool") {
		pools := expandNetworkingSubnetV2AllocationPools(d.Get("allocation_pool").(*schema.Set).List())
		request.AllocationPools = &pools
		needsUpdate = true
	}
	if d.HasChange("dns_nameservers") {
		nameservers := expandToStringSlice(d.Get("dns_nameservers").([]interface{}))
		request.DNSNameservers = &nameservers
		needsUpdate = true
	}
	if d.HasChange("gateway_ip") || d.HasChange("no_gateway") {
		if d.Get("no_gateway").(bool) {
			request.GatewayIP = vpc.Null[string]()
		} else {
			request.GatewayIP = vpc.Value(d.Get("gateway_ip").(string))
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking subnet %s", d.Id())
		if _, err := subnet.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := subnet.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetV2Object, d.Id(), err))
		}
	}

	return resourceVPCNetworkSubnetV2Read(ctx, d, meta)
}

func resourceVPCNetworkSubnetV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking subnet %s", d.Id())
	if err := subnet.Delete(ctx, client, d.Id()); err != nil {
		return networkingCheckDeletedOnDelete(d, err, networkingSubnetV2Object, func() error {
			_, confirmErr := subnet.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceVPCNetworkSubnetV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// flattenNetworkingSubnetV2 writes an API subnet into the resource state.
func flattenNetworkingSubnetV2(d *schema.ResourceData, found *subnet.Subnet) {
	d.Set("network_id", found.NetworkID)
	d.Set("cidr", found.CIDR)
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("ip_version", found.IPVersion)
	d.Set("gateway_ip", derefString(found.GatewayIP))
	d.Set("no_gateway", derefString(found.GatewayIP) == "")
	d.Set("enable_dhcp", found.EnableDHCP)
	d.Set("ipv6_address_mode", derefString(found.IPv6AddressMode))
	d.Set("ipv6_ra_mode", derefString(found.IPv6RAMode))
	d.Set("blocked", found.Blocked)

	if err := d.Set("allocation_pool", flattenNetworkingSubnetV2AllocationPools(found.AllocationPools)); err != nil {
		log.Print(errSettingComplexAttr("allocation_pool", err))
	}
	if err := d.Set("dns_nameservers", found.DNSNameservers); err != nil {
		log.Print(errSettingComplexAttr("dns_nameservers", err))
	}
	if err := d.Set("service_types", found.ServiceTypes); err != nil {
		log.Print(errSettingComplexAttr("service_types", err))
	}
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}

// validateNetworkingSubnetV2CIDR rejects a value that is not a CIDR or that is
// not the network address of its own prefix.
func validateNetworkingSubnetV2CIDR(v interface{}, k string) ([]string, []error) {
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

// validateNetworkingSubnetV2Config rejects the configuration errors that the
// schema itself cannot express, before any API call happens.
func validateNetworkingSubnetV2Config(d *schema.ResourceData) error {
	nameservers := expandToStringSlice(d.Get("dns_nameservers").([]interface{}))
	seen := make(map[string]struct{}, len(nameservers))
	for _, nameserver := range nameservers {
		if _, duplicate := seen[nameserver]; duplicate {
			return fmt.Errorf("dns_nameservers must not contain duplicate entries: %s", nameserver)
		}
		seen[nameserver] = struct{}{}
	}

	if d.Get("ip_version").(int) == networkingSubnetV2IPVersion4 {
		if d.Get("ipv6_address_mode").(string) != "" || d.Get("ipv6_ra_mode").(string) != "" {
			return fmt.Errorf("ipv6_address_mode and ipv6_ra_mode require ip_version to be %d",
				networkingSubnetV2IPVersion6)
		}
	}

	return nil
}

func expandNetworkingSubnetV2AllocationPools(rawPools []interface{}) []subnet.AllocationPool {
	pools := make([]subnet.AllocationPool, 0, len(rawPools))

	for _, rawPool := range rawPools {
		poolMap, ok := rawPool.(map[string]interface{})
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

func flattenNetworkingSubnetV2AllocationPools(pools []subnet.AllocationPool) []interface{} {
	result := make([]interface{}, 0, len(pools))

	for _, pool := range pools {
		result = append(result, map[string]interface{}{
			"start": pool.Start,
			"end":   pool.End,
		})
	}

	return result
}

func flattenNetworkingSubnetV2HostRoutes(routes []subnet.HostRoute) []interface{} {
	result := make([]interface{}, 0, len(routes))

	for _, route := range routes {
		result = append(result, map[string]interface{}{
			"destination_cidr": route.Destination,
			"next_hop":         route.NextHop,
		})
	}

	return result
}
