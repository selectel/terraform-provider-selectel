package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2 "github.com/selectel/dbaas-go/v2"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSV2ClickhouseShardGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSV2ClickhouseShardGroupCreate,
		ReadContext:   resourceDBaaSV2ClickhouseShardGroupRead,
		UpdateContext: resourceDBaaSV2ClickhouseShardGroupUpdate,
		DeleteContext: resourceDBaaSV2ClickhouseShardGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSV2ClickhouseShardGroupImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSV2ClickhouseShardGroupSchema(),
	}
}

func resourceDBaaSV2ClickhouseShardGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	datastoreID := d.Get("datastore_id").(string)
	log.Print(msgGet(objectDatastore, datastoreID))
	datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, datastoreID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, datastoreID, err))
	}

	shardNames := expandDBaaSV2ClickhouseShardNameFromSet(d.Get("shard_names").(*schema.Set))

	shardIDs, err := resolveDBaaSv2ClickhouseShardIDs(&datastore, shardNames)
	if err != nil {
		return diag.FromErr(err)
	}

	shardGroupCreateOpts := dbaas_v2_ch.ShardGroupCreateRequest{
		Name:     d.Get("name").(string),
		ShardIDs: shardIDs,
	}

	descRaw, descOk := d.GetOk("description")
	if descOk {
		desc := descRaw.(string)
		shardGroupCreateOpts.Description = &desc
	}

	log.Print(msgCreate(objectShardGroup, shardGroupCreateOpts))
	shardGroup, err := dbaasClient.ClickHouse.CreateShardGroup(ctx, datastoreID, shardGroupCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectShardGroup, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.ClickHouse, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(shardGroup.ID)

	return resourceDBaaSV2ClickhouseShardGroupRead(ctx, d, meta)
}

func resourceDBaaSV2ClickhouseShardGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	shardGroupID := d.Id()
	datastoreID := d.Get("datastore_id").(string)

	log.Print(msgGet(objectShardGroup, shardGroupID))
	shardGroup, err := getDBaaSV2ClickhouseShardGroup(ctx, dbaasClient, datastoreID, shardGroupID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectShardGroup, shardGroupID, err))
	}
	// shard group not found. Clearing the state.
	if shardGroup == nil {
		d.SetId("")
		return nil
	}

	log.Print(msgGet(objectDatastore, datastoreID))
	datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, datastoreID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, datastoreID, err))
	}

	shardNames, err := resolveDBaaSv2ClickhouseShardNames(&datastore, shardGroup.ShardIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", shardGroup.Name)
	d.Set("shard_names", shardNames)
	d.Set("description", shardGroup.Description)

	return nil
}

func resourceDBaaSV2ClickhouseShardGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}
	var shardGroupUpdateOpts dbaas_v2_ch.ShardGroupUpdateRequest
	shardGroupID := d.Id()
	datastoreID := d.Get("datastore_id").(string)

	log.Print(msgGet(objectDatastore, datastoreID))
	datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, datastoreID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, datastoreID, err))
	}

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		shardGroupUpdateOpts.Description = &desc
	}

	if d.HasChange("shard_names") {
		shardNames := expandDBaaSV2ClickhouseShardNameFromSet(d.Get("shard_names").(*schema.Set))

		shardIDs, err := resolveDBaaSv2ClickhouseShardIDs(&datastore, shardNames)
		if err != nil {
			return diag.FromErr(err)
		}
		shardGroupUpdateOpts.ShardIDs = &shardIDs
	}

	log.Print(msgUpdate(objectShardGroup, shardGroupID, shardGroupUpdateOpts))
	_, err = dbaasClient.ClickHouse.UpdateShardGroup(ctx, datastoreID, shardGroupID, shardGroupUpdateOpts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectShardGroup, shardGroupID, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.ClickHouse, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	return resourceDBaaSV2ClickhouseShardGroupRead(ctx, d, meta)
}

func resourceDBaaSV2ClickhouseShardGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	datastoreID := d.Get("datastore_id").(string)
	shardGroupID := d.Id()

	log.Print(msgDelete(objectShardGroup, shardGroupID))
	err := dbaasClient.ClickHouse.DeleteShardGroup(ctx, datastoreID, shardGroupID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectShardGroup, shardGroupID, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.ClickHouse, datastoreID, timeout)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectShardGroup, shardGroupID, err))
	}

	return nil
}

func resolveDBaaSv2ClickhouseShardIDs(
	datastore *dbaas_v2_ch.DatastoreResponse,
	shardNames []string,
) ([]string, error) {
	shardIDByName := make(map[string]string, len(datastore.NodeGroups))
	shardIDs := make([]string, 0, len(shardNames))

	for _, ng := range datastore.NodeGroups {
		if ng.Role == dbaas_v2_ch.NodeGroupRoleData {
			shardIDByName[ng.Name] = ng.ID
		}
	}

	for _, shardName := range shardNames {
		shardID, exists := shardIDByName[shardName]
		if !exists {
			return nil, fmt.Errorf("shard %s not found in datastore %s", shardName, datastore.ID)
		}

		shardIDs = append(shardIDs, shardID)
	}

	return shardIDs, nil
}

func resolveDBaaSv2ClickhouseShardNames(
	datastore *dbaas_v2_ch.DatastoreResponse,
	shardIDs []string,
) ([]string, error) {
	shardNameByID := make(map[string]string, len(datastore.NodeGroups))
	shardNames := make([]string, 0, len(shardIDs))

	for _, ng := range datastore.NodeGroups {
		if ng.Role == dbaas_v2_ch.NodeGroupRoleData {
			shardNameByID[ng.ID] = ng.Name
		}
	}

	for _, shardID := range shardIDs {
		shardName, exists := shardNameByID[shardID]
		if !exists {
			return nil, fmt.Errorf("shard %s not found in datastore %s", shardID, datastore.ID)
		}

		shardNames = append(shardNames, shardName)
	}

	return shardNames, nil
}

func getDBaaSV2ClickhouseShardGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	shardGroupID string,
) (*dbaas_v2_ch.ShardGroupResponse, error) {
	// no endpoint to get by id
	shardGroups, err := client.ClickHouse.GetShardGroups(ctx, datastoreID)
	if err != nil {
		return nil, fmt.Errorf("error getting shard group %s for datastore %s: %w", shardGroupID, datastoreID, err)
	}

	for _, shardGroup := range shardGroups {
		if shardGroup.ID == shardGroupID {
			return &shardGroup, nil
		}
	}

	return nil, nil
}

func resourceDBaaSV2ClickhouseShardGroupImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("INFRA_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return nil, errors.New("INFRA_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"invalid import ID %q, expected <datastore_id>/<shard_group_id>",
			d.Id(),
		)
	}

	datastoreID := parts[0]
	shardGroupID := parts[1]

	if err := d.Set("datastore_id", datastoreID); err != nil {
		return nil, err
	}

	d.SetId(shardGroupID)

	return []*schema.ResourceData{d}, nil
}
