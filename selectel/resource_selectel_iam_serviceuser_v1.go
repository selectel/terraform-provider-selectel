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
	"github.com/selectel/iam-go/service/roles"
	"github.com/selectel/iam-go/service/serviceusers"
)

func resourceIAMServiceUserV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a Service User in IAM API",
		CreateContext: resourceIAMServiceUserV1Create,
		ReadContext:   resourceIAMServiceUserV1Read,
		UpdateContext: resourceIAMServiceUserV1Update,
		DeleteContext: resourceIAMServiceUserV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates whether the Service User is enabled. True by default.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Service User.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password of the Service User.",
			},
			"role": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Role block of the Service User.",
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

func resourceIAMServiceUserV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	roles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print(msgCreate(objectServiceUser, d.Id()))
	user, err := iamClient.ServiceUsers.Create(ctx, serviceusers.CreateRequest{
		Enabled:  d.Get("enabled").(bool),
		Name:     d.Get("name").(string),
		Password: d.Get("password").(string),
		Roles:    roles,
	})
	if err != nil {
		return diag.FromErr(errCreatingObject(objectServiceUser, err))
	}

	d.SetId(user.ID)

	return resourceIAMServiceUserV1Read(ctx, d, meta)
}

func resourceIAMServiceUserV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectServiceUser, d.Id()))
	user, err := iamClient.ServiceUsers.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectServiceUser, d.Id(), err))
	}

	d.Set("name", user.Name)
	d.Set("enabled", user.Enabled)
	d.Set("role", convertIAMRolesToSet(user.Roles))
	if _, ok := d.GetOk("password"); !ok {
		d.Set("password", importIAMUndefined)
	}

	return nil
}

func resourceIAMServiceUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	password := d.Get("password").(string)
	if password == importIAMUndefined {
		password = ""
	}

	opts := serviceusers.UpdateRequest{
		Enabled:  d.Get("enabled").(bool),
		Name:     d.Get("name").(string),
		Password: password,
	}

	log.Print(msgUpdate(objectServiceUser, d.Id(), fmt.Sprintf("Enabled: %v, Name: %v", opts.Enabled, opts.Name)))
	_, err := iamClient.ServiceUsers.Update(ctx, d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
	}

	if d.HasChange("role") {
		currentUser, err := iamClient.ServiceUsers.Get(ctx, d.Id())
		if err != nil {
			return diag.FromErr(errGettingObject(objectServiceUser, d.Id(), err))
		}
		oldRoles := currentUser.Roles
		newRoles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}

		rolesToUnassign, rolesToAssign := diffRoles(oldRoles, newRoles)

		log.Print(msgUpdate(objectServiceUser, d.Id(), fmt.Sprintf("Roles to unassign: %+v, roles to assign: %+v", rolesToUnassign, rolesToAssign)))
		err = applyServiceUserRoles(ctx, d, iamClient, rolesToUnassign, rolesToAssign)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
		}

		return nil
	}

	return resourceIAMServiceUserV1Read(ctx, d, meta)
}

func resourceIAMServiceUserV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectServiceUser, d.Id()))
	err := iamClient.ServiceUsers.Delete(ctx, d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
		return diag.FromErr(errDeletingObject(objectServiceUser, d.Id(), err))
	}

	return nil
}

func applyServiceUserRoles(ctx context.Context, d *schema.ResourceData, iamClient *iam.Client, rolesToUnassign, rolesToAssign []roles.Role) error {
	if len(rolesToAssign) != 0 {
		err := iamClient.ServiceUsers.AssignRoles(ctx, d.Id(), rolesToAssign)
		if err != nil {
			return err
		}
	}

	if len(rolesToUnassign) != 0 {
		err := iamClient.ServiceUsers.UnassignRoles(ctx, d.Id(), rolesToUnassign)
		if err != nil {
			return err
		}
	}

	return nil
}
