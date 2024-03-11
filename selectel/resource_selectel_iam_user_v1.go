package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
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
			"auth_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     users.Local,
				ForceNew:    true,
				Description: "Authentication type of the User. Can be 'local' or 'federated'.",
			},
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
				Description: "Federation data of the User. Must be set only if 'auth_type' is 'federated'.",
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
				ForceNew:    false,
				Description: "Role block of the User.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"scope": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"project_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
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

	authType := d.Get("auth_type").(string)
	if authType != string(users.Federated) && federation != nil {
		return diag.Errorf("federation can be set only if auth_type is 'federated'")
	}
	if authType == string(users.Federated) && federation == nil {
		return diag.Errorf("federation must be set if auth_type is 'federated'")
	}

	log.Print(msgGet(objectUser, d.Id()))
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
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			d.SetId("")
		}

		return diag.FromErr(errGettingObject(objectUser, d.Id(), err))
	}

	d.Set("keystone_id", user.KeystoneID)
	d.Set("auth_type", user.AuthType)
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
		oldState, newState := d.GetChange("role")
		newRoles, err := convertIAMSetToRoles(newState.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
		oldRoles, err := convertIAMSetToRoles(oldState.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}

		rolesToUnassign, rolesToAssign := manageRoles(oldRoles, newRoles)

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
