package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/groups"
	"github.com/selectel/iam-go/service/roles"
)

func resourceIAMGroupV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a Group in IAM API",
		CreateContext: resourceIAMGroupV1Create,
		ReadContext:   resourceIAMGroupV1Read,
		UpdateContext: resourceIAMGroupV1Update,
		DeleteContext: resourceIAMGroupV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the group.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the group.",
			},
			"role": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Role block of the group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"scope": {
							Type:     schema.TypeString,
							Required: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceIAMGroupV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	roles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	opts := groups.CreateRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	log.Print(msgCreate(objectGroup, opts))

	group, err := iamClient.Groups.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGroup, err))
	}
	d.SetId(group.ID)

	if len(roles) != 0 {
		err = iamClient.Groups.AssignRoles(ctx, group.ID, roles)
		if err != nil {
			return diag.FromErr(errCreatingObject(objectGroup, err))
		}
	}

	return resourceIAMGroupV1Read(ctx, d, meta)
}

func resourceIAMGroupV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectGroup, d.Id()))
	group, err := iamClient.Groups.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGroup, d.Id(), err))
	}

	d.Set("role", convertIAMRolesToSet(group.Roles))

	return nil
}

func resourceIAMGroupV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	description := d.Get("description").(string)

	opts := groups.UpdateRequest{
		Name:        d.Get("name").(string),
		Description: &description,
	}

	log.Print(msgUpdate(objectGroup, d.Id(), fmt.Sprintf("Name: %+v, description: %+v", opts.Name, opts.Description)))
	_, err := iamClient.Groups.Update(ctx, d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectGroup, d.Id(), err))
	}

	if d.HasChange("role") {
		currentGroup, err := iamClient.Groups.Get(ctx, d.Id())
		if err != nil {
			return diag.FromErr(errGettingObject(objectGroup, d.Id(), err))
		}
		oldRoles := currentGroup.Roles
		newRoles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}

		rolesToUnassign, rolesToAssign := diffRoles(oldRoles, newRoles)

		log.Print(msgUpdate(objectGroup, d.Id(), fmt.Sprintf("Roles to unassign: %+v, roles to assign: %+v", rolesToUnassign, rolesToAssign)))
		err = applyGroupRoles(ctx, d, iamClient, rolesToUnassign, rolesToAssign)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectGroup, d.Id(), err))
		}

		return nil
	}

	return resourceIAMGroupV1Read(ctx, d, meta)
}

func resourceIAMGroupV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectGroup, d.Id()))
	err := iamClient.Groups.Delete(ctx, d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrGroupNotFound) {
		return diag.FromErr(errDeletingObject(objectGroup, d.Id(), err))
	}

	d.SetId("")

	return nil
}

func applyGroupRoles(ctx context.Context, d *schema.ResourceData, iamClient *iam.Client, rolesToUnassign, rolesToAssign []roles.Role) error {
	if len(rolesToAssign) != 0 {
		err := iamClient.Groups.AssignRoles(ctx, d.Id(), rolesToAssign)
		if err != nil {
			return err
		}
	}

	if len(rolesToUnassign) != 0 {
		err := iamClient.Groups.UnassignRoles(ctx, d.Id(), rolesToUnassign)
		if err != nil {
			return err
		}
	}

	return nil
}
