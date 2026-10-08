package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2 "github.com/selectel/dbaas-go/v2"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	dbaas_v2_os "github.com/selectel/dbaas-go/v2/opensearch"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSV2OpensearchDatastore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSV2OpensearchDatastoreCreate,
		ReadContext:   resourceDBaaSV2OpensearchDatastoreRead,
		UpdateContext: resourceDBaaSV2OpensearchDatastoreUpdate,
		DeleteContext: resourceDBaaSV2OpensearchDatastoreDelete,
		CustomizeDiff: customdiff.All(
			validateDBaaSV2OpensearchDatastoreDiff,
		),
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSV2OpensearchDatastoreImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute), // resize could take more time if node group with large disk
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSV2OpensearchDatastoreSchema(),
	}
}

func resourceDBaaSV2OpensearchDatastoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	typeID := d.Get("type_id").(string)
	diagErr = validateDBaaSV2DatastoreType(ctx, []string{opensearchDatastoreType}, typeID, dbaasClient)
	if diagErr != nil {
		return diagErr
	}

	nodeGroups := expandDBaasV2OpensearchNodeGroupsCreate(d.Get("node_group").([]any))

	datastoreCreateOpts := dbaas_v2_os.DatastoreCreateRequest{
		Name:       d.Get("name").(string),
		TypeID:     typeID,
		SubnetID:   d.Get("subnet_id").(string),
		NodeGroups: nodeGroups,
	}

	sgRaw, sgOk := d.GetOk("security_groups")
	if sgOk {
		sgSet := sgRaw.(*schema.Set)
		datastoreCreateOpts.SecurityGroups = expandDBaaSV2DatastoreSecurityGroupsFromSet(sgSet)
	}

	logPlatform, logOk := d.GetOk("log_platform")
	if logOk {
		logPlatform, err := expandDBaaSV2OpensearchDatastoreLogPlatform(logPlatform)
		if err != nil {
			return diag.FromErr(errParseDatastoreV2LogPlatform(err))
		}
		datastoreCreateOpts.LogPlatform = &logPlatform
	}

	log.Print(msgCreate(objectDatastore, datastoreCreateOpts))
	// do after log to avoid exposing the password
	datastoreCreateOpts.Password = d.Get("password").(string)

	datastore, err := dbaasClient.Opensearch.CreateDatastore(ctx, datastoreCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(datastore.ID)

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.Opensearch, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	return resourceDBaaSV2OpensearchDatastoreRead(ctx, d, meta)
}

func resourceDBaaSV2OpensearchDatastoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatastore, d.Id()))
	datastore, err := dbaasClient.Opensearch.GetDatastore(ctx, d.Id())

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

	apiNodeGroups := flattenDBaaSV2DatastoreOpensearchNodeGroups(datastore.NodeGroups)
	apiNodeGroupsByID := opensearchNodeGroupsByID(apiNodeGroups)
	apiNodeGroupsByName := opensearchNodeGroupsByName(apiNodeGroups)

	stateGroups := d.Get("node_group").([]any)
	result := make([]any, 0, len(stateGroups))

	for _, raw := range stateGroups {
		st := raw.(map[string]any)

		apiNG := opensearchMatchAPINodeGroup(st, apiNodeGroupsByID, apiNodeGroupsByName)
		if apiNG == nil {
			// the node group was deleted outside of Terraform and is dropped from the state.
			continue
		}
		delete(apiNodeGroupsByID, apiNG["id"].(string))

		elem := opensearchNodeGroupFromAPI(apiNG)
		elem["key"] = st["key"]
		result = append(result, elem)
	}

	// add node groups that were created outside of Terraform.
	for _, apiNG := range apiNodeGroupsByID {
		elem := opensearchNodeGroupFromAPI(apiNG)
		elem["key"] = opensearchImportedNodeGroupKey(apiNG["id"].(string))
		result = append(result, elem)
	}

	if err := d.Set("node_group", result); err != nil {
		log.Print(errSettingComplexAttr("node_group", err))
	}

	return nil
}

func resourceDBaaSV2OpensearchDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}
	timeout := d.Timeout(schema.TimeoutUpdate)

	if d.HasChange("name") {
		if err := updateDBaaSV2OpensearchDatastoreName(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		if err := updateDBaaSV2OpensearchDatastorePassword(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("node_group") {
		oldRaw, newRaw := d.GetChange("node_group")

		oldGroups := oldRaw.([]any)
		newGroups := newRaw.([]any)

		keyToID, err := reconcileDBaaSV2OpensearchNodeGroups(
			ctx,
			dbaasClient,
			d.Id(),
			oldGroups,
			newGroups,
			timeout,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		// rebind ids by key — the final Read is positional
		rewritten := make([]any, 0, len(newGroups))
		for _, raw := range newGroups {
			group := raw.(map[string]any)

			elem := make(map[string]any, len(group)+1)
			maps.Copy(elem, group)

			if id, ok := keyToID[group["key"].(string)]; ok {
				elem["id"] = id
			}

			rewritten = append(rewritten, elem)
		}

		if err := d.Set("node_group", rewritten); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("security_groups") {
		if err := updateDBaaSV2OpensearchDatastoreSecurityGroups(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("log_platform") {
		if err := updateDBaaSV2OpensearchDatastoreLogPlatform(ctx, d, dbaasClient); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDBaaSV2OpensearchDatastoreRead(ctx, d, meta)
}

func updateDBaaSV2OpensearchDatastoreName(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_os.DatastoreUpdateRequest
	updateOpts.Name = d.Get("name").(string)

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	_, err := client.Opensearch.UpdateDatastore(ctx, d.Id(), updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.Opensearch, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2OpensearchDatastorePassword(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_os.DatastoreUpdatePasswordRequest
	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	// do after log to avoid exposing the password
	updateOpts.NewPassword = d.Get("password").(string)
	_, err := client.Opensearch.UpdateDatastorePassword(ctx, d.Id(), updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.Opensearch, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2OpensearchDatastoreLogPlatform(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	var updateOpts dbaas_v2_os.DatastoreLogPlatformRequest
	var err error

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	rawLogPlatform, ok := d.GetOk("log_platform")
	if ok {
		logGroup, expandErr := expandDBaaSV2OpensearchDatastoreLogPlatform(rawLogPlatform)
		if expandErr != nil {
			return errUpdatingObject(objectDatastore, d.Id(), expandErr)
		}
		updateOpts.LogPlatform = logGroup
		_, err = client.Opensearch.EnableLogPlatform(ctx, d.Id(), updateOpts)
	} else {
		err = client.Opensearch.DisableLogPlatform(ctx, d.Id())
	}

	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.Opensearch, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDBaaSV2OpensearchDatastoreSecurityGroups(ctx context.Context, d *schema.ResourceData, client *dbaas_v2.API) error {
	rawSG := d.Get("security_groups")

	securityGroupsSet := rawSG.(*schema.Set)
	securityGroups := expandDBaaSV2DatastoreSecurityGroupsFromSet(securityGroupsSet)

	updateOpts := dbaas_v2_os.DatastoreSecurityGroupsRequest{
		SecurityGroups: securityGroups,
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))

	if _, err := client.Opensearch.UpdateDatastoreSecurityGroups(ctx, d.Id(), updateOpts); err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err := waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.Opensearch, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func reconcileDBaaSV2OpensearchNodeGroups(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	oldGroups []any,
	newGroups []any,
	timeout time.Duration,
) (map[string]string, error) {
	oldByKey := opensearchNodeGroupsByKey(oldGroups)
	newByKey := opensearchNodeGroupsByKey(newGroups)

	keyToID := make(map[string]string, len(newByKey))

	// Create / update.
	for key, newGroup := range newByKey {
		oldGroup, exists := oldByKey[key]

		if !exists {
			createdID, err := createDBaaSV2OpensearchNodeGroup(
				ctx, client, datastoreID, newGroup, timeout,
			)
			if err != nil {
				return nil, fmt.Errorf("creating node group error: %w", err)
			}

			keyToID[key] = createdID

			continue
		}

		oldID := oldGroup["id"].(string)
		keyToID[key] = oldID

		if err := reconcileDBaaSV2OpensearchNodeGroup(
			ctx, client, datastoreID, oldID, oldGroup, newGroup, timeout); err != nil {
			return nil, fmt.Errorf("reconciliation node group error: %w", err)
		}
	}

	// Delete.
	for key, oldGroup := range oldByKey {
		if _, exists := newByKey[key]; exists {
			continue
		}

		oldID := oldGroup["id"].(string)

		if err := deleteDBaaSV2OpensearchNodeGroup(
			ctx, client, datastoreID, oldID, timeout,
		); err != nil {
			return nil, fmt.Errorf("deleting node group error: %w", err)
		}
	}

	return keyToID, nil
}

func reconcileDBaaSV2OpensearchNodeGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	oldGroup map[string]any,
	newGroup map[string]any,
	timeout time.Duration,
) error {
	groupName := newGroup["name"].(string)

	// rename
	oldName := oldGroup["name"].(string)
	if oldName != groupName {
		req := dbaas_v2_os.NodeGroupUpdateRequest{
			Name: groupName,
		}
		if err := updateDBaaSV2OpensearchNodeGroup(
			ctx,
			client,
			datastoreID,
			nodeGroupID,
			req,
			timeout,
		); err != nil {
			return fmt.Errorf("node group %s has rename error: %w", groupName, err)
		}
	}

	// check resize
	oldNodeCount := oldGroup["node_count"].(int)
	newNodeCount := newGroup["node_count"].(int)

	oldFlavor := expandDBaaSV2OpensearchNodeGroupFlavor(oldGroup["flavor"])
	newFlavor := expandDBaaSV2OpensearchNodeGroupFlavor(newGroup["flavor"])

	if newNodeCount != oldNodeCount || !equalDBaaSV2OpensearchFlavor(oldFlavor, newFlavor) {
		req := dbaas_v2_os.NodeGroupResizeRequest{
			NodeCount: newNodeCount,
			Flavor:    newFlavor,
		}
		if err := resizeDBaaSV2OpensearchNodeGroup(
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
		if err := updateDBaaSV2OpensearchNodeGroupPublicIPs(
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

	return nil
}

func resourceDBaaSV2OpensearchDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectDatastore, d.Id()))
	err := dbaasClient.Opensearch.DeleteDatastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatastore, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become deleted", d.Id())
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waiters.WaitForDBaaSV2DatastoreDeleted(ctx, dbaasClient.Opensearch, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDatastore, d.Id(), err))
	}

	return nil
}

func resourceDBaaSV2OpensearchDatastoreImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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

func validateDBaaSV2OpensearchDatastoreDiff(
	_ context.Context,
	diff *schema.ResourceDiff,
	_ any,
) error {
	rawNewGroups := diff.Get("node_group").([]any)

	seenKeys := make(map[string]struct{}, len(rawNewGroups))
	seenNames := make(map[string]struct{}, len(rawNewGroups))
	for _, rawNewGroup := range rawNewGroups {
		newGroup := rawNewGroup.(map[string]any)

		if err := validateDBaaSV2OpensearchNodeGroup(newGroup); err != nil {
			return err
		}

		name, _ := newGroup["name"].(string)

		key, _ := newGroup["key"].(string)
		if _, dup := seenKeys[key]; dup {
			return fmt.Errorf("node_group: duplicate node group key %q", key)
		}
		seenKeys[key] = struct{}{}

		if _, dup := seenNames[name]; dup {
			return fmt.Errorf("node_group: duplicate node group name %q", name)
		}
		seenNames[name] = struct{}{}
	}

	if err := validateDBaaSV2OpensearchNodeGroupsDiff(diff); err != nil {
		return err
	}

	return nil
}

func validateDBaaSV2OpensearchNodeGroup(group map[string]any) error {
	name, _ := group["name"].(string)
	if name == "" {
		return errors.New("node group with empty name")
	}

	nodeCount := group["node_count"].(int)
	if nodeCount < 1 {
		return fmt.Errorf("node group %q with node count < 1", name)
	}

	if err := validateDBaaSV2OpensearchNodeGroupFlavor(group); err != nil {
		return err
	}

	return nil
}

func opensearchNodeGroupsByName(groups []any) map[string]map[string]any {
	result := make(map[string]map[string]any, len(groups))

	for _, raw := range groups {
		group := raw.(map[string]any)
		name := group["name"].(string)
		result[name] = group
	}

	return result
}

func opensearchNodeGroupsByID(groups []any) map[string]map[string]any {
	byID := make(map[string]map[string]any, len(groups))
	for _, raw := range groups {
		g := raw.(map[string]any)
		byID[g["id"].(string)] = g
	}

	return byID
}

func opensearchNodeGroupsByKey(groups []any) map[string]map[string]any {
	result := make(map[string]map[string]any, len(groups))
	for _, raw := range groups {
		g := raw.(map[string]any)
		result[g["key"].(string)] = g
	}

	return result
}

// opensearchMatchAPINodeGroup finds an API node group for a state element:
// by id (update/refresh) or, if id is not set yet, by name (create path).
func opensearchMatchAPINodeGroup(
	st map[string]any,
	byID, byName map[string]map[string]any,
) map[string]any {
	if id, ok := st["id"].(string); ok && id != "" {
		return byID[id]
	}
	if name, ok := st["name"].(string); ok && name != "" {
		return byName[name]
	}

	return nil
}

// opensearchNodeGroupFromAPI copies a flattened API node group without the key.
func opensearchNodeGroupFromAPI(g map[string]any) map[string]any {
	elem := make(map[string]any, len(g)+1)
	maps.Copy(elem, g)

	return elem
}

const opensearchImportedNodeGroupKeyPrefix = "imported-"

func opensearchImportedNodeGroupKey(id string) string {
	return opensearchImportedNodeGroupKeyPrefix + id
}

func validateDBaaSV2OpensearchNodeGroupsDiff(diff *schema.ResourceDiff) error {
	rawOld, rawNew := diff.GetChange("node_group")

	oldGroups, ok := rawOld.([]any)
	if !ok {
		return nil
	}

	newGroups, ok := rawNew.([]any)
	if !ok {
		return nil
	}

	oldByKey := opensearchNodeGroupsByKey(oldGroups)
	newByKey := opensearchNodeGroupsByKey(newGroups)

	for key, oldGroup := range oldByKey {
		newGroup, exists := newByKey[key]
		if !exists {
			continue
		}

		oldRole, _ := oldGroup["role"].(string)
		newRole, _ := newGroup["role"].(string)

		if oldRole != newRole {
			return fmt.Errorf(
				"node_group: changing role of node group with key %q is not allowed",
				key,
			)
		}
	}

	return nil
}

func validateDBaaSV2OpensearchNodeGroupFlavor(group map[string]any) error {
	rawFlavors, ok := group["flavor"].([]any)
	if !ok || len(rawFlavors) == 0 {
		return nil
	}

	flavor, ok := rawFlavors[0].(map[string]any)
	if !ok {
		return nil
	}

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

func equalDBaaSV2OpensearchFlavor(a, b dbaas_v2_os.FlavorForNodeGroupRequest) bool {
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
