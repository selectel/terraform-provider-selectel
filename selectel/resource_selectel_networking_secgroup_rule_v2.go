package selectel

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
)

func resourceNetworkingSecGroupRuleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSecGroupRuleV2Create,
		ReadContext:   resourceNetworkingSecGroupRuleV2Read,
		DeleteContext: resourceNetworkingSecGroupRuleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingSecGroupRuleV2ImportState,
		},
		// A rule has no update path of its own, so every attribute replaces it.
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
			// protocol takes a known name or a numeric IP protocol number as a
			// string; the API is the one that decides which values it knows.
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			// A lone port_range_min is accepted: with an ICMP protocol it is the
			// ICMP type, and the fork validates type and code independently,
			// refusing only a code without a type
			// (neutron/db/securitygroups_db.py, SecurityGroupMissingIcmpType).
			// So a rule for one ICMP type with any code stays expressible.
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
			// The two remote selectors are mutually exclusive.
			"remote_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"remote_ip_prefix"},
			},
			"remote_ip_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"remote_group_id"},
				// Neutron normalises the prefix, so an IPv6 CIDR written in upper
				// case comes back lowered. Comparing the parsed networks keeps
				// that from showing up as a permanent diff.
				DiffSuppressFunc: suppressNetworkingSecGroupRuleV2PrefixDiff,
			},
		},
	}
}

func resourceNetworkingSecGroupRuleV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	remoteGroupID := d.Get("remote_group_id").(string)
	remoteIPPrefix := d.Get("remote_ip_prefix").(string)

	// ConflictsWith rejects the pair at plan time, but only when both are set in
	// the configuration; both being Computed means the check is repeated here
	// against the resolved values.
	if remoteGroupID != "" && remoteIPPrefix != "" {
		return diag.FromErr(errors.New(
			"remote_group_id and remote_ip_prefix are mutually exclusive, please set at most one of them"))
	}

	log.Print("[DEBUG] creating networking security group rule")
	created, err := securitygroup.CreateRule(ctx, client, expandNetworkingSecGroupRuleV2Request(d))
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingSecGroupRuleV2Object, err))
	}

	d.SetId(created.ID)

	return resourceNetworkingSecGroupRuleV2Read(ctx, d, meta)
}

func resourceNetworkingSecGroupRuleV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking security group rule %s", d.Id())
	found, err := securitygroup.GetRule(ctx, client, d.Id())
	if err != nil {
		// Deleting the rule, or the whole group it belongs to, leaves the same
		// confirmed absence. Any other error keeps the resource in the state.
		return networkingCheckDeleted(d, err, networkingSecGroupRuleV2Object)
	}

	d.Set("security_group_id", found.SecurityGroupID)
	d.Set("direction", found.Direction)
	d.Set("ethertype", found.EtherType)
	d.Set("description", found.Description)
	d.Set("protocol", derefString(found.Protocol))
	d.Set("port_range_min", derefInt(found.PortRangeMin))
	d.Set("port_range_max", derefInt(found.PortRangeMax))
	d.Set("remote_group_id", derefString(found.RemoteGroupID))
	d.Set("remote_ip_prefix", derefString(found.RemoteIPPrefix))

	return nil
}

func resourceNetworkingSecGroupRuleV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking security group rule %s", d.Id())
	if err := securitygroup.DeleteRule(ctx, client, d.Id()); err != nil {
		if networkingIsNotFound(err) {
			// The rule, or its group, is already gone: the outcome asked for is
			// the outcome that holds.
			return nil
		}

		return diag.FromErr(errDeletingObject(networkingSecGroupRuleV2Object, d.Id(), err))
	}

	return nil
}

func resourceNetworkingSecGroupRuleV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// expandNetworkingSecGroupRuleV2Request builds the create request from the
// configuration.
//
// port_range_min and port_range_max are read with GetOkExists rather than GetOk
// because zero is a meaningful value for them: with an ICMP protocol they carry
// the ICMP type and code, which Neutron accepts in the range 0..255
// (neutron/db/securitygroups_db.py, _validate_port_range). Dropping a configured
// zero would silently create a rule matching every ICMP type instead of type 0 —
// a wider rule than the configuration asked for.
func expandNetworkingSecGroupRuleV2Request(d *schema.ResourceData) securitygroup.RuleCreateRequest {
	request := securitygroup.RuleCreateRequest{
		SecurityGroupID: d.Get("security_group_id").(string),
		Direction:       d.Get("direction").(string),
		EtherType:       vpcString(d.Get("ethertype").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOk("protocol"); ok {
		request.Protocol = vpcString(v.(string))
	}
	if v, ok := d.GetOkExists("port_range_min"); ok { //nolint:staticcheck // zero is a valid ICMP type
		request.PortRangeMin = vpcInt(v.(int))
	}
	if v, ok := d.GetOkExists("port_range_max"); ok { //nolint:staticcheck // zero is a valid ICMP code
		request.PortRangeMax = vpcInt(v.(int))
	}

	switch {
	case d.Get("remote_group_id").(string) != "":
		request.RemoteSource = securitygroup.ByRemoteGroupID(d.Get("remote_group_id").(string))
	case d.Get("remote_ip_prefix").(string) != "":
		request.RemoteSource = securitygroup.ByRemoteIPPrefix(d.Get("remote_ip_prefix").(string))
	}

	return request
}

// suppressNetworkingSecGroupRuleV2PrefixDiff reports whether two prefixes name
// the same network. Neutron lowercases and canonicalises what it stores, so a
// configuration written differently from the stored form is not a change.
func suppressNetworkingSecGroupRuleV2PrefixDiff(_, stored, configured string, _ *schema.ResourceData) bool {
	if stored == configured {
		return true
	}
	if stored == "" || configured == "" {
		return false
	}

	_, storedNet, err := net.ParseCIDR(stored)
	if err != nil {
		// Not parseable, so fall back to a case-insensitive comparison rather
		// than claiming the values differ.
		return strings.EqualFold(stored, configured)
	}

	_, configuredNet, err := net.ParseCIDR(configured)
	if err != nil {
		return strings.EqualFold(stored, configured)
	}

	return storedNet.String() == configuredNet.String()
}
