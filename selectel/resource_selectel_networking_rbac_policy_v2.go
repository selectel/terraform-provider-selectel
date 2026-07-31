package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/vpc-go/pkg/v2/rbacpolicy"
)

const networkingRBACPolicyV2Object = "networking rbac policy"

// networkingRBACPolicyV2Actions and networkingRBACPolicyV2ObjectTypes are the
// closed sets the upstream Neutron RBAC schema validates against. qos_policy and
// bgpvpn stay in the list even though this provider has no resource for either:
// a policy may reference an existing object of that type by identifier.
var (
	networkingRBACPolicyV2Actions = []string{
		"access_as_external",
		"access_as_shared",
	}
	networkingRBACPolicyV2ObjectTypes = []string{
		"network",
		"qos_policy",
		"security_group",
		"subnetpool",
		"address_scope",
		"address_group",
		"bgpvpn",
	}
)

func resourceNetworkingRBACPolicyV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingRBACPolicyV2Create,
		ReadContext:   resourceNetworkingRBACPolicyV2Read,
		UpdateContext: resourceNetworkingRBACPolicyV2Update,
		DeleteContext: resourceNetworkingRBACPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingRBACPolicyV2ImportState,
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
			// The three attributes below define the policy and its scope, and
			// the API offers no way to change them in place.
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkingRBACPolicyV2Actions, false),
			},
			"object_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkingRBACPolicyV2ObjectTypes, false),
			},
			"object_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// The only attribute the API can update in place.
			"target_tenant": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceNetworkingRBACPolicyV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print("[DEBUG] creating networking rbac policy")
	created, err := rbacpolicy.Create(ctx, client, rbacpolicy.CreateRequest{
		ObjectType:   d.Get("object_type").(string),
		ObjectID:     d.Get("object_id").(string),
		Action:       d.Get("action").(string),
		TargetTenant: d.Get("target_tenant").(string),
	})
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingRBACPolicyV2Object, err))
	}

	d.SetId(created.ID)

	return resourceNetworkingRBACPolicyV2Read(ctx, d, meta)
}

func resourceNetworkingRBACPolicyV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking rbac policy %s", d.Id())
	found, err := rbacpolicy.Get(ctx, client, d.Id())
	if err != nil {
		// Only a not-found drops the resource; an access or transport failure
		// keeps it and reports a diagnostic error.
		return networkingCheckDeleted(d, err, networkingRBACPolicyV2Object)
	}

	d.Set("action", found.Action)
	d.Set("object_type", found.ObjectType)
	d.Set("object_id", found.ObjectID)
	d.Set("target_tenant", found.TargetTenant)

	return nil
}

func resourceNetworkingRBACPolicyV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// target_tenant is the only updatable attribute, so there is nothing else to
	// diff here: every other change is a replacement.
	if d.HasChange("target_tenant") {
		log.Printf("[DEBUG] updating networking rbac policy %s", d.Id())
		if _, err := rbacpolicy.Update(ctx, client, d.Id(), rbacpolicy.UpdateRequest{
			TargetTenant: d.Get("target_tenant").(string),
		}); err != nil {
			return diag.FromErr(errUpdatingObject(networkingRBACPolicyV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingRBACPolicyV2Read(ctx, d, meta)
}

func resourceNetworkingRBACPolicyV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking rbac policy %s", d.Id())
	if err := rbacpolicy.Delete(ctx, client, d.Id()); err != nil {
		// A conflict means the target project still depends on the access, which
		// stays a diagnostic error: an unsuccessful delete must not clear the
		// state as if it had worked.
		return networkingCheckDeletedOnDelete(d, err, networkingRBACPolicyV2Object, func() error {
			_, confirmErr := rbacpolicy.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingRBACPolicyV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
