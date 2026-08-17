package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/rbacpolicy"
)

const vpcRBACPolicyV2Object = "VPC rbac policy"

var (
	vpcRBACPolicyV2Actions = []string{
		"access_as_external",
		"access_as_shared",
	}
	vpcRBACPolicyV2ObjectTypes = []string{
		"network",
		"security_group",
	}
)

func resourceVPCRBACPolicyV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRBACPolicyV2Create,
		ReadContext:   resourceVPCRBACPolicyV2Read,
		UpdateContext: resourceVPCRBACPolicyV2Update,
		DeleteContext: resourceVPCRBACPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRBACPolicyV2ImportState,
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
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcRBACPolicyV2Actions, false),
			},
			"object_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcRBACPolicyV2ObjectTypes, false),
			},
			"object_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_tenant": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceVPCRBACPolicyV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print("[DEBUG] creating VPC rbac policy")
	created, err := rbacpolicy.Create(ctx, client, rbacpolicy.CreateRequest{
		ObjectType:   d.Get("object_type").(string),
		ObjectID:     d.Get("object_id").(string),
		Action:       d.Get("action").(string),
		TargetTenant: d.Get("target_tenant").(string),
	})
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcRBACPolicyV2Object, err))
	}

	d.SetId(created.ID)

	return resourceVPCRBACPolicyV2Read(ctx, d, meta)
}

func resourceVPCRBACPolicyV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC rbac policy %s", d.Id())
	found, err := rbacpolicy.Get(ctx, client, d.Id())
	if err != nil {

		return vpcCheckDeleted(d, err, vpcRBACPolicyV2Object)
	}

	d.Set("action", found.Action)
	d.Set("object_type", found.ObjectType)
	d.Set("object_id", found.ObjectID)
	d.Set("target_tenant", found.TargetTenant)

	return nil
}

func resourceVPCRBACPolicyV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("target_tenant") {
		log.Printf("[DEBUG] updating VPC rbac policy %s", d.Id())
		if _, err := rbacpolicy.Update(ctx, client, d.Id(), rbacpolicy.UpdateRequest{
			TargetTenant: d.Get("target_tenant").(string),
		}); err != nil {
			return diag.FromErr(errUpdatingObject(vpcRBACPolicyV2Object, d.Id(), err))
		}
	}

	return resourceVPCRBACPolicyV2Read(ctx, d, meta)
}

func resourceVPCRBACPolicyV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC rbac policy %s", d.Id())
	if err := rbacpolicy.Delete(ctx, client, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcRBACPolicyV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCRBACPolicyV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
