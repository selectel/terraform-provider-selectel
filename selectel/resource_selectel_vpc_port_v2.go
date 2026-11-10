package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadiff"
)

const vpcPortV2Object = "VPC port"

var portV2Docs = resourceDocs{Name: "port"}

func resourceVPCPortV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPortV2Create,
		ReadContext:   resourceVPCPortV2Read,
		UpdateContext: resourceVPCPortV2Update,
		DeleteContext: resourceVPCPortV2Delete,
		CustomizeDiff: customdiff.All(
			schemadiff.RecomputeOnChange("all_tags", "tags"),
			schemadiff.RecomputeOnChange("all_security_group_ids", "security_group_ids", "no_security_groups"),
		),
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPortV2ImportState,
		},
		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return portV2Docs.identitySchemaByID(exampleResourceID, portIDFromControlPanel, networkRegionFromControlPanel)
			},
		},
		Description: "Creates and manages a port using the v2 public API. " +
			portsMoreInfo + "\n\n" +
			upgradeGuideNote,
		Schema: portV2Docs.withDocsHints(map[string]*schema.Schema{
			"id":         portV2Docs.idResourceSchema(),
			"project_id": portV2Docs.projectIDResourceSchema(),
			"region":     portV2Docs.regionResourceSchema(),
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the network the port belongs to. " +
					networkIDFromResource,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the port.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the port.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Specifies if the port is administratively up.",
			},
			"mac_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsMACAddress,
				Description: "MAC address of the port, for example `fa:16:3e:4c:2c:30`. " +
					"If not specified, it allocates one automatically.",
			},
			"device_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Unique identifier of the device the port is attached to.",
			},
			"device_owner": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Owner of the device the port is attached to.",
			},
			"fixed_ip": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"no_fixed_ip"},
				Description: "Fixed IPs of the port. " +
					"If neither is specified, allocates an address from one of the subnets of the network.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Unique identifier of the subnet the address is allocated from. " +
								"If not specified, selects a suitable subnet of the network. " +
								networkSubnetIDFromResource,
						},
						"ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
							Description: "Specific address to assign, for example `192.168.1.10`. " +
								"If not specified, it allocates one from the subnet.",
						},
					},
				},
			},
			"no_fixed_ip": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"fixed_ip"},
				Description:   "Specifies whether to create the port without any fixed IP.",
			},
			"all_fixed_ips": vpcComputedFixedIPsSchema("All fixed IPs of the port."),
			"security_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"no_security_groups"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of security groups assigned to the port. " +
					"If not specified, the port gets the default security group of the project. " +
					secGroupIDFromResource,
			},
			"no_security_groups": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"security_group_ids"},
				Description:   "Specifies whether to create the port without any security group.",
			},
			"all_security_group_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All unique identifiers of security groups assigned to the port.",
			},
			"allowed_address_pairs": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Additional IP and MAC address pairs allowed on the port.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.Any(validation.IsIPAddress, validation.IsCIDR),
							Description:  "IP address or CIDR of the pair, for example, `192.168.0.10` or `192.168.0.0/24`.",
						},
						"mac_address": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsMACAddress,
							Description: "MAC address of the pair, for example, `00:11:22:33:44:55`. " +
								"If not specified, the pair gets own port MAC address.",
						},
					},
				},
			},
			"extra_dhcp_option": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Extra DHCP options configured on the port.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the option.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value of the option.",
						},
						"ip_version": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  vpcNetworkSubnetV2IPVersion4,
							ValidateFunc: validation.IntInSlice([]int{
								vpcNetworkSubnetV2IPVersion4,
								vpcNetworkSubnetV2IPVersion6,
							}),
							Description: "Internet protocol version of the option. Available versions are `4` or `6`.",
						},
					},
				},
			},
			"dns_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "DNS name of the port. " + privateDNSLearnMore,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the port.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the port.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the port.",
			},
			"port_security_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if port security is enabled.",
			},
			"dhcp_blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if DHCP is administratively blocked on the port.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the port is administratively blocked.",
			},
		}),
	}
}

func resourceVPCPortV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := port.CreateRequest{
		NetworkID:    d.Get("network_id").(string),
		AdminStateUp: new(d.Get("admin_state_up").(bool)),
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = new(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if v, ok := d.GetOk("mac_address"); ok {
		request.MACAddress = new(v.(string))
	}
	if v, ok := d.GetOk("device_id"); ok {
		request.DeviceID = new(v.(string))
	}
	if v, ok := d.GetOk("device_owner"); ok {
		request.DeviceOwner = new(v.(string))
	}
	if v, ok := d.GetOk("dns_name"); ok {
		request.DNSName = new(v.(string))
	}

	if d.Get("no_fixed_ip").(bool) {
		request.FixedIPs = &[]port.FixedIP{}
	} else if v, ok := d.GetOk("fixed_ip"); ok {
		fixedIPs := expandVPCPortV2FixedIPs(v.([]any))
		request.FixedIPs = &fixedIPs
	}

	if d.Get("no_security_groups").(bool) {
		request.SecurityGroups = &[]string{}
	} else if v, ok := d.GetOk("security_group_ids"); ok {
		groups := convert.ToStringSlice(v.(*schema.Set).List())
		request.SecurityGroups = &groups
	}

	if v, ok := d.GetOk("allowed_address_pairs"); ok {
		pairs := expandVPCPortV2AllowedAddressPairs(v.(*schema.Set).List())
		request.AllowedAddressPairs = &pairs
	}
	if v, ok := d.GetOk("extra_dhcp_option"); ok {
		options := expandVPCPortV2ExtraDHCPOptions(v.(*schema.Set).List())
		request.ExtraDHCPOptions = &options
	}
	log.Print("[DEBUG] creating VPC port")
	created, err := port.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcPortV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := convert.ToStringSlice(tags.(*schema.Set).List())
		if _, err := port.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcPortV2Object, created.ID, err))
		}
	}

	return resourceVPCPortV2Read(ctx, d, meta)
}

func resourceVPCPortV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcPortV2Object)
	}

	flattenVPCPortV2(d, found)

	setProjectRegionIdentity(d)

	return nil
}

func resourceVPCPortV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     port.UpdateRequest
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
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = new(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("dns_name") {
		request.DNSName = new(d.Get("dns_name").(string))
		needsUpdate = true
	}
	if d.HasChange("fixed_ip") || d.HasChange("no_fixed_ip") {
		if d.Get("no_fixed_ip").(bool) {
			request.FixedIPs = &[]port.FixedIP{}
		} else {
			fixedIPs := expandVPCPortV2FixedIPs(d.Get("fixed_ip").([]any))
			request.FixedIPs = &fixedIPs
		}
		needsUpdate = true
	}
	if d.HasChange("security_group_ids") || d.HasChange("no_security_groups") {
		if d.Get("no_security_groups").(bool) {
			request.SecurityGroups = &[]string{}
		} else {
			groups := convert.ToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
			request.SecurityGroups = &groups
		}
		needsUpdate = true
	}
	if d.HasChange("allowed_address_pairs") {
		pairs := expandVPCPortV2AllowedAddressPairs(d.Get("allowed_address_pairs").(*schema.Set).List())
		request.AllowedAddressPairs = &pairs
		needsUpdate = true
	}
	if d.HasChange("extra_dhcp_option") {
		before, after := d.GetChange("extra_dhcp_option")
		options := expandVPCPortV2ExtraDHCPOptionsUpdate(
			before.(*schema.Set), after.(*schema.Set))
		request.ExtraDHCPOptions = &options
		needsUpdate = true
	}
	if needsUpdate {
		log.Printf("[DEBUG] updating VPC port %s", d.Id())
		if _, err := port.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcPortV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandVPCUpdateTagsV2(d)
		if _, err := port.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcPortV2Object, d.Id(), err))
		}
	}

	return resourceVPCPortV2Read(ctx, d, meta)
}

func resourceVPCPortV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC port %s", d.Id())
	if err := port.Delete(ctx, client, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcPortV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCPortV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func flattenVPCPortV2(d *schema.ResourceData, found *port.Port) {
	d.Set("network_id", found.NetworkID)
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("admin_state_up", found.AdminStateUp)
	d.Set("mac_address", found.MACAddress)
	d.Set("device_id", found.DeviceID)
	d.Set("device_owner", found.DeviceOwner)
	d.Set("dns_name", found.DNSName)
	d.Set("status", found.Status)
	d.Set("port_security_enabled", found.PortSecurityEnabled)
	d.Set("dhcp_blocked", found.DHCPBlocked)
	d.Set("blocked", found.Blocked)

	if err := d.Set("all_fixed_ips", flattenVPCPortV2FixedIPs(found.FixedIPs)); err != nil {
		log.Print(errSettingComplexAttr("all_fixed_ips", err))
	}
	if err := d.Set("all_security_group_ids", found.SecurityGroups); err != nil {
		log.Print(errSettingComplexAttr("all_security_group_ids", err))
	}
	if err := d.Set("allowed_address_pairs",
		flattenVPCPortV2AllowedAddressPairs(found.AllowedAddressPairs)); err != nil {
		log.Print(errSettingComplexAttr("allowed_address_pairs", err))
	}
	if err := d.Set("extra_dhcp_option",
		flattenVPCPortV2ExtraDHCPOptions(found.ExtraDHCPOptions)); err != nil {
		log.Print(errSettingComplexAttr("extra_dhcp_option", err))
	}
	flattenVPCTagsV2(d, found.Tags)
}

func expandVPCPortV2FixedIPs(raw []any) []port.FixedIP {
	fixedIPs := make([]port.FixedIP, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}

		fixedIPs = append(fixedIPs, port.FixedIP{
			SubnetID:  entry["subnet_id"].(string),
			IPAddress: entry["ip_address"].(string),
		})
	}

	return fixedIPs
}

func flattenVPCPortV2FixedIPs(fixedIPs []port.FixedIP) []any {
	result := make([]any, 0, len(fixedIPs))

	for _, fixedIP := range fixedIPs {
		result = append(result, map[string]any{
			"subnet_id":  fixedIP.SubnetID,
			"ip_address": fixedIP.IPAddress,
		})
	}

	return result
}

func expandVPCPortV2AllowedAddressPairs(raw []any) []port.AllowedAddressPair {
	pairs := make([]port.AllowedAddressPair, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}

		pairs = append(pairs, port.AllowedAddressPair{
			IPAddress:  entry["ip_address"].(string),
			MACAddress: entry["mac_address"].(string),
		})
	}

	return pairs
}

func flattenVPCPortV2AllowedAddressPairs(pairs []port.AllowedAddressPair) []any {
	result := make([]any, 0, len(pairs))

	for _, pair := range pairs {
		result = append(result, map[string]any{
			"ip_address":  pair.IPAddress,
			"mac_address": pair.MACAddress,
		})
	}

	return result
}

func expandVPCPortV2ExtraDHCPOptions(raw []any) []port.ExtraDHCPOption {
	options := make([]port.ExtraDHCPOption, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}

		option := port.ExtraDHCPOption{
			Name:      entry["name"].(string),
			Value:     entry["value"].(string),
			IPVersion: entry["ip_version"].(int),
		}

		options = append(options, option)
	}

	return options
}

// VPC API removes DHCP options only when they are sent with a null value.
func expandVPCPortV2ExtraDHCPOptionsUpdate(before, after *schema.Set) []port.UpdateExtraDHCPOption {
	configured := expandVPCPortV2ExtraDHCPOptions(after.List())

	options := make([]port.UpdateExtraDHCPOption, 0, len(configured))
	kept := make(map[string]struct{}, len(configured))
	for _, option := range configured {
		value := option.Value
		options = append(options, port.UpdateExtraDHCPOption{
			Name:      option.Name,
			Value:     &value,
			IPVersion: new(option.IPVersion),
		})
		kept[option.Name] = struct{}{}
	}

	for _, option := range expandVPCPortV2ExtraDHCPOptions(before.List()) {
		if _, ok := kept[option.Name]; ok {
			continue
		}

		options = append(options, port.UpdateExtraDHCPOption{Name: option.Name})
	}

	return options
}

func flattenVPCPortV2ExtraDHCPOptions(options []port.ExtraDHCPOption) []any {
	result := make([]any, 0, len(options))

	for _, option := range options {
		result = append(result, map[string]any{
			"name":       option.Name,
			"value":      option.Value,
			"ip_version": option.IPVersion,
		})
	}

	return result
}
