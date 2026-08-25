package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
)

const vpcPortV2Object = "VPC port"

func resourceVPCPortV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPortV2Create,
		ReadContext:   resourceVPCPortV2Read,
		UpdateContext: resourceVPCPortV2Update,
		DeleteContext: resourceVPCPortV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPortV2ImportState,
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"device_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"fixed_ip": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"no_fixed_ip"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"no_fixed_ip": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"fixed_ip"},
			},
			"all_fixed_ips": vpcComputedFixedIPsSchema(),
			"security_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"no_security_groups"},
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
			"no_security_groups": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"security_group_ids"},
			},
			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allowed_address_pairs": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"extra_dhcp_option": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip_version": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  vpcNetworkSubnetV2IPVersion4,
							ValidateFunc: validation.IntInSlice([]int{
								vpcNetworkSubnetV2IPVersion4,
								vpcNetworkSubnetV2IPVersion6,
							}),
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dhcp_blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
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
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
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
