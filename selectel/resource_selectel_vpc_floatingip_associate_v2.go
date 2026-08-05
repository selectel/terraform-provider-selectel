package selectel

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

const networkingFloatingIPAssociateV2Object = "networking floating ip association"

func resourceVPCFloatingIPAssociateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCFloatingIPAssociateV2Create,
		ReadContext:   resourceVPCFloatingIPAssociateV2Read,
		UpdateContext: resourceVPCFloatingIPAssociateV2Update,
		DeleteContext: resourceVPCFloatingIPAssociateV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCFloatingIPAssociateV2ImportState,
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
			"floating_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceVPCFloatingIPAssociateV2Create(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	address := d.Get("floating_ip").(string)
	found, err := floatingip.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{
			Filters: map[string][]string{"floating_ip_address": {address}},
		},
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding floating ip %q: %w", address, err))
	}

	floatingIPID, err := networkingFloatingIPAssociateV2Result(found, address)
	if err != nil {
		return diag.FromErr(err)
	}

	request := floatingip.UpdateRequest{
		PortID: vpc.Value(d.Get("port_id").(string)),
	}
	if fixedIP, ok := d.GetOk("fixed_ip"); ok {
		request.FixedIPAddress = vpc.Value(fixedIP.(string))
	}

	log.Printf("[DEBUG] associating networking floating ip %s", floatingIPID)
	if _, err := floatingip.Update(ctx, client, floatingIPID, request); err != nil {
		return diag.FromErr(errUpdatingObject(networkingFloatingIPAssociateV2Object, floatingIPID, err))
	}

	d.SetId(floatingIPID)

	return resourceVPCFloatingIPAssociateV2Read(ctx, d, meta)
}

func resourceVPCFloatingIPAssociateV2Read(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking floating ip association %s", d.Id())
	found, err := floatingip.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingFloatingIPAssociateV2Object)
	}

	flattenNetworkingFloatingIPAssociateV2(d, found)

	return nil
}

func resourceVPCFloatingIPAssociateV2Update(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := floatingip.UpdateRequest{
		PortID: vpc.Value(d.Get("port_id").(string)),
	}
	if d.HasChange("fixed_ip") {
		if fixedIP := d.Get("fixed_ip").(string); fixedIP == "" {
			request.FixedIPAddress = vpc.Null[string]()
		} else {
			request.FixedIPAddress = vpc.Value(fixedIP)
		}
	}

	log.Printf("[DEBUG] updating networking floating ip association %s", d.Id())
	if _, err := floatingip.Update(ctx, client, d.Id(), request); err != nil {
		return diag.FromErr(errUpdatingObject(networkingFloatingIPAssociateV2Object, d.Id(), err))
	}

	return resourceVPCFloatingIPAssociateV2Read(ctx, d, meta)
}

func resourceVPCFloatingIPAssociateV2Delete(
	ctx context.Context, d *schema.ResourceData, meta interface{},
) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] disassociating networking floating ip %s", d.Id())
	if _, err := floatingip.Update(ctx, client, d.Id(), floatingip.UpdateRequest{
		PortID: vpc.Null[string](),
	}); err != nil {
		// Unlike deletion of a managed object, absence of the floating IP is
		// an error here: this resource owns only the association and must not
		// claim that it detached an object that no longer exists.
		return diag.FromErr(errUpdatingObject(networkingFloatingIPAssociateV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCFloatingIPAssociateV2ImportState(
	_ context.Context, d *schema.ResourceData, meta interface{},
) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func networkingFloatingIPAssociateV2Result(
	found []floatingip.FloatingIP, address string,
) (string, error) {
	switch len(found) {
	case 0:
		return "", fmt.Errorf("no floating ip found with address %q", address)
	case 1:
		return found[0].ID, nil
	default:
		return "", fmt.Errorf("more than one floating ip found with address %q", address)
	}
}

func flattenNetworkingFloatingIPAssociateV2(
	d *schema.ResourceData, found *floatingip.FloatingIP,
) {
	d.Set("floating_ip", found.FloatingIPAddress)
	d.Set("port_id", derefString(found.PortID))
	d.Set("fixed_ip", derefString(found.FixedIPAddress))
}
