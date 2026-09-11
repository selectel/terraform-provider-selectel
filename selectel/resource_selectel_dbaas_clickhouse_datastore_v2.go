package selectel

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSV2ClickhouseDatastore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSV2ClickhouseDatastoreCreate,
		ReadContext:   resourceDBaaSV2ClickhouseDatastoreRead,
		UpdateContext: resourceDBaaSV2ClickhouseDatastoreUpdate,
		DeleteContext: resourceDBaaSV2ClickhouseDatastoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSV2ClickhouseDatastoreImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSV2ClickhouseDatastoreSchema(),
	}
}

func resourceDBaaSV2ClickhouseDatastoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	var datastoreCreateOpts dbaas_v2_ch.DatastoreCreateRequest
	// TODO: // prepare datastoreCreateOpts

	log.Print(msgCreate(objectDatastore, datastoreCreateOpts))
	datastore, err := dbaasClient.ClickHouse.CreateDatastore(ctx, datastoreCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.ClickHouse, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(datastore.ID)

	return resourceDBaaSV2ClickhouseDatastoreRead(ctx, d, meta)
}

func resourceDBaaSV2ClickhouseDatastoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatastore, d.Id()))
	datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, d.Id(), err))
	}

	// TODO: Fill Set
	d.Set("name", datastore.Name)
	// Do other

	return nil
}

func resourceDBaaSV2ClickhouseDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	// if diagErr != nil {
	// 	return diagErr
	// }
	// timeout := d.Timeout(schema.TimeoutUpdate)

	// TODO: check d.HaseChange for params and update them

	return resourceDBaaSV2ClickhouseDatastoreRead(ctx, d, meta)
}

func resourceDBaaSV2ClickhouseDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectDatastore, d.Id()))
	err := dbaasClient.ClickHouse.DeleteDatastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatastore, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become deleted", d.Id())
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waiters.WaitForDBaaSV2DatastoreDeleted(ctx, dbaasClient.ClickHouse, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatastore, d.Id(), err))
	}
	return nil
}

func resourceDBaaSV2ClickhouseDatastoreImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
