package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	dbaas_v2 "github.com/selectel/dbaas-go/v2"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

const clickhouseDatastoreType = "clickhouse"

func getDBaaSV2Client(d *schema.ResourceData, meta any) (*dbaas_v2.API, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for dbaas: %w", err))
	}

	err = validateRegion(selvpcClient, DBaaSv2, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(DBaaSv2, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get endpoint to init dbaas v2 client: %w", err))
	}

	client, err := dbaas.NewDBAASClientV2WithRetry(selvpcClient.GetXAuthToken(), endpoint.URL, dbaas_v2.RetryConfig{MaxRetries: 3})
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't create dbaas v2 client: %w", err))
	}

	return client, nil
}

func validateDBaaSV2DatastoreType(ctx context.Context, expectedDatastoreTypeEngines []string, typeID string, client *dbaas_v2.API) diag.Diagnostics {
	// no endpoint to get a datastore type in v2

	response, err := client.DatastoreType.GetDatastoreTypeList(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("getting datastore type list: %w", err))
	}

	if response.Errors != "" {
		log.Printf("[WARN] datastore types got with error: %s", response.Errors)
	}

	var datastoreType *dbaas_v2_common.DatastoreTypeResponse

	for _, dt := range response.DatastoreTypes {
		if dt.ID == typeID {
			datastoreType = &dt
			break
		}
	}

	if datastoreType == nil {
		return diag.FromErr(errors.New("Couldnt get datastore type with id " + typeID))
	}
	if !containDatastoreType(expectedDatastoreTypeEngines, datastoreType.Engine) {
		return diag.FromErr(errors.New(buildDatastoreTypeErrorMessage(expectedDatastoreTypeEngines, datastoreType.Engine)))
	}

	return nil
}

func flattenDBaaSV2DatastoreClickhouseNodeGroups(nodeGroups []dbaas_v2_ch.NodeGroupResponse) []any {
	flattenedNodeGroups := make([]any, len(nodeGroups))
	for i, ng := range nodeGroups {
		flattenedInstances := make([]any, len(ng.Instances))
		for j, instance := range ng.Instances {
			flattenedInstance := map[string]any{
				"id":                instance.ID,
				"ip":                instance.IP,
				"floating_ip":       instance.FloatingIP,
				"availability_zone": instance.AvailabilityZone,
				"hostname":          instance.Hostname,
			}
			flattenedInstances[j] = flattenedInstance
		}

		flattenedNG := map[string]any{
			"id":             ng.ID,
			"name":           ng.Name,
			"role":           ng.Role,
			"node_count":     ng.NodeCount,
			"weight":         ng.Weight,
			"has_public_ips": ng.HasPublicIPs,
			"status":         ng.Status,
			"flavor":         flattenDBaaSV2ClickhouseNodeGroupFlavor(ng.Flavor),
			"instances":      flattenedInstances,
		}

		flattenedNodeGroups[i] = flattenedNG
	}

	return flattenedNodeGroups
}

func flattenDBaaSV2ClickhouseNodeGroupFlavor(f dbaas_v2_ch.FlavorResponse) []any {
	if f.Type == dbaas_v2_common.FlavorTypeFlexible {
		return []any{
			map[string]any{
				"type":      f.Type,
				"disk":      f.Disk,
				"ram":       f.RAM,
				"vcpus":     f.VCPUs,
				"disk_type": f.DiskType,
			},
		}
	}
	// for FIXED
	return []any{
		map[string]any{
			"id":   f.ID,
			"type": f.Type,
		},
	}
}

func expandDBaasV2ClickhouseNodeGroupsCreate(nodeGroups []any) []dbaas_v2_ch.NodeGroupCreateRequest {
	result := make([]dbaas_v2_ch.NodeGroupCreateRequest, 0, len(nodeGroups))

	for _, rawGroup := range nodeGroups {
		result = append(result, expandDBaaSV2ClickhouseNodeGroupCreate(rawGroup))
	}

	return result
}

func expandDBaaSV2ClickhouseNodeGroupCreate(raw any) dbaas_v2_ch.NodeGroupCreateRequest {
	ng := raw.(map[string]any)

	req := dbaas_v2_ch.NodeGroupCreateRequest{
		Name:      ng["name"].(string),
		Role:      dbaas_v2_ch.NodeGroupRole(ng["role"].(string)),
		NodeCount: ng["node_count"].(int),
		Flavor:    expandDBaaSV2ClickhouseNodeGroupFlavor(ng["flavor"]),
	}

	if weight, ok := ng["weight"]; ok && weight != 0 {
		w := weight.(int)
		req.Weight = &w
	}

	if hasFIP, ok := ng["has_public_ips"]; ok && hasFIP != nil {
		h := hasFIP.(bool)
		req.HasPublicIPs = &h
	}

	return req
}

func expandDBaaSV2ClickhouseNodeGroupFlavor(raw any) dbaas_v2_ch.FlavorForNodeGroupRequest {
	flavors := raw.([]any)
	if len(flavors) == 0 {
		return dbaas_v2_ch.FlavorForNodeGroupRequest{}
	}

	data := flavors[0].(map[string]any)
	diskType := data["disk_type"].(string)
	flavorType := data["type"].(string)

	if flavorType == string(dbaas_v2_common.FlavorTypeFIXED) && diskType == "" {
		diskType = string(dbaas_v2_common.FlavorDiskLocal)
	}

	return dbaas_v2_ch.FlavorForNodeGroupRequest{
		ID:       data["id"].(string),
		Type:     dbaas_v2_common.FlavorType(flavorType),
		Disk:     data["disk"].(int),
		DiskType: dbaas_v2_common.FlavorDiskType(diskType),
		RAM:      data["ram"].(int),
		VCPUs:    data["vcpus"].(int),
	}
}

func expandDBaaSV2ClickhouseDatastoreLogPlatform(raw any) (dbaas_v2_ch.DatastoreLogGroup, error) {
	var res dbaas_v2_ch.DatastoreLogGroup

	logPlatform := raw.([]any)
	if len(logPlatform) == 0 {
		return res, errors.New("log_group is not set")
	}

	logGroup := logPlatform[0].(map[string]any)
	res.LogGroup = logGroup["log_group"].(string)

	return res, nil
}

func expandDBaaSV2DatastoreSecurityGroupsFromSet(securityGroupsSet *schema.Set) []string {
	result := make([]string, 0, securityGroupsSet.Len())
	for _, value := range securityGroupsSet.List() {
		result = append(result, value.(string))
	}

	return result
}

func expandDBaaSV2ClickhouseShardNameFromSet(shardNamesSet *schema.Set) []string {
	if shardNamesSet == nil {
		return nil
	}

	result := make([]string, 0, shardNamesSet.Len())
	for _, value := range shardNamesSet.List() {
		result = append(result, value.(string))
	}

	return result
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

func createDBaaSV2ClickhouseNodeGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupData any,
	timeout time.Duration,
) error {
	createOpts := expandDBaaSV2ClickhouseNodeGroupCreate(nodeGroupData)
	extraMsg := fmt.Sprintf("create node group %+v", createOpts)

	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))
	_, err := client.ClickHouse.CreateNodeGroup(ctx, datastoreID, createOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

func deleteDBaaSV2ClickhouseNodeGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	timeout time.Duration,
) error {
	extraMsg := fmt.Sprintf("delete node group %s", nodeGroupID)
	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))
	err := client.ClickHouse.DeleteNodeGroup(ctx, datastoreID, nodeGroupID)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

func resizeDBaaSV2ClickhouseNodeGroup(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	resizeData dbaas_v2_ch.NodeGroupResizeRequest,
	timeout time.Duration,
) error {
	extraMsg := fmt.Sprintf("resize node group %s: %+v", nodeGroupID, resizeData)

	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))

	_, err := client.ClickHouse.ResizeNodeGroup(ctx, datastoreID, nodeGroupID, resizeData)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

func resizeDBaaSV2ClickhouseNodeGroupDeleteInstances(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	resizeData dbaas_v2_ch.NodeGroupDeleteInstancesRequest,
	timeout time.Duration,
) error {
	extraMsg := fmt.Sprintf("resize node group %s: %+v", nodeGroupID, resizeData)

	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))

	_, err := client.ClickHouse.DeleteNodeGroupInstances(ctx, datastoreID, nodeGroupID, resizeData)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

func updateDBaaSV2ClickhouseNodeGroupPublicIPs(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	hasPublicIPs bool,
	timeout time.Duration,
) error {
	updateOpts := dbaas_v2_ch.NodeGroupUpdateFloatingIPsRequest{
		HasPublicIPs: hasPublicIPs,
	}
	extraMsg := fmt.Sprintf("update public IPs for node group %s: %+v", nodeGroupID, updateOpts)

	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))

	_, err := client.ClickHouse.UpdateNodeGroupFloatingIPs(ctx, datastoreID, nodeGroupID, updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

func updateDBaaSV2ClickhouseNodeGroupWeight(
	ctx context.Context,
	client *dbaas_v2.API,
	datastoreID string,
	nodeGroupID string,
	weight int,
	timeout time.Duration,
) error {
	updateOpts := dbaas_v2_ch.NodeGroupUpdateWeightRequest{
		Weight: weight,
	}
	extraMsg := fmt.Sprintf("update weight for node group %s: %+v", nodeGroupID, updateOpts)

	log.Print(msgUpdate(objectDatastore, datastoreID, extraMsg))

	_, err := client.ClickHouse.UpdateNodeGroupWeight(ctx, datastoreID, nodeGroupID, updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	err = waiters.WaitForDBaaSV2DatastoreRunningActive(ctx, client.ClickHouse, datastoreID, timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, datastoreID, err)
	}

	return nil
}

type dbaasV2ConfigurationParameterSearchFilter struct {
	datastoreTypeID string
	name            string
}

func expandDBaaSV2ConfigurationParameterSearchFilter(filterSet *schema.Set) dbaasV2ConfigurationParameterSearchFilter {
	filter := dbaasV2ConfigurationParameterSearchFilter{}
	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

	datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]
	if ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	return filter
}
