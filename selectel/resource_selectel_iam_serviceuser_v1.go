package selectel

import (
	"context"
	"errors"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				ForceNew:    false,
				Description: "Indicates whether the Service User is enabled. True by default.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Name of the Service User.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Sensitive:   true,
				Description: "Password of the Service User.",
			},
			"role": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "Role block of the Service User.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
							ValidateFunc: validation.StringInSlice([]string{
								string(roles.AccountOwner),
								string(roles.Billing),
								string(roles.IAMAdmin),
								string(roles.Member),
								string(roles.Reader),
								string(roles.ObjectStorageAdmin),
								string(roles.ObjectStorageUser),
							}, false),
						},
						"scope": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
							ValidateFunc: validation.StringInSlice([]string{
								string(roles.Account),
								string(roles.Project),
							}, false),
						},
						"project_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
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
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errGettingObject(objectServiceUser, d.Id(), err))
	}

	d.Set("name", user.Name)
	d.Set("enabled", user.Enabled)
	d.Set("role", convertIAMRolesToSet(user.Roles))
	if _, ok := d.GetOk("password"); !ok {
		d.Set("password", importIAMFieldValueFailed)
	}

	return nil
}

func resourceIAMServiceUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	password := d.Get("password").(string)
	if password == "UNKNOWN" {
		password = ""
	}

	_, err := iamClient.ServiceUsers.Update(ctx, d.Id(), serviceusers.UpdateRequest{
		Enabled:  d.Get("enabled").(bool),
		Name:     d.Get("name").(string),
		Password: password,
	})
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
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
			err := iamClient.ServiceUsers.AssignRoles(ctx, d.Id(), rolesToAssign)
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
			}
		}

		if len(rolesToUnassign) != 0 {
			err := iamClient.ServiceUsers.UnassignRoles(ctx, d.Id(), rolesToUnassign)
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
			}
		}
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
