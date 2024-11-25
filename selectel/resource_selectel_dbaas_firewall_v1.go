package selectel

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	schemas "github.com/terraform-providers/terraform-provider-selectel/selectel/schemas/dbaas"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSFirewallV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSFirewallV1Update,
		ReadContext:   resourceDBaaSFirewallV1Read,
		UpdateContext: resourceDBaaSFirewallV1Update,
		DeleteContext: resourceDBaaSFirewallV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSFirewallV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: schemas.ResourceDBaaSFirewallV1Schema(),
	}
}

func resourceDBaaSFirewallV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	rawIPs := d.Get("ips").([]interface{})
	firewallOpts, err := resourceDBaaSDatastoreV1FirewallOptsFromList(rawIPs)
	if err != nil {
		return diag.FromErr(errParseDatastoreV1Firewall(err))
	}

	log.Print(msgUpdate(objectDatastore, datastoreID, firewallOpts))
	_, err = dbaasClient.FirewallDatastore(ctx, datastoreID, firewallOpts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDatastore, datastoreID, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSDatastoreV1ActiveState(ctx, dbaasClient, datastoreID, timeout)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDatastore, datastoreID, err))
	}

	return resourceDBaaSFirewallV1Read(ctx, d, meta)
}

func resourceDBaaSFirewallV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDatastore, datastoreID))
	datastore, err := dbaasClient.Datastore(ctx, datastoreID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, datastoreID, err))
	}

	checksum, err := firewallChecksum(datastore.Firewall, datastoreID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func resourceDBaaSFirewallV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	datastoreID := d.Get("datastore_id").(string)

	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	firewallOpts := getEmptyFirewallOpts()

	log.Print(msgUpdate(objectDatastore, datastoreID, firewallOpts))
	_, err := dbaasClient.FirewallDatastore(ctx, datastoreID, firewallOpts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDatastore, datastoreID, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastoreID)
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSDatastoreV1ActiveState(ctx, dbaasClient, datastoreID, timeout)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDatastore, datastoreID, err))
	}

	return nil
}

func resourceDBaaSFirewallV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func getEmptyFirewallOpts() dbaas.DatastoreFirewallOpts {
	return dbaas.DatastoreFirewallOpts{IPs: []string{}}
}

func resourceDBaaSDatastoreV1FirewallOptsFromSet(firewallSet *schema.Set) (dbaas.DatastoreFirewallOpts, error) {
	if firewallSet.Len() == 0 {
		return getEmptyFirewallOpts(), nil
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

func resourceDBaaSDatastoreV1FirewallOptsFromList(rawList []interface{}) (dbaas.DatastoreFirewallOpts, error) {
	if len(rawList) == 0 {
		return getEmptyFirewallOpts(), nil
	}

	ipsList := make([]string, len(rawList))
	for i := range rawList {
		ipsList[i] = rawList[i].(string)
	}

	var firewall dbaas.DatastoreFirewallOpts
	firewall.IPs = append(firewall.IPs, ipsList...)

	return firewall, nil
}

func firewallChecksum(firewall []dbaas.Firewall, datastoreID string) (string, error) {
	ipsList := make([]string, len(firewall))
	for _, rule := range firewall {
		ipsList = append(ipsList, rule.IP)
	}

	ipsList = append(ipsList, datastoreID)
	checksum, err := stringListChecksum(ipsList)
	if err != nil {
		return "", err
	}

	return checksum, nil
}
