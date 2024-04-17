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
	"github.com/selectel/iam-go/service/users"
)

func resourceIAMUserV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a User in IAM API",
		CreateContext: resourceIAMUserV1Create,
		ReadContext:   resourceIAMUserV1Read,
		UpdateContext: resourceIAMUserV1Update,
		DeleteContext: resourceIAMUserV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email of the User.",
			},
			"federation": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Federation data of the User.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"role": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Role block of the User.",
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
			"keystone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIAMUserV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	roles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	federation, err := convertIAMListToUserFederation(d.Get("federation").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	authType := string(users.Local)
	if federation != nil {
		authType = string(users.Federated)
	}

	log.Print(msgCreate(objectUser, d.Id()))
	user, err := iamClient.Users.Create(ctx, users.CreateRequest{
		AuthType:   users.AuthType(authType),
		Email:      d.Get("email").(string),
		Federation: federation,
		Roles:      roles,
	})
	if err != nil {
		return diag.FromErr(errCreatingObject(objectUser, err))
	}
	d.SetId(user.ID)

	return resourceIAMUserV1Read(ctx, d, meta)
}

func resourceIAMUserV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectUser, d.Id()))
	user, err := iamClient.Users.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectUser, d.Id(), err))
	}

	d.Set("keystone_id", user.KeystoneID)
	if _, ok := d.GetOk("email"); !ok {
		d.Set("email", importIAMUndefined)
	}
	if user.Federation != nil {
		d.Set("federation", convertIAMFederationToList(user.Federation))
	}
	d.Set("role", convertIAMRolesToSet(user.Roles))

	return nil
}

func resourceIAMUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("role") {
		currentUser, err := iamClient.Users.Get(ctx, d.Id())
		if err != nil {
			return diag.FromErr(errGettingObject(objectUser, d.Id(), err))
		}
		oldRoles := currentUser.Roles
		newRoles, err := convertIAMSetToRoles(d.Get("role").(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}

		rolesToUnassign, rolesToAssign := diffRoles(oldRoles, newRoles)

		log.Print(msgUpdate(objectUser, d.Id(), fmt.Sprintf("Roles to unassign: %+v, roles to assign: %+v", rolesToUnassign, rolesToAssign)))
		err = applyUserRoles(ctx, d, iamClient, rolesToUnassign, rolesToAssign)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectUser, d.Id(), err))
		}

		return nil
	}

	return resourceIAMUserV1Read(ctx, d, meta)
}

func resourceIAMUserV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectUser, d.Id()))
	err := iamClient.Users.Delete(ctx, d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
		return diag.FromErr(errDeletingObject(objectUser, d.Id(), err))
	}

	return nil
}

func applyUserRoles(ctx context.Context, d *schema.ResourceData, iamClient *iam.Client, rolesToUnassign, rolesToAssign []roles.Role) error {
	if len(rolesToAssign) != 0 {
		err := iamClient.Users.AssignRoles(ctx, d.Id(), rolesToAssign)
		if err != nil {
			return err
		}
	}

	if len(rolesToUnassign) != 0 {
		err := iamClient.Users.UnassignRoles(ctx, d.Id(), rolesToUnassign)
		if err != nil {
			return err
		}
	}

	return nil
}
