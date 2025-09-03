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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSMySQLDatastoreV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSMySQLDatastoreV1Create,
		ReadContext:   resourceDBaaSMySQLDatastoreV1Read,
		UpdateContext: resourceDBaaSMySQLDatastoreV1Update,
		DeleteContext: resourceDBaaSMySQLDatastoreV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSMySQLDatastoreV1ImportState,
		},
		CustomizeDiff: customdiff.All(
			refreshDatastoreInstancesOutputsDiff,
		),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSMySQLDatastoreV1Schema(),
	}
}

func resourceDBaaSMySQLDatastoreV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	flavorID, flavorIDOk := d.GetOk("flavor_id")
	flavorRaw, flavorOk := d.GetOk("flavor")

	if flavorIDOk == flavorOk {
		return diag.FromErr(errors.New("either 'flavor' or 'flavor_id' must be provided"))
	}

	typeID := d.Get("type_id").(string)
	diagErr = validateDatastoreType(ctx, []string{mySQLDatastoreType, mySQLNativeDatastoreType}, typeID, dbaasClient)
	if diagErr != nil {
		return diagErr
	}

	restoreSet := d.Get("restore").(*schema.Set)
	restore, err := resourceDBaaSDatastoreV1RestoreOptsFromSet(restoreSet)
	if err != nil {
		return diag.FromErr(errParseDatastoreV1Restore(err))
	}
	floatingIPsSet := d.Get("floating_ips").(*schema.Set)
	floatingIPsSchema, err := resourceDBaaSDatastoreV1FloatingIPsOptsFromSet(floatingIPsSet)
	if err != nil {
		return diag.FromErr(errParseDatastoreV1FloatingIPs(err))
	}

	datastoreCreateOpts := dbaas.DatastoreCreateOpts{
		ProjectID:   d.Get("project_id").(string),
		Name:        d.Get("name").(string),
		TypeID:      typeID,
		SubnetID:    d.Get("subnet_id").(string),
		NodeCount:   d.Get("node_count").(int),
		Restore:     restore,
		Config:      d.Get("config").(map[string]interface{}),
		FloatingIPs: floatingIPsSchema,
	}

	if flavorOk {
		flavorSet := flavorRaw.(*schema.Set)
		flavor, err := resourceDBaaSDatastoreV1FlavorFromSet(flavorSet)
		if err != nil {
			return diag.FromErr(errParseDatastoreV1Flavor(err))
		}

		datastoreCreateOpts.Flavor = flavor
	}

	if flavorIDOk {
		datastoreCreateOpts.FlavorID = flavorID.(string)
	}

	backupRetentionDays, ok := d.GetOk("backup_retention_days")
	if ok {
		datastoreCreateOpts.BackupRetentionDays = backupRetentionDays.(int)
	}

	logPlatform, logPlatformOk := d.GetOk("log_platform")
	if logPlatformOk {
		datastoreCreateOpts.LogPlatform = &dbaas.DatastoreLogGroup{LogGroup: logPlatform.(string)}
	}

	log.Print(msgCreate(objectDatastore, datastoreCreateOpts))
	datastore, err := dbaasClient.CreateDatastore(ctx, datastoreCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSDatastoreV1ActiveState(ctx, dbaasClient, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(datastore.ID)

	return resourceDBaaSMySQLDatastoreV1Read(ctx, d, meta)
}

func resourceDBaaSMySQLDatastoreV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatastore, d.Id()))
	datastore, err := dbaasClient.Datastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, d.Id(), err))
	}
	d.Set("name", datastore.Name)
	d.Set("status", datastore.Status)
	d.Set("project_id", datastore.ProjectID)
	d.Set("subnet_id", datastore.SubnetID)
	d.Set("type_id", datastore.TypeID)
	d.Set("node_count", datastore.NodeCount)
	d.Set("enabled", datastore.Enabled)
	d.Set("flavor_id", datastore.FlavorID)
	d.Set("backup_retention_days", datastore.BackupRetentionDays)
	d.Set("log_platform", datastore.LogPlatform.LogGroup)

	flavor := resourceDBaaSDatastoreV1FlavorToSet(datastore.Flavor)
	if err := d.Set("flavor", flavor); err != nil {
		log.Print(errSettingComplexAttr("flavor", err))
	}

	if err := d.Set("connections", datastore.Connection); err != nil {
		log.Print(errSettingComplexAttr("connections", err))
	}

	instances := resourceDBaaSDatastoreV1InstancesToList(datastore.Instances)
	if err := d.Set("instances", instances); err != nil {
		log.Print(errSettingComplexAttr("instances", err))
	}

	configMap := make(map[string]string)
	for key, value := range datastore.Config {
		configMap[key] = convertFieldToStringByType(value)
	}
	if err := d.Set("config", configMap); err != nil {
		log.Print(errSettingComplexAttr("config", err))
	}

	return nil
}

func resourceDBaaSMySQLDatastoreV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("name") {
		err := updateDatastoreName(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("firewall") {
		err := updateDatastoreFirewall(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("node_count") || d.HasChange("flavor") || d.HasChange("flavor_id") {
		err := resizeDatastore(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("config") {
		err := updateDatastoreConfig(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("backup_retention_days") {
		err := updateDatastoreBackups(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("floating_ips") {
		err := updateDatastoreFloatingIPs(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("log_platform") {
		if err := dbaasLogPlatformUpdate(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDBaaSMySQLDatastoreV1Read(ctx, d, meta)
}

func resourceDBaaSMySQLDatastoreV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectDatastore, d.Id()))
	err := dbaasClient.DeleteDatastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatastore, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    waiters.DBaaSDatastoreV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	log.Printf("[DEBUG] waiting for datastore %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the datastore %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSMySQLDatastoreV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
