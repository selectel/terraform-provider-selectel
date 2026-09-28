package selectel

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

const vpcFloatingIPAssociateV2Object = "VPC floating ip association"

func resourceVPCFloatingIPAssociateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCFloatingIPAssociateV2Create,
		ReadContext:   resourceVPCFloatingIPAssociateV2Read,
		UpdateContext: resourceVPCFloatingIPAssociateV2Update,
		DeleteContext: resourceVPCFloatingIPAssociateV2Delete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(vpcV2DefaultStatusTimeout),
			Update: schema.DefaultTimeout(vpcV2DefaultStatusTimeout),
		},
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			"port_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fixed_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	}
}

func resourceVPCFloatingIPAssociateV2Create(
	ctx context.Context, d *schema.ResourceData, meta any,
) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
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

	floatingIPID, err := vpcFloatingIPAssociateV2Result(found, address)
	if err != nil {
		return diag.FromErr(err)
	}

	request := floatingip.UpdateRequest{
		PortID: vpc.Value(d.Get("port_id").(string)),
	}
	if fixedIP, ok := d.GetOk("fixed_ip"); ok {
		request.FixedIPAddress = vpc.Value(fixedIP.(string))
	}

	log.Printf("[DEBUG] associating VPC floating ip %s", floatingIPID)
	if _, err := floatingip.Update(ctx, client, floatingIPID, request); err != nil {
		return diag.FromErr(errUpdatingObject(vpcFloatingIPAssociateV2Object, floatingIPID, err))
	}

	d.SetId(floatingIPID)

	diags := waitForVPCFloatingIPV2Active(ctx, client, floatingIPID,
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutCreate))
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceVPCFloatingIPAssociateV2Read(ctx, d, meta)...)
}

func resourceVPCFloatingIPAssociateV2Read(
	ctx context.Context, d *schema.ResourceData, meta any,
) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC floating ip association %s", d.Id())
	found, err := floatingip.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcFloatingIPAssociateV2Object)
	}

	flattenVPCFloatingIPAssociateV2(d, found)

	return nil
}

func resourceVPCFloatingIPAssociateV2Update(
	ctx context.Context, d *schema.ResourceData, meta any,
) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := floatingip.UpdateRequest{
		PortID: vpc.Value(d.Get("port_id").(string)),
	}

	if d.HasChange("fixed_ip") {
		request.FixedIPAddress = vpc.Value(d.Get("fixed_ip").(string))
	}

	log.Printf("[DEBUG] updating VPC floating ip association %s", d.Id())
	if _, err := floatingip.Update(ctx, client, d.Id(), request); err != nil {
		return diag.FromErr(errUpdatingObject(vpcFloatingIPAssociateV2Object, d.Id(), err))
	}

	diags := waitForVPCFloatingIPV2Active(ctx, client, d.Id(),
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutUpdate))
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceVPCFloatingIPAssociateV2Read(ctx, d, meta)...)
}

func resourceVPCFloatingIPAssociateV2Delete(
	ctx context.Context, d *schema.ResourceData, meta any,
) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] disassociating VPC floating ip %s", d.Id())
	if _, err := floatingip.Update(ctx, client, d.Id(), floatingip.UpdateRequest{
		PortID: vpc.Null[string](),
	}); err != nil {
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return nil
		}

		return diag.FromErr(errUpdatingObject(vpcFloatingIPAssociateV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCFloatingIPAssociateV2ImportState(
	_ context.Context, d *schema.ResourceData, meta any,
) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func vpcFloatingIPAssociateV2Result(
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

func flattenVPCFloatingIPAssociateV2(
	d *schema.ResourceData, found *floatingip.FloatingIP,
) {
	d.Set("floating_ip", found.FloatingIPAddress)
	d.Set("port_id", found.PortID)
	d.Set("fixed_ip", found.FixedIPAddress)
}
