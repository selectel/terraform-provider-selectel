package selectel

import (
	"context"
	"errors"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"auth_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "local",
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
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Federation data of the User. Can be set only if 'auth_type' is 'federated'.",
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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

	roles, err := convertIAMRoles(d.Get("role").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	federation, err := convertIAMUserFederation(d.Get("federation").(map[string]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	authType := d.Get("auth_type").(string)
	if authType != "local" && authType != "federated" {
		return diag.Errorf("auth_type can be only 'local' or 'federated'")
	}
	if authType == "local" && federation != nil {
		return diag.Errorf("federation can be set only if auth_type is 'federated'")
	}
	user, err := iamClient.Users.Create(ctx, users.CreateRequest{
		AuthType:   users.AuthType(d.Get("auth_type").(string)),
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
			return nil
		}

		return diag.FromErr(errGettingObject(objectUser, d.Id(), err))
	}

	d.Set("keystone_id", user.KeystoneID)
	d.Set("auth_type", user.AuthType)
	if _, ok := d.GetOk("email"); !ok {
		d.Set("email", importFailedIAMFieldValue)
	}
	if user.Federation != nil {
		d.Set("federation", flattenIAMUserFederation(user.Federation))
	}
	d.Set("role", flattenIAMRoles(user.Roles))

	return nil
}

func resourceIAMUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("role") {
		oldState, newState := d.GetChange("role")
		newRoles, err := convertIAMRoles(newState.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
		oldRoles, err := convertIAMRoles(oldState.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}

		rolesToUnassign := make([]roles.Role, 0)
		rolesToAssign := make([]roles.Role, 0)

		for _, oldRole := range oldRoles {
			if !slices.Contains(newRoles, oldRole) {
				rolesToUnassign = append(rolesToUnassign, oldRole)
			}
		}

		for _, newRole := range newRoles {
			if !slices.Contains(oldRoles, newRole) {
				rolesToAssign = append(rolesToAssign, newRole)
			}
		}
		if len(rolesToAssign) != 0 {
			err := iamClient.Users.AssignRoles(ctx, d.Id(), rolesToAssign)
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectUser, d.Id(), err))
			}
		}

		if len(rolesToUnassign) != 0 {
			err := iamClient.Users.UnassignRoles(ctx, d.Id(), rolesToUnassign)
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectUser, d.Id(), err))
			}
		}
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
