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

var rbacPolicyDocs = resourceDocs{Name: "RBAC policy"}

func resourceVPCRBACPolicyV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRBACPolicyV2Create,
		ReadContext:   resourceVPCRBACPolicyV2Read,
		UpdateContext: resourceVPCRBACPolicyV2Update,
		DeleteContext: resourceVPCRBACPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRBACPolicyV2ImportState,
		},
		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return rbacPolicyDocs.identitySchemaByID(exampleResourceID, rbacPolicyIDFromOpenStackCLI, networkRegionFromControlPanel)
			},
		},
		Description: "Creates and manages an RBAC policy that grants another project access to a networking object using public API v2. " +
			"For more information about managing network access, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/manage-access/).\n\n" +
			upgradeGuideNote,
		Schema: rbacPolicyDocs.withDocsHints(map[string]*schema.Schema{
			"id":         rbacPolicyDocs.idResourceSchema(),
			"project_id": rbacPolicyDocs.projectIDResourceSchema(),
			"region":     rbacPolicyDocs.regionResourceSchema(),
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcRBACPolicyV2Actions, false),
				Description: "Access granted by the policy. Available values are `access_as_shared` and `access_as_external`. " +
					"`access_as_external` is applicable only when `object_type` is `network`.",
			},
			"object_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpcRBACPolicyV2ObjectTypes, false),
				Description:  "Type of the object to share. Available values are `network` and `security_group`.",
			},
			"object_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier of the object being shared. " +
					"Retrieved from the [selectel_vpc_network_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) or [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) resource.",
			},
			"target_tenant": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Unique identifier of the project the access is granted to. " +
					"Changing this moves the grant from one project to another. " +
					"To get the ID of an existing project, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project.",
			},
		}),
	}
}

func resourceVPCRBACPolicyV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceVPCRBACPolicyV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	setProjectRegionIdentity(d)

	return nil
}

func resourceVPCRBACPolicyV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceVPCRBACPolicyV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceVPCRBACPolicyV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
