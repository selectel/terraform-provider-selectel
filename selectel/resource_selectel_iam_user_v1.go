package selectel

import (
	"context"
	"errors"
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
				Required:    true,
				ForceNew:    true,
				Description: "Authentication type of the User. Can be 'local' or 'federative'.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email of the User.",
			},
			"federation": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Federation data of the User. Can be set only if 'auth_type' is 'federative'.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"external_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
					},
				},
			},
			"role": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Description: "List of roles of the User.",
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

	roles := getIAMUserRolesFromList(d.Get("role").([]interface{}))
	federation, err := getIAMUserFederationFromSet(d.Get("federation").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
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
		d.Set("email", "UNKNOWN")
	}
	if user.Federation != nil {
		d.Set("federation", flattenIAMUserFederation(user.Federation))
	}
	d.Set("role", flattenIAMUserRoles(user.Roles))

	return nil
}

func resourceIAMUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("email") || d.HasChange("auth_type") {
		resourceIAMUserV1Delete(ctx, d, meta)
		resourceIAMUserV1Create(ctx, d, meta)
	}

	if d.HasChange("role") {
		oldState, newState := d.GetChange("role")
		newRoles := getIAMUserRolesFromList(newState.([]interface{}))
		oldRoles := getIAMUserRolesFromList(oldState.([]interface{}))

		rolesToUnassign := make([]users.Role, 0)
		rolesToAssign := make([]users.Role, 0)

		for _, oldRole := range oldRoles {
			if !containsUsersRole(newRoles, oldRole) {
				rolesToUnassign = append(rolesToUnassign, oldRole)
			}
		}

		for _, newRole := range newRoles {
			if !containsUsersRole(oldRoles, newRole) {
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
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errDeletingObject(objectUser, d.Id(), err))
	}

	return nil
}
