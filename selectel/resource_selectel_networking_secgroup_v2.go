package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
)

const (
	networkingSecGroupV2Object = "networking security group"

	// networkingSecGroupRuleV2Object names the rule entity in diagnostics; the
	// rule resource itself belongs to its own capability.
	networkingSecGroupRuleV2Object = "networking security group rule"
)

func resourceNetworkingSecGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSecGroupV2Create,
		ReadContext:   resourceNetworkingSecGroupV2Read,
		UpdateContext: resourceNetworkingSecGroupV2Update,
		DeleteContext: resourceNetworkingSecGroupV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingSecGroupV2ImportState,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Optional without Computed: the API default is an empty string, which
			// is the zero value anyway, and Computed would keep the description
			// from being cleared by removing it from the configuration.
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Computed, unlike description: the API default is true, so an unset
			// attribute must take the value the API assigned rather than false.
			"stateful": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			// Whether the rules the API creates with the group are removed. It
			// only takes effect at create time, so changing it replaces the group.
			"delete_default_rules": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Read-only: a group becomes shared through an RBAC policy, which
			// selectel_networking_rbac_policy_v2 manages, not through this group.
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// Identifiers of every rule actually present on the group, whichever
			// resource created it — including rules the Neutron fork adds itself.
			"security_group_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceNetworkingSecGroupV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := securitygroup.CreateRequest{Name: vpcString(d.Get("name").(string))}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOkExists("stateful"); ok { //nolint:staticcheck // GetOkExists distinguishes an explicit false from an unset bool
		request.Stateful = vpcBool(v.(bool))
	}

	log.Print("[DEBUG] creating networking security group")
	created, err := securitygroup.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingSecGroupV2Object, err))
	}

	// The identifier is recorded before the non-atomic steps below. If one of
	// them fails, the group is still tracked and the next apply can finish the
	// job instead of creating a second group.
	d.SetId(created.ID)

	if d.Get("delete_default_rules").(bool) {
		if err := deleteNetworkingSecGroupV2DefaultRules(ctx, client, created); err != nil {
			return diag.FromErr(err)
		}
	}

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := securitygroup.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSecGroupV2Object, created.ID, err))
		}
	}

	return resourceNetworkingSecGroupV2Read(ctx, d, meta)
}

func resourceNetworkingSecGroupV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking security group %s", d.Id())
	found, err := securitygroup.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingSecGroupV2Object)
	}

	flattenNetworkingSecGroupV2(d, found)

	return nil
}

func resourceNetworkingSecGroupV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     securitygroup.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = vpcString(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("stateful") {
		request.Stateful = vpcBool(d.Get("stateful").(bool))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking security group %s", d.Id())
		if _, err := securitygroup.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSecGroupV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := securitygroup.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSecGroupV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingSecGroupV2Read(ctx, d, meta)
}

func resourceNetworkingSecGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking security group %s", d.Id())
	if err := securitygroup.Delete(ctx, client, d.Id()); err != nil {
		// A group still attached to a port is a conflict, and the default group
		// is refused without cloud-administrator privileges. Both stay
		// diagnostic errors that keep the resource in the state.
		return networkingCheckDeletedOnDelete(d, err, networkingSecGroupV2Object, func() error {
			_, confirmErr := securitygroup.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingSecGroupV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// deleteNetworkingSecGroupV2DefaultRules removes the rules the API created
// together with the group. The rules come from the create response, so no extra
// read is needed to find them.
func deleteNetworkingSecGroupV2DefaultRules(
	ctx context.Context, client *vpc.Client, group *securitygroup.SecurityGroup,
) error {
	for _, rule := range group.SecurityGroupRules {
		log.Printf("[DEBUG] deleting default rule %s of networking security group %s", rule.ID, group.ID)
		if err := securitygroup.DeleteRule(ctx, client, rule.ID); err != nil {
			// An already absent rule is not a failure: the outcome asked for is
			// the outcome that holds.
			if networkingIsNotFound(err) {
				continue
			}

			return errDeletingObject(networkingSecGroupRuleV2Object, rule.ID, err)
		}
	}

	return nil
}

// flattenNetworkingSecGroupV2 writes an API security group into the state.
func flattenNetworkingSecGroupV2(d *schema.ResourceData, found *securitygroup.SecurityGroup) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("stateful", found.Stateful)
	d.Set("shared", found.Shared)

	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
	if err := d.Set("security_group_rules", flattenNetworkingSecGroupV2RuleIDs(found.SecurityGroupRules)); err != nil {
		log.Print(errSettingComplexAttr("security_group_rules", err))
	}
}

func flattenNetworkingSecGroupV2RuleIDs(rules []securitygroup.SecurityGroupRule) []string {
	result := make([]string, 0, len(rules))
	for _, rule := range rules {
		result = append(result, rule.ID)
	}

	return result
}
