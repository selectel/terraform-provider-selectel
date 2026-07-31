package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

const (
	networkingPortForwardingV2Object = "networking port forwarding"

	// networkingPortForwardingV2IDSeparator separates the floating IP from the
	// rule in the import identifier.
	networkingPortForwardingV2IDSeparator = "/"
)

func resourceNetworkingPortForwardingV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingPortForwardingV2Create,
		ReadContext:   resourceNetworkingPortForwardingV2Read,
		UpdateContext: resourceNetworkingPortForwardingV2Update,
		DeleteContext: resourceNetworkingPortForwardingV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingPortForwardingV2ImportState,
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
			// The rule lives under one floating IP, so pointing it at another
			// one replaces the resource.
			"floatingip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// The remaining attributes are updated in place. Their values are
			// validated by the API, not by this schema.
			"internal_port_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"internal_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"internal_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"external_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceNetworkingPortForwardingV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	request := floatingip.PortForwardingCreateRequest{
		Protocol:          vpcString(d.Get("protocol").(string)),
		InternalPortID:    vpcString(d.Get("internal_port_id").(string)),
		InternalIPAddress: vpcString(d.Get("internal_ip_address").(string)),
		InternalPort:      vpcInt(d.Get("internal_port").(int)),
		ExternalPort:      vpcInt(d.Get("external_port").(int)),
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}

	log.Printf("[DEBUG] creating port forwarding on floating ip %s", floatingIPID)
	created, err := floatingip.CreatePortForwarding(ctx, client, floatingIPID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingPortForwardingV2Object, err))
	}

	d.SetId(created.ID)

	return resourceNetworkingPortForwardingV2Read(ctx, d, meta)
}

func resourceNetworkingPortForwardingV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	log.Printf("[DEBUG] getting port forwarding %s on floating ip %s", d.Id(), floatingIPID)
	found, err := floatingip.GetPortForwarding(ctx, client, floatingIPID, d.Id())
	if err != nil {
		// A not-found covers both cases the requirement distinguishes in prose
		// but not in outcome: the rule was removed, or its floating IP was, and
		// the rule cannot exist without it. A denied access or a transport
		// failure keeps the resource in the state instead.
		return networkingCheckDeleted(d, err, networkingPortForwardingV2Object)
	}

	d.Set("floatingip_id", floatingIPID)
	d.Set("protocol", found.Protocol)
	d.Set("internal_port_id", found.InternalPortID)
	d.Set("internal_ip_address", found.InternalIPAddress)
	d.Set("internal_port", found.InternalPort)
	d.Set("external_port", found.ExternalPort)
	d.Set("description", found.Description)

	return nil
}

func resourceNetworkingPortForwardingV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     floatingip.PortForwardingUpdateRequest
		needsUpdate bool
	)

	if d.HasChange("protocol") {
		request.Protocol = vpcString(d.Get("protocol").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port_id") {
		request.InternalPortID = vpcString(d.Get("internal_port_id").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_ip_address") {
		request.InternalIPAddress = vpcString(d.Get("internal_ip_address").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port") {
		request.InternalPort = vpcInt(d.Get("internal_port").(int))
		needsUpdate = true
	}
	if d.HasChange("external_port") {
		request.ExternalPort = vpcInt(d.Get("external_port").(int))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating port forwarding %s", d.Id())
		if _, err := floatingip.UpdatePortForwarding(
			ctx, client, d.Get("floatingip_id").(string), d.Id(), request,
		); err != nil {
			return diag.FromErr(errUpdatingObject(networkingPortForwardingV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingPortForwardingV2Read(ctx, d, meta)
}

func resourceNetworkingPortForwardingV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	log.Printf("[DEBUG] deleting port forwarding %s on floating ip %s", d.Id(), floatingIPID)
	if err := floatingip.DeletePortForwarding(ctx, client, floatingIPID, d.Id()); err != nil {
		// A rule that is already gone, or whose floating IP is, makes Delete an
		// idempotent success - confirmed by a read, so a policy failure masked
		// as absence is not mistaken for one.
		return networkingCheckDeletedOnDelete(d, err, networkingPortForwardingV2Object, func() error {
			_, confirmErr := floatingip.GetPortForwarding(ctx, client, floatingIPID, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingPortForwardingV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	floatingIPID, ruleID, err := parseNetworkingPortForwardingV2ID(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(ruleID)
	d.Set("floatingip_id", floatingIPID)

	return []*schema.ResourceData{d}, nil
}

// parseNetworkingPortForwardingV2ID splits an import identifier of the form
// "<floatingip_id>/<id>" into exactly two non-empty parts.
func parseNetworkingPortForwardingV2ID(id string) (string, string, error) {
	parts := strings.Split(id, networkingPortForwardingV2IDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(
			"unable to parse %s ID '%s': expected the format <floatingip_id>%s<id>",
			networkingPortForwardingV2Object, id, networkingPortForwardingV2IDSeparator)
	}

	return parts[0], parts[1], nil
}
