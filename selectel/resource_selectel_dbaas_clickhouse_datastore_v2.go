package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2 "github.com/selectel/dbaas-go/v2"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSV2ClickhouseDatastore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSV2ClickhouseDatastoreCreate,
		ReadContext:   resourceDBaaSV2ClickhouseDatastoreRead,
		UpdateContext: resourceDBaaSV2ClickhouseDatastoreUpdate,
		DeleteContext: resourceDBaaSV2ClickhouseDatastoreDelete,
		CustomizeDiff: customdiff.All(
			validateDBaaSV2ClickhouseDatastoreDiff,
		),
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSV2ClickhouseDatastoreImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute), // resize could take more time if node group with large disk
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

	typeID := d.Get("type_id").(string)
	diagErr = validateDBaaSV2DatastoreType(ctx, []string{clickhouseDatastoreType}, typeID, dbaasClient)
	if diagErr != nil {
		return diagErr
	}

	nodeGroups := expandDBaasV2ClickhouseNodeGroupsCreate(d.Get("node_group").([]any))

	datastoreCreateOpts := dbaas_v2_ch.DatastoreCreateRequest{
		Name:       d.Get("name").(string),
		TypeID:     typeID,
		SubnetID:   d.Get("subnet_id").(string),
		Config:     d.Get("config").(map[string]any),
		NodeGroups: nodeGroups,
	}

	sgRaw, sgOk := d.GetOk("security_groups")
	if sgOk {
		sgSet := sgRaw.(*schema.Set)
		datastoreCreateOpts.SecurityGroups = expandDBaaSV2DatastoreSecurityGroupsFromSet(sgSet)
	}

	logPlatform, logOk := d.GetOk("log_platform")
	if logOk {
		logPlatform, err := expandDBaaSV2ClickhouseDatastoreLogPlatform(logPlatform)
		if err != nil {
			return diag.FromErr(errParseDatastoreV2LogPlatform(err))
		}
		datastoreCreateOpts.LogPlatform = &logPlatform
	}

	log.Print(msgCreate(objectDatastore, datastoreCreateOpts))
	// do after log to avoid exposing the password
	datastoreCreateOpts.Password = d.Get("password").(string)

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

	d.Set("allow_reduce_nodes", d.Get("allow_reduce_nodes"))

	return resourceDBaaSV2ClickhouseDatastoreRead(ctx, d, meta)
}

func resourceDBaaSV2ClickhouseDatastoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatastore, d.Id()))
	datastore, err := dbaasClient.ClickHouse.GetDatastore(ctx, d.Id())

	var dbaasError *dbaas_v2.DBaaSAPIError
	if err != nil {
		if errors.As(err, &dbaasError) && dbaasError.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errGettingObject(objectDatastore, d.Id(), err))
	}

	d.Set("name", datastore.Name)
	d.Set("status", datastore.Status)
	d.Set("state", datastore.State)
	d.Set("project_id", datastore.ProjectID)
	d.Set("subnet_id", datastore.SubnetID)
	d.Set("type_id", datastore.TypeID)
	d.Set("security_groups", datastore.SecurityGroups)

	if datastore.LogPlatform.LogGroup != "" {
		d.Set("log_platform", []any{
			map[string]any{
				"log_group": datastore.LogPlatform.LogGroup,
			},
		})
	}

	// sort node goups from api as in config
	apiNodeGroups := flattenDBaaSV2DatastoreClickhouseNodeGroups(datastore.NodeGroups)
	apiNodeGroupsMap := clickhouseNodeGroupsByName(apiNodeGroups)
	configNodeGroups := d.Get("node_group").([]any)
	sortedNodeGroups := make([]any, 0, len(apiNodeGroups))

	for _, ng := range configNodeGroups {
		ngMap := ng.(map[string]any)
		name := ngMap["name"].(string)

		if apiNG, found := apiNodeGroupsMap[name]; found {
			sortedNodeGroups = append(sortedNodeGroups, apiNG)
			// delete ng which was handeled
			delete(apiNodeGroupsMap, name)
		}
	}
	// add an api node group that is not in the HCL (not created using Terraform)
	for _, apiGroup := range apiNodeGroupsMap {
		sortedNodeGroups = append(sortedNodeGroups, apiGroup)
	}

	if err := d.Set("node_group", sortedNodeGroups); err != nil {
		log.Print(errSettingComplexAttr("node_group", err))
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

func resourceDBaaSV2ClickhouseDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}
	timeout := d.Timeout(schema.TimeoutUpdate)

	if d.HasChange("name") {
		if err := updateDBaaSV2ClickhouseDatastoreName(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		if err := updateDBaaSV2ClickhouseDatastorePassword(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	allowReduceNodes := d.Get("allow_reduce_nodes")
	d.Set("allow_reduce_nodes", allowReduceNodes)

	if d.HasChange("node_group") {
		oldRaw, newRaw := d.GetChange("node_group")

		oldGroups := oldRaw.([]any)
		newGroups := newRaw.([]any)

		if err := reconcileDBaaSV2ClickhouseNodeGroups(
			ctx,
			dbaasClient,
			d.Id(),
			oldGroups,
			newGroups,
			timeout,
			allowReduceNodes.(bool),
		); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("config") {
		err := updateDBaaSv2ClickhouseDatastoreConfig(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("security_groups") {
		if err := updateDBaaSV2ClickhouseDatastoreSecurityGroups(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("log_platform") {
		if err := updateDBaaSV2ClickhouseDatastoreLogPlatform(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDBaaSV2ClickhouseDatastoreRead(ctx, d, meta)
}

func reconcileDBaaSV2ClickhouseNodeGroups(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	oldGroups []any,
	newGroups []any,
	timeout time.Duration,
	allowReduceNodes bool,
) error {
	oldByName := clickhouseNodeGroupsByName(oldGroups)
	newByName := clickhouseNodeGroupsByName(newGroups)

	// Create / update.
	for name, newGroup := range newByName {
		oldGroup, exists := oldByName[name]

		if !exists {
			log.Printf("[DEBUG] creating node group %s for datastore %s", name, datastoreID)

			if err := createDBaaSV2ClickhouseNodeGroup(
				ctx, client, datastoreID, newGroup, timeout,
			); err != nil {
				return fmt.Errorf("creating node group error: %w", err)
			}

			continue
		}

		oldID := oldGroup["id"].(string)

		if err := reconcileDBaaSV2ClickhouseNodeGroup(
			ctx, client, datastoreID, oldID, oldGroup, newGroup, timeout, allowReduceNodes); err != nil {
			return fmt.Errorf("reconciliation node group error: %w", err)
		}
	}

	// Delete.
	for name, oldGroup := range oldByName {
		if _, exists := newByName[name]; exists {
			continue
		}

		oldID := oldGroup["id"].(string)

		log.Printf("[DEBUG] deleting node group %s (id=%s) for datastore %s", name, oldID, datastoreID)

		if err := deleteDBaaSV2ClickhouseNodeGroup(
			ctx, client, datastoreID, oldID, timeout,
		); err != nil {
			return fmt.Errorf("deleting node group error: %w", err)
		}
	}

	return nil
}

func reconcileDBaaSV2ClickhouseNodeGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	oldGroup map[string]any,
	newGroup map[string]any,
	timeout time.Duration,
	allowReduceNodes bool,
) error {
	groupName := newGroup["name"].(string)

	// check resize
	oldNodeCount := oldGroup["node_count"].(int)
	newNodeCount := newGroup["node_count"].(int)

	oldFlavor := expandDBaaSV2ClickhouseNodeGroupFlavor(oldGroup["flavor"])
	newFlavor := expandDBaaSV2ClickhouseNodeGroupFlavor(newGroup["flavor"])

	if newNodeCount < oldNodeCount {
		if allowReduceNodes {
			log.Printf("[DEBUG] reducing node group %s (id=%s): %d → %d nodes", groupName, nodeGroupID, oldNodeCount, newNodeCount)

			targetIDs, err := getInstanceIDsToReduceClickhouseNodeGroupCount(
				oldGroup["instances"].([]any), oldNodeCount, newNodeCount,
			)
			if err != nil {
				return fmt.Errorf("node group %s: %w", groupName, err)
			}

			log.Printf("[DEBUG] node group %s: deleting instances: %v", groupName, targetIDs)

			if err := resizeDBaaSV2ClickhouseNodeGroupDeleteInstances(
				ctx,
				client,
				datastoreID,
				nodeGroupID,
				dbaas_v2_ch.NodeGroupDeleteInstancesRequest{
					Instances: targetIDs,
				},
				timeout,
			); err != nil {
				return fmt.Errorf("node group %s has resize (reduce count) error: %w", groupName, err)
			}
		} else {
			return fmt.Errorf(
				"node group %s: use `allow_reduce_nodes = true` to reduce node count", groupName,
			)
		}
	}

	if newNodeCount > oldNodeCount || !equalDBaaSV2ClickhouseFlavor(oldFlavor, newFlavor) {
		log.Printf("[DEBUG] resizing node group %s (id=%s): nodes=%d, flavor=%+v", groupName, nodeGroupID, newNodeCount, newFlavor)

		req := dbaas_v2_ch.NodeGroupResizeRequest{
			NodeCount: newNodeCount,
			Flavor:    newFlavor,
		}
		if err := resizeDBaaSV2ClickhouseNodeGroup(
			ctx,
			client,
			datastoreID,
			nodeGroupID,
			req,
			timeout,
		); err != nil {
			return fmt.Errorf("node group %s has resize error: %w", groupName, err)
		}
	}

	// check fip
	oldHasPublicIPs := oldGroup["has_public_ips"].(bool)
	newHasPublicIPs := newGroup["has_public_ips"].(bool)
	if oldHasPublicIPs != newHasPublicIPs {
		log.Printf("[DEBUG] updating public IPs for node group %s (id=%s): %v → %v", groupName, nodeGroupID, oldHasPublicIPs, newHasPublicIPs)

		if err := updateDBaaSV2ClickhouseNodeGroupPublicIPs(
			ctx,
			client,
			datastoreID,
			nodeGroupID,
			newHasPublicIPs,
			timeout,
		); err != nil {
			return fmt.Errorf("node group %s has update public IPs error: %w", groupName, err)
		}
	}

	// check weight
	oldWeight := oldGroup["weight"].(int)
	newWeight := newGroup["weight"].(int)
	if oldWeight != newWeight {
		log.Printf("[DEBUG] updating weight for node group %s (id=%s): %d → %d", groupName, nodeGroupID, oldWeight, newWeight)

		if err := updateDBaaSV2ClickhouseNodeGroupWeight(
			ctx,
			client,
			datastoreID,
			nodeGroupID,
			newWeight,
			timeout,
		); err != nil {
			return fmt.Errorf("node group %s has update weight error: %w", groupName, err)
		}
	}

	return nil
}

func getInstanceIDsToReduceClickhouseNodeGroupCount(oldInstances []any, oldNodeCount, newNodeCount int) ([]string, error) {
	instancesLen := len(oldInstances)
	if instancesLen != oldNodeCount {
		return nil, fmt.Errorf(
			"can't reduce node count because of the desynchronization between instances and node_count (%d and %d)", instancesLen, oldNodeCount)
	}
	nodeCountToDelete := oldNodeCount - newNodeCount
	if nodeCountToDelete <= 0 {
		return nil, fmt.Errorf("bad node count to reduce %d", nodeCountToDelete)
	}

	startIndex := instancesLen - nodeCountToDelete
	instancesToDelete := oldInstances[startIndex:]

	targetIDs := make([]string, len(instancesToDelete))
	for i, rawInstance := range instancesToDelete {
		if instance, ok := rawInstance.(map[string]any); ok {
			targetIDs[i] = instance["id"].(string)
		} else {
			return nil, errors.New("can't parse instance from state to reduce node count")
		}
	}

	return targetIDs, nil
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

func validateDBaaSV2ClickhouseDatastoreDiff(
	_ context.Context,
	diff *schema.ResourceDiff,
	_ any,
) error {
	rawNewGroups, ok := diff.Get("node_group").([]any)
	if !ok {
		return nil
	}

	seen := make(map[string]map[string]any, len(rawNewGroups))
	for _, rawNewGroup := range rawNewGroups {
		newGroup := rawNewGroup.(map[string]any)

		if err := validateDBaaSV2ClickHouseNodeGroup(newGroup); err != nil {
			return err
		}

		name, _ := newGroup["name"].(string)

		if _, dup := seen[name]; dup {
			return fmt.Errorf("node_group: duplicate group name %q", name)
		}
		seen[name] = newGroup
	}

	if err := validateDBaaSV2ClickhouseNodeGroupsDiff(diff); err != nil {
		return err
	}

	return nil
}

func validateDBaaSV2ClickHouseNodeGroup(group map[string]any) error {
	name := group["name"].(string)
	role := group["role"].(string)
	hasPublicIPs := group["has_public_ips"].(bool)
	weight := group["weight"].(int)
	nodeCount := group["node_count"].(int)

	if name == "" {
		return errors.New("node group with empty name")
	}

	if nodeCount < 1 {
		return fmt.Errorf("node group %q with node count < 1", name)
	}

	switch role {
	case string(dbaas_v2_ch.NodeGroupRoleData):
		if weight <= 0 {
			return fmt.Errorf("node group %q with role DATA must have weight > 0", name)
		}

	case string(dbaas_v2_ch.NodeGroupRoleKeeper):
		if hasPublicIPs {
			return fmt.Errorf("node group %q with role KEEPER cannot have public IPs", name)
		}

		if weight > 0 {
			return fmt.Errorf("node group %q with role KEEPER cannot have weight > 0", name)
		}
	}

	if err := validateDBaaSV2ClickHouseNodeGroupFlavor(group); err != nil {
		return err
	}

	return nil
}

func clickhouseNodeGroupsByName(groups []any) map[string]map[string]any {
	result := make(map[string]map[string]any, len(groups))

	for _, raw := range groups {
		group := raw.(map[string]any)
		name := group["name"].(string)
		result[name] = group
	}

	return result
}

func validateDBaaSV2ClickhouseNodeGroupsDiff(diff *schema.ResourceDiff) error {
	rawOld, rawNew := diff.GetChange("node_group")

	oldGroups, ok := rawOld.([]any)
	if !ok {
		return nil
	}

	newGroups, ok := rawNew.([]any)
	if !ok {
		return nil
	}

	oldByName := clickhouseNodeGroupsByName(oldGroups)
	newByName := clickhouseNodeGroupsByName(newGroups)

	// can't change role for existing group
	for name, oldGroup := range oldByName {
		newGroup, exists := newByName[name]
		if !exists {
			continue
		}

		oldRole, _ := oldGroup["role"].(string)
		newRole, _ := newGroup["role"].(string)

		if oldRole != newRole {
			return fmt.Errorf(
				"node_group: changing role of node group %q is not allowed",
				name,
			)
		}
	}

	if len(oldByName) == len(newByName) {
		for name := range oldByName {
			if _, exists := newByName[name]; !exists {
				return fmt.Errorf(
					"node_group: changing name of node group %q is not allowed",
					name,
				)
			}
		}
	}

	return nil
}

func validateDBaaSV2ClickHouseNodeGroupFlavor(group map[string]any) error {
	rawFlavors := group["flavor"].([]any)
	if len(rawFlavors) == 0 {
		return nil
	}

	flavor := rawFlavors[0].(map[string]any)

	flavorType := flavor["type"].(string)

	switch flavorType {
	case string(dbaas_v2_common.FlavorTypeFIXED):
		if flavor["id"].(string) == "" {
			return errors.New(
				"flavor.id is required for FIXED flavor",
			)
		}

		if flavor["disk"].(int) != 0 ||
			flavor["ram"].(int) != 0 ||
			flavor["vcpus"].(int) != 0 {
			return errors.New(
				"FIXED flavor cannot specify disk, ram or vcpus",
			)
		}

		if flavor["disk_type"].(string) != "" {
			return errors.New(
				"flavor.disk_type cannot be specified for FIXED flavor",
			)
		}

	case string(dbaas_v2_common.FlavorTypeFlexible):
		if flavor["id"].(string) != "" {
			return errors.New(
				"flavor.id cannot be specified for FLEXIBLE flavor",
			)
		}

		if flavor["disk"].(int) <= 0 ||
			flavor["ram"].(int) <= 0 ||
			flavor["vcpus"].(int) <= 0 {
			return errors.New(
				"disk, ram and vcpus must be greater than 0 for FLEXIBLE flavor",
			)
		}

		if flavor["disk_type"].(string) == "" {
			return errors.New(
				"flavor.disk_type is required for FLEXIBLE flavor",
			)
		}
	}

	return nil
}

func equalDBaaSV2ClickhouseFlavor(a, b dbaas_v2_ch.FlavorForNodeGroupRequest) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case dbaas_v2_common.FlavorTypeFIXED:
		return a.ID == b.ID

	case dbaas_v2_common.FlavorTypeFlexible:
		return a.Disk == b.Disk &&
			a.RAM == b.RAM &&
			a.VCPUs == b.VCPUs &&
			a.DiskType == b.DiskType

	default:
		return false
	}
}

func updateDBaaSv2ClickhouseDatastoreConfig(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var configOpts dbaas_v2_ch.DatastoreConfigRequest
	datastore, err := client.ClickHouse.GetDatastore(ctx, d.Id())
	if err != nil {
		return err
	}
	config := d.Get("config").(map[string]any)

	for param := range datastore.Config {
		// if a param was deleted from config by client we will set nil to reset it to default on api side.
		if _, ok := config[param]; !ok {
			config[param] = nil
		}
	}

	configOpts.Config = config

	log.Print(msgUpdate(objectDatastore, d.Id(), configOpts))
	_, err = client.ClickHouse.UpdateDatastoreConfig(ctx, d.Id(), configOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2ClickhouseDatastoreName(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_ch.DatastoreUpdateRequest
	updateOpts.Name = d.Get("name").(string)

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	_, err := client.ClickHouse.UpdateDatastore(ctx, d.Id(), updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2ClickhouseDatastorePassword(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_ch.DatastoreUpdatePasswordRequest
	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	// do after log to avoid exposing the password
	updateOpts.NewPassword = d.Get("password").(string)
	_, err := client.ClickHouse.UpdateDatastorePassword(ctx, d.Id(), updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2ClickhouseDatastoreLogPlatform(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_ch.DatastoreLogPlatformRequest
	var err error

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	rawLogPlatform, ok := d.GetOk("log_platform")
	if ok {
		logGroup, expandErr := expandDBaaSV2ClickhouseDatastoreLogPlatform(rawLogPlatform)
		if expandErr != nil {
			return errUpdatingObject(objectDatastore, d.Id(), expandErr)
		}
		updateOpts.LogPlatform = logGroup
		_, err = client.ClickHouse.EnableLogPlatform(ctx, d.Id(), updateOpts)
	} else {
		err = client.ClickHouse.DisableLogPlatform(ctx, d.Id())
	}

	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2ClickhouseDatastoreSecurityGroups(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	rawSG := d.Get("security_groups")

	securityGroupsSet := rawSG.(*schema.Set)
	// may be use v2 expand
	securityGroups, err := resourceDBaaSDatastoreV1SecurityGroupsFromSet(securityGroupsSet)
	if err != nil {
		return errParseDatastoreV1SecurityGroups(err)
	}

	updateOpts := dbaas_v2_ch.DatastoreSecurityGroupsRequest{
		SecurityGroups: securityGroups,
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))

	if _, err := client.ClickHouse.UpdateDatastoreSecurityGroups(ctx, d.Id(), updateOpts); err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}
