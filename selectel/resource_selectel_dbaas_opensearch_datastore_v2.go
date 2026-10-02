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

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, dbaasClient.Opensearch, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(datastore.ID)

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

	// sort node goups from api as in config
	apiNodeGroups := flattenDBaaSV2DatastoreOpensearchNodeGroups(datastore.NodeGroups)
	apiNodeGroupsMap := opensearchNodeGroupsByName(apiNodeGroups)
	configNodeGroups := d.Get("node_group").([]any)
	sortedNodeGroups := make([]any, 0, len(apiNodeGroups))

	for _, ng := range configNodeGroups {
		ngMap := ng.(map[string]interface{})
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

		if err := reconcileDBaaSV2OpensearchNodeGroups(
			ctx,
			dbaasClient,
			d.Id(),
			oldGroups,
			newGroups,
			timeout,
		); err != nil {
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

func reconcileDBaaSV2OpensearchNodeGroups(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	oldGroups []any,
	newGroups []any,
	timeout time.Duration,
) error {

	oldByName := opensearchNodeGroupsByName(oldGroups)
	newByName := opensearchNodeGroupsByName(newGroups)

	// Create / update.
	for name, newGroup := range newByName {
		oldGroup, exists := oldByName[name]

		if !exists {
			if err := createDBaaSV2OpensearchNodeGroup(
				ctx, client, datastoreID, newGroup, timeout,
			); err != nil {
				return fmt.Errorf("creating node group error: %w", err)
			}
			continue
		}

		oldID := oldGroup["id"].(string)

		if err := reconcileDBaaSV2OpensearchNodeGroup(
			ctx, client, datastoreID, oldID, oldGroup, newGroup, timeout); err != nil {
			return fmt.Errorf("reconciliation node group error: %w", err)
		}
	}

	// Delete.
	for name, oldGroup := range oldByName {

		if _, exists := newByName[name]; exists {
			continue
		}

		oldID := oldGroup["id"].(string)

		if err := deleteDBaaSV2OpensearchNodeGroup(
			ctx, client, datastoreID, oldID, timeout,
		); err != nil {
			return fmt.Errorf("deleting node group error: %w", err)
		}
	}

	return nil
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
	ctx context.Context,
	diff *schema.ResourceDiff,
	meta any,
) error {
	rawNewGroups, ok := diff.Get("node_group").([]any)
	if !ok {
		return nil
	}

	seen := make(map[string]map[string]any, len(rawNewGroups))
	for _, rawNewGroup := range rawNewGroups {
		newGroup := rawNewGroup.(map[string]any)

		if err := validateDBaaSV2OpensearchNodeGroup(newGroup); err != nil {
			return err
		}

		name, _ := newGroup["name"].(string)

		if _, dup := seen[name]; dup {
			return fmt.Errorf("node_group: duplicate group name %q", name)
		}
		seen[name] = newGroup
	}

	if err := validateDBaaSV2OpensearchNodeGroupsDiff(diff); err != nil {
		return err
	}
	return nil
}

func validateDBaaSV2OpensearchNodeGroup(group map[string]any) error {
	name := group["name"].(string)
	nodeCount := group["node_count"].(int)

	if name == "" {
		return errors.New("node group with empty name.")
	}

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

	oldByName := opensearchNodeGroupsByName(oldGroups)
	newByName := opensearchNodeGroupsByName(newGroups)

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

func validateDBaaSV2OpensearchNodeGroupFlavor(group map[string]any) error {
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
