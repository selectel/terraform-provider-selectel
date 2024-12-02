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

func resourceDBaaSExtensionV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSExtensionV1Create,
		ReadContext:   resourceDBaaSExtensionV1Read,
		DeleteContext: resourceDBaaSExtensionV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSExtensionV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSPostgreSQLExtensionV1Schema(),
	}
}

func resourceDBaaSExtensionV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	extensionCreateOpts := dbaas.ExtensionCreateOpts{
		AvailableExtensionID: d.Get("available_extension_id").(string),
		DatastoreID:          d.Get("datastore_id").(string),
		DatabaseID:           d.Get("database_id").(string),
	}

	log.Print(msgCreate(objectExtension, extensionCreateOpts))
	extension, err := dbaasClient.CreateExtension(ctx, extensionCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectExtension, err))
	}

	log.Printf("[DEBUG] waiting for extension %s to become 'ACTIVE'", extension.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSExtensionV1ActiveState(ctx, dbaasClient, extension.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectExtension, err))
	}

	d.SetId(extension.ID)

	return resourceDBaaSExtensionV1Read(ctx, d, meta)
}

func resourceDBaaSExtensionV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectExtension, d.Id()))
	extension, err := dbaasClient.Extension(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectExtension, d.Id(), err))
	}

	d.Set("available_extension_id", extension.AvailableExtensionID)
	d.Set("datastore_id", extension.DatastoreID)
	d.Set("database_id", extension.DatabaseID)

	return nil
}

func resourceDBaaSExtensionV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectExtension, d.Id()))
	err := dbaasClient.DeleteExtension(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectExtension, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    waiters.DBaaSExtensionV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	log.Printf("[DEBUG] waiting for extension %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the extension %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSExtensionV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
