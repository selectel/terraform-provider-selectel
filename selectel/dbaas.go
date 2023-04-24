package selectel

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/tokens"
)

const (
	ru1DBaaSV1Endpoint = "https://ru-1.dbaas.selcloud.ru/v1"
	ru2DBaaSV1Endpoint = "https://ru-2.dbaas.selcloud.ru/v1"
	ru3DBaaSV1Endpoint = "https://ru-3.dbaas.selcloud.ru/v1"
	ru7DBaaSV1Endpoint = "https://ru-7.dbaas.selcloud.ru/v1"
	ru8DBaaSV1Endpoint = "https://ru-8.dbaas.selcloud.ru/v1"
	ru9DBaaSV1Endpoint = "https://ru-9.dbaas.selcloud.ru/v1"
	uz1DBaaSV1Endpoint = "https://uz-1.dbaas.selcloud.ru/v1"
)

func getDBaaSV1Endpoint(region string) (endpoint string) {
	switch region {
	case ru1Region:
		endpoint = ru1DBaaSV1Endpoint
	case ru2Region:
		endpoint = ru2DBaaSV1Endpoint
	case ru3Region:
		endpoint = ru3DBaaSV1Endpoint
	case ru7Region:
		endpoint = ru7DBaaSV1Endpoint
	case ru8Region:
		endpoint = ru8DBaaSV1Endpoint
	case ru9Region:
		endpoint = ru9DBaaSV1Endpoint
	case uz1Region:
		endpoint = uz1DBaaSV1Endpoint
	}

	return
}

func getDBaaSClient(ctx context.Context, d *schema.ResourceData, meta interface{}) (*dbaas.API, diag.Diagnostics) {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return nil, diag.FromErr((errCreatingObject(objectToken, err)))
	}

	region := d.Get("region").(string)
	endpoint := getDBaaSV1Endpoint(region)
	client, err := dbaas.NewDBAASClient(token.ID, endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}

func stringChecksum(s string) (string, error) {
	h := md5.New() // #nosec
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs), nil
}

func stringListChecksum(s []string) (string, error) {
	sort.Strings(s)
	checksum, err := stringChecksum(strings.Join(s, ""))
	if err != nil {
		return "", err
	}

	return checksum, nil
}

func baseTestAccCheckDBaaSV1EntityExists(ctx context.Context, rs *terraform.ResourceState, testAccProvider *schema.Provider) (*dbaas.API, error) {
	var projectID, endpoint string
	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}
	if region, ok := rs.Primary.Attributes["region"]; ok {
		endpoint = getDBaaSV1Endpoint(region)
	}

	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()

	tokenOpts := tokens.TokenOpts{
		ProjectID: projectID,
	}
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return nil, errCreatingObject(objectToken, err)
	}

	dbaasClient, err := dbaas.NewDBAASClient(token.ID, endpoint)
	if err != nil {
		return nil, err
	}

	return dbaasClient, nil
}

func convertFieldToStringByType(field interface{}) string {
	switch fieldValue := field.(type) {
	case int:
		return strconv.Itoa(fieldValue)
	case float64:
		return strconv.FormatFloat(fieldValue, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(fieldValue), 'f', -1, 32)
	case string:
		return fieldValue
	case bool:
		return strconv.FormatBool(fieldValue)
	default:
		return ""
	}
}

func RandomWithPrefix(name string) string {
	return fmt.Sprintf("%s_%d", name, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}

func flavorSchema() *schema.Resource {
	return resourceDBaaSDatastoreV1().Schema["flavor"].Elem.(*schema.Resource)
}

func flavorHashSetFunc() schema.SchemaSetFunc {
	return schema.HashResource(flavorSchema())
}

func waitForDBaaSDatastoreV1ActiveState(
	ctx context.Context, client *dbaas.API, datastoreID string, timeout time.Duration,
) error {
	pending := []string{
		string(dbaas.StatusPendingCreate),
		string(dbaas.StatusPendingUpdate),
		string(dbaas.StatusResizing),
	}
	target := []string{
		string(dbaas.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    dbaasDatastoreV1StateRefreshFunc(ctx, client, datastoreID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the datastore %s to become 'ACTIVE': %s",
			datastoreID, err)
	}

	return nil
}

func dbaasDatastoreV1StateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Datastore(ctx, datastoreID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}

func dbaasDatastoreV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Datastore(ctx, datastoreID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}

			return nil, "", err
		}

		return d, strconv.Itoa(http.StatusOK), err
	}
}

func resourceDBaaSDatastoreV1FlavorFromSet(flavorSet *schema.Set) (*dbaas.Flavor, error) {
	if flavorSet.Len() == 0 {
		return nil, nil
	}
	var resourceVcpusRaw, resourceRAMRaw, resourceDiskRaw interface{}
	var ok bool

	resourceFlavorMap := flavorSet.List()[0].(map[string]interface{})
	if resourceVcpusRaw, ok = resourceFlavorMap["vcpus"]; !ok {
		return &dbaas.Flavor{}, errors.New("flavor.vcpus value isn't provided")
	}
	if resourceRAMRaw, ok = resourceFlavorMap["ram"]; !ok {
		return &dbaas.Flavor{}, errors.New("flavor.ram value isn't provided")
	}
	if resourceDiskRaw, ok = resourceFlavorMap["disk"]; !ok {
		return &dbaas.Flavor{}, errors.New("flavor.disk value isn't provided")
	}

	resourceVcpus := resourceVcpusRaw.(int)
	resourceRAM := resourceRAMRaw.(int)
	resourceDisk := resourceDiskRaw.(int)

	flavor := &dbaas.Flavor{
		Vcpus: resourceVcpus,
		RAM:   resourceRAM,
		Disk:  resourceDisk,
	}

	return flavor, nil
}

func resourceDBaaSDatastoreV1FlavorToSet(flavor dbaas.Flavor) *schema.Set {
	flavorSet := &schema.Set{
		F: flavorHashSetFunc(),
	}

	flavorSet.Add(map[string]interface{}{
		"vcpus": flavor.Vcpus,
		"ram":   flavor.RAM,
		"disk":  flavor.Disk,
	})

	return flavorSet
}

func resourceDBaaSDatastoreV1FirewallOptsFromSet(firewallSet *schema.Set) (dbaas.DatastoreFirewallOpts, error) {
	if firewallSet.Len() == 0 {
		return dbaas.DatastoreFirewallOpts{IPs: []string{}}, nil
	}

	var resourceIPsRaw interface{}
	var ok bool

	resourceFirewallRaw := firewallSet.List()[0].(map[string]interface{})
	if resourceIPsRaw, ok = resourceFirewallRaw["ips"]; !ok {
		return dbaas.DatastoreFirewallOpts{}, errors.New("firewall.ips value isn't provided")
	}
	resourceIPRaw := resourceIPsRaw.([]interface{})
	var firewall dbaas.DatastoreFirewallOpts
	for _, ip := range resourceIPRaw {
		firewall.IPs = append(firewall.IPs, ip.(string))
	}

	return firewall, nil
}

func resourceDBaaSDatastoreV1RestoreOptsFromSet(restoreSet *schema.Set) (*dbaas.Restore, error) {
	if restoreSet.Len() == 0 {
		return nil, nil
	}
	var resourceDatastoreIDRaw, resourceTargetTimeRaw interface{}
	var ok bool

	resourceRestoreMap := restoreSet.List()[0].(map[string]interface{})
	if resourceDatastoreIDRaw, ok = resourceRestoreMap["datastore_id"]; !ok {
		return &dbaas.Restore{}, errors.New("restore.datastore_id value isn't provided")
	}
	if resourceTargetTimeRaw, ok = resourceRestoreMap["target_time"]; !ok {
		return &dbaas.Restore{}, errors.New("restore.target_time value isn't provided")
	}

	resourceDatastoreID := resourceDatastoreIDRaw.(string)
	resourceTargetTime := resourceTargetTimeRaw.(string)

	restore := &dbaas.Restore{
		DatastoreID: resourceDatastoreID,
		TargetTime:  resourceTargetTime,
	}

	return restore, nil
}

func updateDatastoreName(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	var updateOpts dbaas.DatastoreUpdateOpts
	updateOpts.Name = d.Get("name").(string)

	log.Print(msgUpdate(objectDatastore, d.Id(), updateOpts))
	_, err := client.UpdateDatastore(ctx, d.Id(), updateOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDatastoreFirewall(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	firewallSet := d.Get("firewall").(*schema.Set)
	firewallOpts, err := resourceDBaaSDatastoreV1FirewallOptsFromSet(firewallSet)
	if err != nil {
		return errParseDatastoreV1Firewall(err)
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), firewallOpts))
	_, err = client.FirewallDatastore(ctx, d.Id(), firewallOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func updateDatastoreConfig(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	var configOpts dbaas.DatastoreConfigOpts
	datastore, err := client.Datastore(ctx, d.Id())
	if err != nil {
		return err
	}
	config := d.Get("config").(map[string]interface{})

	for param := range datastore.Config {
		if _, ok := config[param]; !ok {
			config[param] = nil
		}
	}

	configOpts.Config = config

	log.Print(msgUpdate(objectDatastore, d.Id(), configOpts))
	_, err = client.ConfigDatastore(ctx, d.Id(), configOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func resizeDatastore(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	var resizeOpts dbaas.DatastoreResizeOpts
	nodeCount := d.Get("node_count").(int)
	resizeOpts.NodeCount = nodeCount

	flavorID := d.Get("flavor_id")
	flavorRaw := d.Get("flavor")

	flavorSet := flavorRaw.(*schema.Set)
	flavor, err := resourceDBaaSDatastoreV1FlavorFromSet(flavorSet)
	if err != nil {
		return errParseDatastoreV1Resize(err)
	}

	typeID := d.Get("type_id").(string)
	datastoreType, err := client.DatastoreType(ctx, typeID)
	if err != nil {
		return errors.New("Couldnt get datastore type with id" + typeID)
	}
	if datastoreType.Engine == "redis" {
		resizeOpts.Flavor = nil
		resizeOpts.FlavorID = flavorID.(string)
	} else {
		resizeOpts.Flavor = flavor
		resizeOpts.FlavorID = flavorID.(string)
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), resizeOpts))
	_, err = client.ResizeDatastore(ctx, d.Id(), resizeOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func validateDatastoreType(ctx context.Context, expectedDatastoreTypeEngine string, typeID string, client *dbaas.API) diag.Diagnostics {
	datastoreType, err := client.DatastoreType(ctx, typeID)
	if err != nil {
		return diag.FromErr(errors.New("Couldnt get datastore type with id" + typeID))
	}
	if datastoreType.Engine != expectedDatastoreTypeEngine {
		return diag.FromErr(errors.New("Provided datastore type must have a " + expectedDatastoreTypeEngine + " engine, not " + datastoreType.Engine))
	}

	return nil
}

func waitForDBaaSDatabaseV1ActiveState(
	ctx context.Context, client *dbaas.API, databaseID string, timeout time.Duration,
) error {
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

// Databases

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

		return d, strconv.Itoa(http.StatusOK), err
	}
}

func dbaasDatabaseV1LocaleDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// The default locale value - C is the same as null value, so we need to suppress
	if old == "C" && new == "" {
		return true
	}

	return false
}

// Users

func waitForDBaaSUserV1ActiveState(
	ctx context.Context, client *dbaas.API, userID string, timeout time.Duration,
) error {
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
		Refresh:    dbaasUserV1StateRefreshFunc(ctx, client, userID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the user %s to become 'ACTIVE': %s",
			userID, err)
	}

	return nil
}

func dbaasUserV1StateRefreshFunc(ctx context.Context, client *dbaas.API, userID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.User(ctx, userID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}

func dbaasUserV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, userID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.User(ctx, userID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}

			return nil, "", err
		}

		return d, strconv.Itoa(http.StatusOK), err
	}
}

// Slots

func waitForDBaaSLogicalReplicationSlotV1ActiveState(
	ctx context.Context, client *dbaas.API, slotID string, timeout time.Duration) error {
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
		Refresh:    dbaasLogicalReplicationSlotV1StateRefreshFunc(ctx, client, slotID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the slot %s to become 'ACTIVE': %s",
			slotID, err)
	}

	return nil
}

func dbaasLogicalReplicationSlotV1StateRefreshFunc(ctx context.Context, client *dbaas.API, slotID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.LogicalReplicationSlot(ctx, slotID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}

func dbaasLogicalReplicationSlotV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, slotID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.LogicalReplicationSlot(ctx, slotID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}

			return nil, "", err
		}

		return d, strconv.Itoa(http.StatusOK), err
	}
}
