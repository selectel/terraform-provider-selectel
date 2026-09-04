package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

const (
	vpcPortForwardingV2Object      = "VPC port forwarding"
	vpcPortForwardingV2IDSeparator = "/"
)

var vpcPortForwardingV2Protocols = []string{"tcp", "udp", "icmp", "ipv6-icmp", "sctp", "dccp"}

func resourceVPCPortForwardingV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPortForwardingV2Create,
		ReadContext:   resourceVPCPortForwardingV2Read,
		UpdateContext: resourceVPCPortForwardingV2Update,
		DeleteContext: resourceVPCPortForwardingV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPortForwardingV2ImportState,
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
			"floatingip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internal_port_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"internal_ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"internal_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
				ExactlyOneOf: []string{"internal_port", "internal_port_range"},
			},
			"internal_port_range": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateVPCPortForwardingV2PortRange,
				ExactlyOneOf: []string{"internal_port", "internal_port_range"},
			},
			"external_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
				ExactlyOneOf: []string{"external_port", "external_port_range"},
			},
			"external_port_range": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateVPCPortForwardingV2PortRange,
				ExactlyOneOf: []string{"external_port", "external_port_range"},
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(vpcPortForwardingV2Protocols, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVPCPortForwardingV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	request := floatingip.PortForwardingCreateRequest{
		Protocol:          d.Get("protocol").(string),
		InternalPortID:    d.Get("internal_port_id").(string),
		InternalIPAddress: d.Get("internal_ip_address").(string),
	}
	if v, ok := d.GetOk("internal_port"); ok {
		request.InternalPort = new(v.(int))
	}
	if v, ok := d.GetOk("internal_port_range"); ok {
		request.InternalPortRange = new(v.(string))
	}
	if v, ok := d.GetOk("external_port"); ok {
		request.ExternalPort = new(v.(int))
	}
	if v, ok := d.GetOk("external_port_range"); ok {
		request.ExternalPortRange = new(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}

	log.Printf("[DEBUG] creating port forwarding on floating ip %s", floatingIPID)
	created, err := floatingip.CreatePortForwarding(ctx, client, floatingIPID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcPortForwardingV2Object, err))
	}

	d.SetId(created.ID)

	return resourceVPCPortForwardingV2Read(ctx, d, meta)
}

func resourceVPCPortForwardingV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	log.Printf("[DEBUG] getting port forwarding %s on floating ip %s", d.Id(), floatingIPID)
	found, err := floatingip.GetPortForwarding(ctx, client, floatingIPID, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcPortForwardingV2Object)
	}

	d.Set("floatingip_id", floatingIPID)
	d.Set("protocol", found.Protocol)
	d.Set("internal_port_id", found.InternalPortID)
	d.Set("internal_ip_address", found.InternalIPAddress)
	d.Set("internal_port", found.InternalPort)
	d.Set("external_port", found.ExternalPort)
	d.Set("internal_port_range", found.InternalPortRange)
	d.Set("external_port_range", found.ExternalPortRange)
	d.Set("description", found.Description)

	return nil
}

func resourceVPCPortForwardingV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     floatingip.PortForwardingUpdateRequest
		needsUpdate bool
	)

	if d.HasChange("protocol") {
		request.Protocol = new(d.Get("protocol").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port_id") {
		request.InternalPortID = new(d.Get("internal_port_id").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_ip_address") {
		request.InternalIPAddress = new(d.Get("internal_ip_address").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port") {
		request.InternalPort = new(d.Get("internal_port").(int))
		needsUpdate = true
	}
	if d.HasChange("internal_port_range") {
		request.InternalPortRange = new(d.Get("internal_port_range").(string))
		needsUpdate = true
	}
	if d.HasChange("external_port") {
		request.ExternalPort = new(d.Get("external_port").(int))
		needsUpdate = true
	}
	if d.HasChange("external_port_range") {
		request.ExternalPortRange = new(d.Get("external_port_range").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = new(d.Get("description").(string))
		needsUpdate = true
	}

	if needsUpdate {
		floatingIPID := d.Get("floatingip_id").(string)

		log.Printf("[DEBUG] updating port forwarding %s on floating ip %s", d.Id(), floatingIPID)
		if _, err := floatingip.UpdatePortForwarding(ctx, client, floatingIPID, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcPortForwardingV2Object, d.Id(), err))
		}
	}

	return resourceVPCPortForwardingV2Read(ctx, d, meta)
}

func resourceVPCPortForwardingV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	log.Printf("[DEBUG] deleting port forwarding %s on floating ip %s", d.Id(), floatingIPID)
	if err := floatingip.DeletePortForwarding(ctx, client, floatingIPID, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcPortForwardingV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCPortForwardingV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	floatingIPID, ruleID, err := parseVPCPortForwardingV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(ruleID)
	d.Set("floatingip_id", floatingIPID)

	return []*schema.ResourceData{d}, nil
}

func parseVPCPortForwardingV2ID(id string) (string, string, error) {
	parts := strings.Split(id, vpcPortForwardingV2IDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(
			"unable to parse %s ID '%s': expected the format <floatingip_id>%s<id>",
			vpcPortForwardingV2Object, id, vpcPortForwardingV2IDSeparator)
	}

	return parts[0], parts[1], nil
}

func validateVPCPortForwardingV2PortRange(value any, key string) ([]string, []error) {
	portRange := value.(string)
	if portRange == "" {
		return nil, nil
	}

	bounds := strings.Split(portRange, ":")
	if len(bounds) != 2 {
		return nil, []error{fmt.Errorf(
			"invalid %s %q for selectel_vpc_port_forwarding_v2: expected <first>:<last>",
			key, portRange)}
	}

	ports := make([]int, 0, len(bounds))
	for _, bound := range bounds {
		port, err := strconv.Atoi(bound)
		if err != nil || port < 1 || port > 65535 {
			return nil, []error{fmt.Errorf(
				"invalid %s %q for selectel_vpc_port_forwarding_v2: ports must be between 1 and 65535",
				key, portRange)}
		}
		ports = append(ports, port)
	}
	if ports[0] > ports[1] {
		return nil, []error{fmt.Errorf(
			"invalid %s %q for selectel_vpc_port_forwarding_v2: the first port must not exceed the last one",
			key, portRange)}
	}

	return nil, nil
}
