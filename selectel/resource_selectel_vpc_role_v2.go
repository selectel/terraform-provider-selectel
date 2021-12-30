package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/roles"
)

func resourceVPCRoleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRoleV2Create,
		ReadContext:   resourceVPCRoleV2Read,
		DeleteContext: resourceVPCRoleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCRoleV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()

	opts := roles.RoleOpt{
		ProjectID: d.Get("project_id").(string),
		UserID:    d.Get("user_id").(string),
	}

	log.Print(msgCreate(objectRole, opts))
	role, _, err := roles.Create(ctx, resellV2Client, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRole, err))
	}

	d.SetId(resourceVPCRoleV2BuildID(role.ProjectID, role.UserID))

	return nil
}

func resourceVPCRoleV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()

	log.Print(msgGet(objectRole, d.Id()))
	projectID, userID, err := resourceVPCRoleV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(errParseID(objectRole, d.Id()))
	}
	projectRoles, _, err := roles.ListProject(ctx, resellV2Client, projectID)
	if err != nil {
		return diag.FromErr(errSearchingProjectRole(projectID, err))
	}

	found := false
	for _, role := range projectRoles {
		if role.UserID == userID {
			found = true
			d.Set("project_id", role.ProjectID)
			d.Set("user_id", role.UserID)
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceVPCRoleV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()

	projectID, userID, err := resourceVPCRoleV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	opts := roles.RoleOpt{
		ProjectID: projectID,
		UserID:    userID,
	}

	log.Print(msgDelete(objectRole, d.Id()))
	response, err := roles.Delete(ctx, resellV2Client, opts)
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")

				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectRole, d.Id(), err))
	}

	return nil
}

func resourceVPCRoleV2BuildID(projectID, userID string) string {
	return fmt.Sprintf("%s/%s", projectID, userID)
}

func resourceVPCRoleV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unable to parse id: '%s'", id)
	}

	return idParts[0], idParts[1], nil
}
