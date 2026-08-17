package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/pointer"
)

const (
	vpcPortForwardingV2Object      = "VPC port forwarding"
	vpcPortForwardingV2IDSeparator = "/"
)

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

func resourceVPCPortForwardingV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	floatingIPID := d.Get("floatingip_id").(string)

	request := floatingip.PortForwardingCreateRequest{
		Protocol:          pointer.To(d.Get("protocol").(string)),
		InternalPortID:    pointer.To(d.Get("internal_port_id").(string)),
		InternalIPAddress: pointer.To(d.Get("internal_ip_address").(string)),
		InternalPort:      pointer.To(d.Get("internal_port").(int)),
		ExternalPort:      pointer.To(d.Get("external_port").(int)),
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = pointer.To(v.(string))
	}

	log.Printf("[DEBUG] creating port forwarding on floating ip %s", floatingIPID)
	created, err := floatingip.CreatePortForwarding(ctx, client, floatingIPID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcPortForwardingV2Object, err))
	}

	d.SetId(created.ID)

	return resourceVPCPortForwardingV2Read(ctx, d, meta)
}

func resourceVPCPortForwardingV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	d.Set("description", found.Description)

	return nil
}

func resourceVPCPortForwardingV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     floatingip.PortForwardingUpdateRequest
		needsUpdate bool
	)

	if d.HasChange("protocol") {
		request.Protocol = pointer.To(d.Get("protocol").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port_id") {
		request.InternalPortID = pointer.To(d.Get("internal_port_id").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_ip_address") {
		request.InternalIPAddress = pointer.To(d.Get("internal_ip_address").(string))
		needsUpdate = true
	}
	if d.HasChange("internal_port") {
		request.InternalPort = pointer.To(d.Get("internal_port").(int))
		needsUpdate = true
	}
	if d.HasChange("external_port") {
		request.ExternalPort = pointer.To(d.Get("external_port").(int))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = pointer.To(d.Get("description").(string))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating port forwarding %s", d.Id())
		if _, err := floatingip.UpdatePortForwarding(
			ctx, client, d.Get("floatingip_id").(string), d.Id(), request,
		); err != nil {
			return diag.FromErr(errUpdatingObject(vpcPortForwardingV2Object, d.Id(), err))
		}
	}

	return resourceVPCPortForwardingV2Read(ctx, d, meta)
}

func resourceVPCPortForwardingV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceVPCPortForwardingV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
