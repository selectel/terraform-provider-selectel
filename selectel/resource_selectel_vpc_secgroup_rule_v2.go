package selectel

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

const vpcSecGroupRuleV2Object = "VPC security group rule"

var secGroupRuleDocs = resourceDocs{Name: "security group rule"}

func resourceVPCSecGroupRuleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCSecGroupRuleV2Create,
		ReadContext:   resourceVPCSecGroupRuleV2Read,
		DeleteContext: resourceVPCSecGroupRuleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCSecGroupRuleV2ImportState,
		},
		Description: "Creates and manages a rule of a security group using public API v2. " +
			"For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).\n\n" +
			upgradeGuideNote,
		Schema: secGroupRuleDocs.withDocsHints(map[string]*schema.Schema{
			"id":         secGroupRuleDocs.idResourceSchema(),
			"project_id": projectIDResourceSchema(),
			"region":     secGroupRuleDocs.regionResourceSchema(),
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the security group the rule belongs to. " +
					"Retrieved from the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) resource.",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ingress", "egress"}, false),
				Description:  "Direction of the traffic. Available values are `ingress` or `egress`.",
			},
			"ethertype": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, false),
				Description:  "Internet protocol version of the rule. Available values are `IPv4` or `IPv6`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Description of the rule.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVPCSecGroupRuleV2Protocol,
				Description: "Protocol of the rule. " +
					"As an alternative, you can specify an IP protocol number from `0` to `255` as a string. " +
					"If not specified, the rule matches any protocol. " +
					"Learn more about the security group rules in the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/about-security-groups/).",
			},
			"port_range_min": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol"},
				ValidateFunc: validation.IntBetween(0, 65535),
				Description: "Lower bound of the port range. Available values are from `0` to `65535`. Requires `protocol`. " +
					"With `protocol = \"icmp\"`, specifies the ICMP **type** from `0` to `255` and can be set without `port_range_max`. " +
					"Learn more about the security group rules in the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/about-security-groups/).",
			},
			"port_range_max": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol", "port_range_min"},
				ValidateFunc: validation.IntBetween(0, 65535),
				Description: "Upper bound of the port range. Available values are from `0` to `65535`. " +
					"Requires `protocol` and `port_range_min`. With `protocol = \"icmp\"`, specifies the ICMP **code** from `0` to `255`. " +
					"Learn more about the security group rules in the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/about-security-groups/).",
			},
			"remote_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"remote_ip_prefix"},
				Description: "Unique identifier of the security group whose members the rule matches. " +
					"Conflicts with `remote_ip_prefix`. " +
					"Retrieved from the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/vpc_secgroup_v2) data source.",
			},
			"remote_ip_prefix": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"remote_group_id"},
				DiffSuppressFunc: suppressVPCSecGroupRuleV2PrefixDiff,
				ValidateFunc:     validation.IsCIDR,
				Description: "CIDR the rule matches, for example `192.168.1.0/24`. " +
					"Conflicts with `remote_group_id`.",
			},
		}),
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
	if err := importProjectRegion(d, meta); err != nil {
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
	// Zero is a valid ICMP type and code, so an explicit 0 must reach the API.
	if schemaconfig.IsSet(d, "port_range_min") {
		request.PortRangeMin = new(d.Get("port_range_min").(int))
	}
	if schemaconfig.IsSet(d, "port_range_max") {
		request.PortRangeMax = new(d.Get("port_range_max").(int))
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

func validateVPCSecGroupRuleV2Protocol(value any, key string) ([]string, []error) {
	protocol := value.(string)
	switch protocol {
	case "", "ah", "dccp", "egp", "esp", "gre", "icmp", "igmp", "ipip",
		"ipv6-encap", "ipv6-frag", "ipv6-icmp", "ipv6-nonxt", "ipv6-opts",
		"ipv6-route", "ospf", "pgm", "rsvp", "sctp", "tcp", "udp", "udplite", "vrrp":
		return nil, nil
	}

	protocolNumber, err := strconv.Atoi(protocol)
	if err != nil {
		return nil, []error{fmt.Errorf("unknown %q %s for selectel_vpc_secgroup_rule_v2: %w", value, key, err)}
	}
	if protocolNumber < 0 || protocolNumber > 255 {
		return nil, []error{fmt.Errorf("unknown %q %s for selectel_vpc_secgroup_rule_v2", value, key)}
	}

	return nil, nil
}
