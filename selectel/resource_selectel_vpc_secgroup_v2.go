package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/securitygrouprule"
)

const vpcSecGroupV2Object = "VPC security group"

func resourceVPCSecGroupV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCSecGroupV2Create,
		ReadContext:   resourceVPCSecGroupV2Read,
		UpdateContext: resourceVPCSecGroupV2Update,
		DeleteContext: resourceVPCSecGroupV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCSecGroupV2ImportState,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"stateful": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
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
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"security_group_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
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
	if v, ok := d.GetOkExists("stateful"); ok { //nolint:staticcheck
		request.Stateful = new(v.(bool))
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
		values := expandToStringSlice(tags.(*schema.Set).List())
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
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
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
	if err := securitygroup.Delete(ctx, client, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcSecGroupV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCSecGroupV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
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

	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
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
