package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSGrantV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSGrantV1Create,
		ReadContext:   resourceDBaaSGrantV1Read,
		DeleteContext: resourceDBaaSGrantV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSGrantV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSGrantV1Schema(),
	}
}

func resourceDBaaSGrantV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	grantCreateOpts := dbaas.GrantCreateOpts{
		DatastoreID: d.Get("datastore_id").(string),
		DatabaseID:  d.Get("database_id").(string),
		UserID:      d.Get("user_id").(string),
	}

	log.Print(msgCreate(objectGrant, grantCreateOpts))
	grant, err := dbaasClient.CreateGrant(ctx, grantCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGrant, err))
	}

	log.Printf("[DEBUG] waiting for grant %s to become 'ACTIVE'", grant.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSGrantV1ActiveState(ctx, dbaasClient, grant.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGrant, err))
	}

	d.SetId(grant.ID)

	return resourceDBaaSGrantV1Read(ctx, d, meta)
}

func resourceDBaaSGrantV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectGrant, d.Id()))
	grant, err := dbaasClient.Grant(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGrant, d.Id(), err))
	}
	d.Set("user_id", grant.UserID)
	d.Set("database_id", grant.DatabaseID)
	d.Set("datastore_id", grant.DatastoreID)
	d.Set("status", grant.Status)

	return nil
}

func resourceDBaaSGrantV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectGrant, d.Id()))
	err := dbaasClient.DeleteGrant(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGrant, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    waiters.DBaaSGrantV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	log.Printf("[DEBUG] waiting for grant %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the grant %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSGrantV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("INFRA_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return nil, errors.New("INFRA_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}
