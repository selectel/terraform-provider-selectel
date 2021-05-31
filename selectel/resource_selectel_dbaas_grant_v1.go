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
		Schema: map[string]*schema.Schema{

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
			"database_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDBaaSGrantV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	selMutexKV.Lock(datastoreID)
	defer selMutexKV.Unlock(datastoreID)

	databaseID := d.Get("database_id").(string)
	selMutexKV.Lock(databaseID)
	defer selMutexKV.Unlock(databaseID)

	userID := d.Get("user_id").(string)
	selMutexKV.Lock(userID)
	defer selMutexKV.Unlock(userID)

	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	grantCreateOpts := dbaas.GrantCreateOpts{
		DatastoreID: datastoreID,
		DatabaseID:  databaseID,
		UserID:      userID,
	}

	log.Print(msgGet(objectGrant, d.Id()))
	grant, err := dbaasClient.CreateGrant(ctx, grantCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGrant, err))
	}

	log.Printf("[DEBUG] waiting for grant %s to become 'ACTIVE'", grant.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForDBaaSGrantV1ActiveState(ctx, dbaasClient, grant.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGrant, err))
	}

	d.SetId(grant.ID)

	return resourceDBaaSGrantV1Read(ctx, d, meta)
}

func resourceDBaaSGrantV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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
	datastoreID := d.Get("datastore_id").(string)

	selMutexKV.Lock(datastoreID)
	defer selMutexKV.Unlock(datastoreID)

	databaseID := d.Get("database_id").(string)
	selMutexKV.Lock(databaseID)
	defer selMutexKV.Unlock(databaseID)

	userID := d.Get("user_id").(string)
	selMutexKV.Lock(userID)
	defer selMutexKV.Unlock(userID)

	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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
		Refresh:    dbaasGrantV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
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
		return nil, errors.New("SEL_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return nil, errors.New("SEL_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}

func waitForDBaaSGrantV1ActiveState(
	ctx context.Context, client *dbaas.API, grantID string, timeout time.Duration) error {
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
		Refresh:    dbaasGrantV1StateRefreshFunc(ctx, client, grantID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the grant %s to become 'ACTIVE': %s",
			grantID, err)
	}

	return nil
}

func dbaasGrantV1StateRefreshFunc(ctx context.Context, client *dbaas.API, grantID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Grant(ctx, grantID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}

func dbaasGrantV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, grantID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Grant(ctx, grantID)
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
