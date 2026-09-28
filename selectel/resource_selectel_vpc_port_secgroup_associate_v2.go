package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadiff"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/sliceutil"
)

const vpcPortSecGroupAssociateV2Object = "VPC port security group association"

func resourceVPCPortSecGroupAssociateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPortSecGroupAssociateV2Create,
		ReadContext:   resourceVPCPortSecGroupAssociateV2Read,
		UpdateContext: resourceVPCPortSecGroupAssociateV2Update,
		DeleteContext: resourceVPCPortSecGroupAssociateV2Delete,
		CustomizeDiff: schemadiff.RecomputeOnChange("all_security_group_ids", "security_group_ids", "enforce"),
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPortSecGroupAssociateV2ImportState,
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
			"port_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enforce": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVPCPortSecGroupAssociateV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	portID := d.Get("port_id").(string)
	desired := d.Get("security_group_ids").(*schema.Set)

	// VPC API replaces the whole list, so serialize updates per port.
	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	groups := convert.ToStringSlice(desired.List())
	if !d.Get("enforce").(bool) {
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(vpcPortV2Object, portID, err))
		}

		groups = sliceutil.Union(found.SecurityGroups, groups)
	}

	log.Printf("[DEBUG] setting %d security group(s) on VPC port %s", len(groups), portID)
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errCreatingObject(vpcPortSecGroupAssociateV2Object, err))
	}

	d.SetId(portID)

	return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
}

func resourceVPCPortSecGroupAssociateV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting security groups of VPC port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcPortV2Object)
	}

	d.Set("port_id", found.ID)
	if err := d.Set("all_security_group_ids", found.SecurityGroups); err != nil {
		log.Print(errSettingComplexAttr("all_security_group_ids", err))
	}

	managed := found.SecurityGroups
	if !d.Get("enforce").(bool) {
		desired := d.Get("security_group_ids").(*schema.Set)
		managed = sliceutil.Intersect(found.SecurityGroups, convert.ToStringSlice(desired.List()))
	}

	if err := d.Set("security_group_ids", managed); err != nil {
		log.Print(errSettingComplexAttr("security_group_ids", err))
	}

	return nil
}

func resourceVPCPortSecGroupAssociateV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.HasChange("security_group_ids") && !d.HasChange("enforce") {
		return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
	}

	portID := d.Get("port_id").(string)
	desired := d.Get("security_group_ids").(*schema.Set)

	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	groups := convert.ToStringSlice(desired.List())
	if !d.Get("enforce").(bool) {
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(vpcPortV2Object, portID, err))
		}

		previous, _ := d.GetChange("security_group_ids")
		unmanaged := sliceutil.Subtract(
			found.SecurityGroups,
			convert.ToStringSlice(previous.(*schema.Set).List()))
		groups = sliceutil.Union(unmanaged, groups)
	}

	log.Printf("[DEBUG] updating security groups of VPC port %s to %d entries", portID, len(groups))
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errUpdatingObject(vpcPortSecGroupAssociateV2Object, portID, err))
	}

	return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
}

func resourceVPCPortSecGroupAssociateV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	portID := d.Get("port_id").(string)

	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	found, err := port.Get(ctx, client, portID)
	if err != nil {
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return nil
		}

		return diag.FromErr(errGettingObject(vpcPortV2Object, portID, err))
	}

	groups := []string{}
	if !d.Get("enforce").(bool) {
		groups = sliceutil.Subtract(
			found.SecurityGroups,
			convert.ToStringSlice(d.Get("security_group_ids").(*schema.Set).List()))
	}

	log.Printf("[DEBUG] releasing security groups of VPC port %s", portID)
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errDeletingObject(vpcPortSecGroupAssociateV2Object, portID, err))
	}

	return nil
}

func resourceVPCPortSecGroupAssociateV2ImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(vpcPortV2Object, d.Id(), err)
	}

	d.Set("port_id", found.ID)

	return []*schema.ResourceData{d}, nil
}
