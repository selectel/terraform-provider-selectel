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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/dbaas-go"
)

func resourceDBaaSDatabaseV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSDatabaseV1Create,
		ReadContext:   resourceDBaaSDatabaseV1Read,
		UpdateContext: resourceDBaaSDatabaseV1Update,
		DeleteContext: resourceDBaaSDatabaseV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSDatabaseV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"datastore_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"lc_collate": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
			},
			"lc_ctype": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDBaaSDatabaseV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	selMutexKV.Lock(datastoreID)
	defer selMutexKV.Unlock(datastoreID)

	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	databaseCreateOpts := dbaas.DatabaseCreateOpts{
		DatastoreID: datastoreID,
		Name:        d.Get("name").(string),
		OwnerID:     d.Get("owner_id").(string),
		LcCollate:   d.Get("lc_collate").(string),
		LcCtype:     d.Get("lc_ctype").(string),
	}

	log.Print(msgGet(objectDatabase, d.Id()))
	database, err := dbaasClient.CreateDatabase(ctx, databaseCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatabase, err))
	}

	log.Printf("[DEBUG] waiting for database %s to become 'ACTIVE'", database.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForDBaaSDatabaseV1ActiveState(ctx, dbaasClient, database.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatabase, err))
	}

	d.SetId(database.ID)

	return resourceDBaaSDatabaseV1Read(ctx, d, meta)
}

func resourceDBaaSDatabaseV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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

func resourceDBaaSDatabaseV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	selMutexKV.Lock(datastoreID)
	defer selMutexKV.Unlock(datastoreID)

	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("owner_id") {
		ownerID := d.Get("owner_id").(string)

		selMutexKV.Lock(ownerID)
		defer selMutexKV.Unlock(ownerID)

		updateOpts := dbaas.DatabaseUpdateOpts{
			OwnerID: ownerID,
		}

		log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
		_, err := dbaasClient.UpdateDatabase(ctx, d.Id(), updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectDatabase, d.Id(), err))
		}

		log.Printf("[DEBUG] waiting for database %s to become 'ACTIVE'", d.Id())
		timeout := d.Timeout(schema.TimeoutCreate)
		err = waitForDBaaSDatabaseV1ActiveState(ctx, dbaasClient, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectDatabase, d.Id(), err))
		}
	}

	return resourceDBaaSDatabaseV1Read(ctx, d, meta)
}

func resourceDBaaSDatabaseV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	selMutexKV.Lock(datastoreID)
	defer selMutexKV.Unlock(datastoreID)

	ownerIDRaw, ownerIDOk := d.GetOk("owner_id")
	if ownerIDOk {
		ownerID := ownerIDRaw.(string)
		selMutexKV.Lock(ownerID)
		defer selMutexKV.Unlock(ownerID)
	}

	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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
		Refresh:    dbaasDatabaseV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	log.Printf("[DEBUG] waiting for database %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the database %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSDatabaseV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("SEL_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return nil, errors.New("SEL_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}

func waitForDBaaSDatabaseV1ActiveState(
	ctx context.Context, client *dbaas.API, databaseID string, timeout time.Duration) error {
	pending := []string{
		string(dbaas.StatusPendingCreate),
		string(dbaas.StatusPendingUpdate),
	}
	target := []string{
		string(dbaas.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    dbaasDatabaseV1StateRefreshFunc(ctx, client, databaseID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the database %s to become 'ACTIVE': %s",
			databaseID, err)
	}

	return nil
}

func dbaasDatabaseV1StateRefreshFunc(ctx context.Context, client *dbaas.API, databaseID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Database(ctx, databaseID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}

func dbaasDatabaseV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Database(ctx, datastoreID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}
			return nil, "", err
		}

		return d, strconv.Itoa(200), err
	}
}

func dbaasDatabaseV1LocaleDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// The default locale value - C is the same as null value, so we need to suppress
	if old == "C" && new == "" {
		return true
	}
	return false
}
