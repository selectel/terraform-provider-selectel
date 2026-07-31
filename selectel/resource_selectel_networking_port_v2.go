package selectel

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/vpc-go/pkg/v2/port"
)

const networkingPortV2Object = "networking port"

// networkingPortV2VNICTypes are the vnic types Neutron accepts.
var networkingPortV2VNICTypes = []string{
	"normal", "direct", "direct-physical", "macvtap",
	"baremetal", "virtio-forwarder", "remote-managed",
}

func resourceNetworkingPortV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingPortV2Create,
		ReadContext:   resourceNetworkingPortV2Read,
		UpdateContext: resourceNetworkingPortV2Update,
		DeleteContext: resourceNetworkingPortV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingPortV2ImportState,
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
				Computed: true,
			},
			// The owner may set mac_address at create, but only admin may
			// change it afterwards, so a change replaces the port.
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
							Required: true,
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
			"all_fixed_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
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
							Default:  networkingSubnetV2IPVersion4,
							ValidateFunc: validation.IntInSlice([]int{
								networkingSubnetV2IPVersion4,
								networkingSubnetV2IPVersion6,
							}),
						},
					},
				},
			},
			// Only vnic_type is a configuration argument: host_id and profile
			// are reserved for admin and are not returned by the SDK either.
			"binding": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vnic_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(networkingPortV2VNICTypes, false),
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
			"dns_assignment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			// Read-only: the cloud's policy grants none of the granular VPC
			// roles permission to change these.
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"qos_policy_id": {
				Type:     schema.TypeString,
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

func resourceNetworkingPortV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := validateNetworkingPortV2Config(d); err != nil {
		return diag.FromErr(err)
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := port.CreateRequest{NetworkID: d.Get("network_id").(string)}
	if v, ok := d.GetOk("name"); ok {
		request.Name = vpcString(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOkExists("admin_state_up"); ok {
		request.AdminStateUp = vpcBool(v.(bool))
	}
	if v, ok := d.GetOk("mac_address"); ok {
		request.MACAddress = vpcString(v.(string))
	}
	if v, ok := d.GetOk("device_id"); ok {
		request.DeviceID = vpcString(v.(string))
	}
	if v, ok := d.GetOk("device_owner"); ok {
		request.DeviceOwner = vpcString(v.(string))
	}
	if v, ok := d.GetOk("dns_name"); ok {
		request.DNSName = vpcString(v.(string))
	}

	// An empty list is how "no fixed IP at all" is expressed, as opposed to
	// omitting the field, which lets the API allocate one.
	if d.Get("no_fixed_ip").(bool) {
		request.FixedIPs = &[]port.FixedIP{}
	} else if v, ok := d.GetOk("fixed_ip"); ok {
		fixedIPs := expandNetworkingPortV2FixedIPs(v.([]interface{}))
		request.FixedIPs = &fixedIPs
	}

	if d.Get("no_security_groups").(bool) {
		request.SecurityGroups = &[]string{}
	} else if v, ok := d.GetOk("security_group_ids"); ok {
		groups := expandToStringSlice(v.(*schema.Set).List())
		request.SecurityGroups = &groups
	}

	if v, ok := d.GetOk("allowed_address_pairs"); ok {
		pairs := expandNetworkingPortV2AllowedAddressPairs(v.(*schema.Set).List())
		request.AllowedAddressPairs = &pairs
	}
	if v, ok := d.GetOk("extra_dhcp_option"); ok {
		options := expandNetworkingPortV2ExtraDHCPOptions(v.(*schema.Set).List())
		request.ExtraDHCPOptions = &options
	}
	if vnicType := networkingPortV2VNICType(d); vnicType != "" {
		request.BindingVNICType = vpcString(vnicType)
	}

	log.Print("[DEBUG] creating networking port")
	created, err := port.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingPortV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := port.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingPortV2Object, created.ID, err))
		}
	}

	return resourceNetworkingPortV2Read(ctx, d, meta)
}

func resourceNetworkingPortV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingPortV2Object)
	}

	flattenNetworkingPortV2(d, found)

	return nil
}

func resourceNetworkingPortV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := validateNetworkingPortV2Config(d); err != nil {
		return diag.FromErr(err)
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     port.UpdateRequest
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
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = vpcBool(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("dns_name") {
		request.DNSName = vpcString(d.Get("dns_name").(string))
		needsUpdate = true
	}
	if d.HasChange("fixed_ip") || d.HasChange("no_fixed_ip") {
		if d.Get("no_fixed_ip").(bool) {
			request.FixedIPs = &[]port.FixedIP{}
		} else {
			fixedIPs := expandNetworkingPortV2FixedIPs(d.Get("fixed_ip").([]interface{}))
			request.FixedIPs = &fixedIPs
		}
		needsUpdate = true
	}
	if d.HasChange("security_group_ids") || d.HasChange("no_security_groups") {
		if d.Get("no_security_groups").(bool) {
			request.SecurityGroups = &[]string{}
		} else {
			groups := expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
			request.SecurityGroups = &groups
		}
		needsUpdate = true
	}
	if d.HasChange("allowed_address_pairs") {
		pairs := expandNetworkingPortV2AllowedAddressPairs(d.Get("allowed_address_pairs").(*schema.Set).List())
		request.AllowedAddressPairs = &pairs
		needsUpdate = true
	}
	if d.HasChange("extra_dhcp_option") {
		options := expandNetworkingPortV2ExtraDHCPOptions(d.Get("extra_dhcp_option").(*schema.Set).List())
		request.ExtraDHCPOptions = &options
		needsUpdate = true
	}
	if d.HasChange("binding") {
		if vnicType := networkingPortV2VNICType(d); vnicType != "" {
			request.BindingVNICType = vpcString(vnicType)
			needsUpdate = true
		}
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking port %s", d.Id())
		if _, err := port.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingPortV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := port.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingPortV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingPortV2Read(ctx, d, meta)
}

func resourceNetworkingPortV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking port %s", d.Id())
	if err := port.Delete(ctx, client, d.Id()); err != nil {
		return networkingCheckDeletedOnDelete(d, err, networkingPortV2Object, func() error {
			_, confirmErr := port.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingPortV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// validateNetworkingPortV2Config rejects what the schema cannot express, before
// any API call happens.
func validateNetworkingPortV2Config(d *schema.ResourceData) error {
	if d.Get("no_security_groups").(bool) {
		if groups, ok := d.GetOk("security_group_ids"); ok && groups.(*schema.Set).Len() > 0 {
			return fmt.Errorf("no_security_groups conflicts with a non-empty security_group_ids")
		}
	}

	if d.Get("no_fixed_ip").(bool) {
		if fixedIPs, ok := d.GetOk("fixed_ip"); ok && len(fixedIPs.([]interface{})) > 0 {
			return fmt.Errorf("no_fixed_ip conflicts with a non-empty fixed_ip")
		}
	}

	return nil
}

// networkingPortV2VNICType reads the single vnic_type out of the binding block.
func networkingPortV2VNICType(d *schema.ResourceData) string {
	raw, ok := d.GetOk("binding")
	if !ok {
		return ""
	}

	blocks, ok := raw.([]interface{})
	if !ok || len(blocks) == 0 {
		return ""
	}

	block, ok := blocks[0].(map[string]interface{})
	if !ok {
		return ""
	}

	vnicType, _ := block["vnic_type"].(string)

	return vnicType
}

// flattenNetworkingPortV2 writes an API port into the state.
func flattenNetworkingPortV2(d *schema.ResourceData, found *port.Port) {
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
	d.Set("qos_policy_id", derefString(found.QoSPolicyID))
	d.Set("dhcp_blocked", found.DHCPBlocked)
	d.Set("blocked", found.Blocked)

	if err := d.Set("all_fixed_ips", flattenNetworkingPortV2FixedIPs(found.FixedIPs)); err != nil {
		log.Print(errSettingComplexAttr("all_fixed_ips", err))
	}
	if err := d.Set("all_security_group_ids", found.SecurityGroups); err != nil {
		log.Print(errSettingComplexAttr("all_security_group_ids", err))
	}
	if err := d.Set("allowed_address_pairs",
		flattenNetworkingPortV2AllowedAddressPairs(found.AllowedAddressPairs)); err != nil {
		log.Print(errSettingComplexAttr("allowed_address_pairs", err))
	}
	if err := d.Set("extra_dhcp_option",
		flattenNetworkingPortV2ExtraDHCPOptions(found.ExtraDHCPOptions)); err != nil {
		log.Print(errSettingComplexAttr("extra_dhcp_option", err))
	}
	if err := d.Set("binding", []interface{}{
		map[string]interface{}{"vnic_type": found.BindingVNICType},
	}); err != nil {
		log.Print(errSettingComplexAttr("binding", err))
	}
	if err := d.Set("dns_assignment",
		flattenNetworkingPortV2DNSAssignment(found.DNSAssignment)); err != nil {
		log.Print(errSettingComplexAttr("dns_assignment", err))
	}
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}

func expandNetworkingPortV2FixedIPs(raw []interface{}) []port.FixedIP {
	fixedIPs := make([]port.FixedIP, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
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

func flattenNetworkingPortV2FixedIPs(fixedIPs []port.FixedIP) []interface{} {
	result := make([]interface{}, 0, len(fixedIPs))

	for _, fixedIP := range fixedIPs {
		result = append(result, map[string]interface{}{
			"subnet_id":  fixedIP.SubnetID,
			"ip_address": fixedIP.IPAddress,
		})
	}

	return result
}

func expandNetworkingPortV2AllowedAddressPairs(raw []interface{}) []port.AllowedAddressPair {
	pairs := make([]port.AllowedAddressPair, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
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

func flattenNetworkingPortV2AllowedAddressPairs(pairs []port.AllowedAddressPair) []interface{} {
	result := make([]interface{}, 0, len(pairs))

	for _, pair := range pairs {
		result = append(result, map[string]interface{}{
			"ip_address":  pair.IPAddress,
			"mac_address": pair.MACAddress,
		})
	}

	return result
}

func expandNetworkingPortV2ExtraDHCPOptions(raw []interface{}) []port.ExtraDHCPOption {
	options := make([]port.ExtraDHCPOption, 0, len(raw))

	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		option := port.ExtraDHCPOption{
			Name:  entry["name"].(string),
			Value: entry["value"].(string),
		}
		if ipVersion, ok := entry["ip_version"].(int); ok && ipVersion != 0 {
			option.IPVersion = vpcInt(ipVersion)
		}

		options = append(options, option)
	}

	return options
}

func flattenNetworkingPortV2ExtraDHCPOptions(options []port.ExtraDHCPOption) []interface{} {
	result := make([]interface{}, 0, len(options))

	for _, option := range options {
		result = append(result, map[string]interface{}{
			"name":       option.Name,
			"value":      option.Value,
			"ip_version": derefInt(option.IPVersion),
		})
	}

	return result
}

func flattenNetworkingPortV2DNSAssignment(assignments []port.DNSAssignment) []interface{} {
	result := make([]interface{}, 0, len(assignments))

	for _, assignment := range assignments {
		result = append(result, map[string]interface{}{
			"ip_address": assignment.IPAddress,
			"hostname":   assignment.Hostname,
			"fqdn":       assignment.FQDN,
		})
	}

	return result
}
