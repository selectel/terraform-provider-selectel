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

func resourceDBaaSPostgreSQLDatabaseV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSPostgreSQLDatabaseV1Create,
		ReadContext:   resourceDBaaSPostgreSQLDatabaseV1Read,
		UpdateContext: resourceDBaaSPostgreSQLDatabaseV1Update,
		DeleteContext: resourceDBaaSPostgreSQLDatabaseV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSPostgreSQLDatabaseV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSPostgreSQLDatabaseV1Schema(),
	}
}

func resourceDBaaSPostgreSQLDatabaseV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	databaseCreateOpts := dbaas.DatabaseCreateOpts{
		DatastoreID: d.Get("datastore_id").(string),
		Name:        d.Get("name").(string),
		OwnerID:     d.Get("owner_id").(string),
		LcCollate:   d.Get("lc_collate").(string),
		LcCtype:     d.Get("lc_ctype").(string),
	}

	log.Print(msgCreate(objectDatabase, databaseCreateOpts))
	database, err := dbaasClient.CreateDatabase(ctx, databaseCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatabase, err))
	}

	log.Printf("[DEBUG] waiting for database %s to become 'ACTIVE'", database.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSDatabaseV1ActiveState(ctx, dbaasClient, database.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatabase, err))
	}

	d.SetId(database.ID)

	return resourceDBaaSPostgreSQLDatabaseV1Read(ctx, d, meta)
}

func resourceDBaaSPostgreSQLDatabaseV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatabase, d.Id()))
	database, err := dbaasClient.Database(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatabase, d.Id(), err))
	}
	d.Set("datastore_id", database.DatastoreID)
	d.Set("name", database.Name)
	d.Set("status", database.Status)
	if database.OwnerID != "" {
		d.Set("owner_id", database.OwnerID)
	}
	if database.LcCollate != "" {
		d.Set("lc_collate", database.LcCollate)
	}
	if database.LcCtype != "" {
		d.Set("lc_ctype", database.LcCtype)
	}

	return nil
}

func resourceDBaaSPostgreSQLDatabaseV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("owner_id") {
		updateOpts := dbaas.DatabaseUpdateOpts{
			OwnerID: d.Get("owner_id").(string),
		}

		log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
		_, err := dbaasClient.UpdateDatabase(ctx, d.Id(), updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectDatabase, d.Id(), err))
		}

		log.Printf("[DEBUG] waiting for database %s to become 'ACTIVE'", d.Id())
		timeout := d.Timeout(schema.TimeoutCreate)
		err = waiters.WaitForDBaaSDatabaseV1ActiveState(ctx, dbaasClient, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectDatabase, d.Id(), err))
		}
	}

	return resourceDBaaSPostgreSQLDatabaseV1Read(ctx, d, meta)
}

func resourceDBaaSPostgreSQLDatabaseV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectDatabase, d.Id()))
	err := dbaasClient.DeleteDatabase(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatabase, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    waiters.DBaaSDatabaseV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	log.Printf("[DEBUG] waiting for database %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the database %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSPostgreSQLDatabaseV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
