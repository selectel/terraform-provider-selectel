package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadiff"
)

const vpcSecGroupV2Object = "VPC security group"

var secGroupDocs = resourceDocs{Name: "security group"}

func resourceVPCSecGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCSecGroupV2Create,
		ReadContext:   resourceVPCSecGroupV2Read,
		UpdateContext: resourceVPCSecGroupV2Update,
		DeleteContext: resourceVPCSecGroupV2Delete,
		CustomizeDiff: schemadiff.RecomputeOnChange("all_tags", "tags"),
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(vpcV2DefaultDeleteTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCSecGroupV2ImportState,
		},
		Description: "Creates and manages a security group using public API v2. " +
			"For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/).\n\n" +
			upgradeGuideNote,
		Schema: secGroupDocs.withDocsHints(map[string]*schema.Schema{
			"id":         secGroupDocs.idResourceSchema(),
			"project_id": projectIDResourceSchema(),
			"region":     secGroupDocs.regionResourceSchema(),
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the security group.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the security group.",
			},
			"stateful": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Specifies if the security group is stateful.",
			},
			"delete_default_rules": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Specifies whether to delete the default egress rules. Default value is `false`.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the security group.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the security group.",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if another project has been granted access to this group.",
			},
			"security_group_rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of rules contained in the security group.",
			},
		}),
	}
}

func resourceVPCSecGroupV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := securitygroup.CreateRequest{Name: new(d.Get("name").(string))}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if schemaconfig.IsSet(d, "stateful") {
		request.Stateful = new(d.Get("stateful").(bool))
	}

	log.Print("[DEBUG] creating VPC security group")
	created, err := securitygroup.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcSecGroupV2Object, err))
	}

	d.SetId(created.ID)

	if d.Get("delete_default_rules").(bool) {
		if err := deleteVPCSecGroupV2DefaultRules(ctx, client, created); err != nil {
			return diag.FromErr(err)
		}
	}

	if tags, ok := d.GetOk("tags"); ok {
		values := convert.ToStringSlice(tags.(*schema.Set).List())
		if _, err := securitygroup.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcSecGroupV2Object, created.ID, err))
		}
	}

	return resourceVPCSecGroupV2Read(ctx, d, meta)
}

func resourceVPCSecGroupV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC security group %s", d.Id())
	found, err := securitygroup.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcSecGroupV2Object)
	}

	flattenVPCSecGroupV2(d, found)

	return nil
}

func resourceVPCSecGroupV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     securitygroup.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = new(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = new(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("stateful") {
		request.Stateful = new(d.Get("stateful").(bool))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating VPC security group %s", d.Id())
		if _, err := securitygroup.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcSecGroupV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandVPCUpdateTagsV2(d)
		if _, err := securitygroup.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcSecGroupV2Object, d.Id(), err))
		}
	}

	return resourceVPCSecGroupV2Read(ctx, d, meta)
}

func resourceVPCSecGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC security group %s", d.Id())
	err = deleteVPCV2WithRetry(ctx, vpcSecGroupV2Object, d.Id(),
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutDelete), vpcV2ConflictIsTransient,
		func(ctx context.Context) error { return securitygroup.Delete(ctx, client, d.Id()) })
	if err != nil {
		return diag.FromErr(errDeletingObject(vpcSecGroupV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCSecGroupV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func deleteVPCSecGroupV2DefaultRules(
	ctx context.Context, client *vpc.Client, group *securitygroup.SecurityGroup,
) error {
	for _, rule := range group.SecurityGroupRules {
		log.Printf("[DEBUG] deleting default rule %s of VPC security group %s", rule.ID, group.ID)
		if err := securitygrouprule.Delete(ctx, client, rule.ID); err != nil {
			if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
				continue
			}

			return errDeletingObject(vpcSecGroupRuleV2Object, rule.ID, err)
		}
	}

	return nil
}

func flattenVPCSecGroupV2(d *schema.ResourceData, found *securitygroup.SecurityGroup) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("stateful", found.Stateful)
	d.Set("shared", found.Shared)

	flattenVPCTagsV2(d, found.Tags)

	if err := d.Set("security_group_rules", flattenVPCSecGroupV2RuleIDs(found.SecurityGroupRules)); err != nil {
		log.Print(errSettingComplexAttr("security_group_rules", err))
	}
}

func flattenVPCSecGroupV2RuleIDs(rules []securitygrouprule.Rule) []string {
	result := make([]string, 0, len(rules))
	for _, rule := range rules {
		result = append(result, rule.ID)
	}

	return result
}
