package selvpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/roles"
)

func resourceResellRoleV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellRoleV2Create,
		Read:   resourceResellRoleV2Read,
		Delete: resourceResellRoleV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceResellRoleV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	opts := roles.RoleOpt{
		ProjectID: d.Get("project_id").(string),
		UserID:    d.Get("user_id").(string),
	}

	log.Print(msgCreate(objectRole, opts))
	role, _, err := roles.Create(ctx, resellV2Client, opts)
	if err != nil {
		return errCreatingObject(objectRole, err)
	}

	d.SetId(resourceResellRoleV2BuildID(role.ProjectID, role.UserID))

	return nil
}

func resourceResellRoleV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgGet(objectRole, d.Id()))
	projectID, userID, err := resourceResellRoleV2ParseID(d.Id())
	if err != nil {
		return errParseID(objectRole, d.Id())
	}
	projectRoles, _, err := roles.ListProject(ctx, resellV2Client, projectID)
	if err != nil {
		return errSearchingProjectRole(projectID, err)
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

func resourceResellRoleV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID, userID, err := resourceResellRoleV2ParseID(d.Id())
	if err != nil {
		return err
	}

	opts := roles.RoleOpt{
		ProjectID: projectID,
		UserID:    userID,
	}

	log.Print(msgDelete(objectRole, d.Id()))
	response, err := roles.Delete(ctx, resellV2Client, opts)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject(objectRole, d.Id(), err)
	}

	return nil
}

func resourceResellRoleV2BuildID(projectID, userID string) string {
	return fmt.Sprintf("%s/%s", projectID, userID)
}

func resourceResellRoleV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unable to parse id: '%s'", id)
	}

	return idParts[0], idParts[1], nil
}
