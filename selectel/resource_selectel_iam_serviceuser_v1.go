package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
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
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Description: "List of roles of the Service User.",
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
		},
	}
}

func resourceIAMServiceUserV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	roles := getIAMServiceUserRolesFromList(d.Get("role").([]interface{}))

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
	d.Set("role", flattenIAMServiceUserRoles(user.Roles))
	if _, ok := d.GetOk("password"); !ok {
		d.Set("password", "UNKNOWN")
	}

	return nil
}

func resourceIAMServiceUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var password string
	if d.Get("password") == "UNKNOWN" {
		password = ""
	} else {
		password = d.Get("password").(string)
	}

	_, err := iamClient.ServiceUsers.Update(ctx, d.Id(), serviceusers.UpdateRequest{
		Enabled:  d.Get("enabled").(bool),
		Name:     d.Get("name").(string),
		Password: password,
	})

	if d.HasChange("role") {
		oldState, newState := d.GetChange("role")
		newRoles := getIAMServiceUserRolesFromList(newState.([]interface{}))
		oldRoles := getIAMServiceUserRolesFromList(oldState.([]interface{}))

		rolesToUnassign := make([]serviceusers.Role, 0)
		rolesToAssign := make([]serviceusers.Role, 0)

		for _, oldRole := range oldRoles {
			if !containsServiceUsersRole(newRoles, oldRole) {
				rolesToUnassign = append(rolesToUnassign, oldRole)
			}
		}

		for _, newRole := range newRoles {
			if !containsServiceUsersRole(oldRoles, newRole) {
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
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectServiceUser, d.Id(), err))
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
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errDeletingObject(objectServiceUser, d.Id(), err))
	}

	return nil
}
