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

func resourceDBaaSRedisDatastoreV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSRedisDatastoreV1Create,
		ReadContext:   resourceDBaaSRedisDatastoreV1Read,
		UpdateContext: resourceDBaaSRedisDatastoreV1Update,
		DeleteContext: resourceDBaaSRedisDatastoreV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSRedisDatastoreV1ImportState,
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
				ForceNew: false,
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
					uz1Region,
				}, false),
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"node_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connections": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"flavor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcpus": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ram": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"firewall": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ips": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: false,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"restore": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"target_time": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"redis_password": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceDBaaSRedisDatastoreV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	flavorID, flavorIDOk := d.GetOk("flavor_id")

	typeID := d.Get("type_id").(string)
	diagErr = validateDatastoreType(ctx, []string{"redis"}, typeID, dbaasClient)
	if diagErr != nil {
		return diagErr
	}

	restoreSet := d.Get("restore").(*schema.Set)
	restore, err := resourceDBaaSDatastoreV1RestoreOptsFromSet(restoreSet)
	if err != nil {
		return diag.FromErr(errParseDatastoreV1Restore(err))
	}

	datastoreCreateOpts := dbaas.DatastoreCreateOpts{
		Name:      d.Get("name").(string),
		TypeID:    typeID,
		SubnetID:  d.Get("subnet_id").(string),
		NodeCount: d.Get("node_count").(int),
		Restore:   restore,
		Config:    d.Get("config").(map[string]interface{}),
	}

	if flavorIDOk {
		datastoreCreateOpts.FlavorID = flavorID.(string)
	}

	redisPassword, redisPasswordOk := d.GetOk("redis_password")
	if redisPasswordOk {
		datastoreCreateOpts.RedisPassword = redisPassword.(string)
	}

	log.Print(msgCreate(objectDatastore, datastoreCreateOpts))
	datastore, err := dbaasClient.CreateDatastore(ctx, datastoreCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", datastore.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, dbaasClient, datastore.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDatastore, err))
	}

	d.SetId(datastore.ID)

	return resourceDBaaSRedisDatastoreV1Read(ctx, d, meta)
}

func resourceDBaaSRedisDatastoreV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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

	flavor := resourceDBaaSDatastoreV1FlavorToSet(datastore.Flavor)
	if err := d.Set("flavor", flavor); err != nil {
		log.Print(errSettingComplexAttr("flavor", err))
	}

	if err := d.Set("connections", datastore.Connection); err != nil {
		log.Print(errSettingComplexAttr("connections", err))
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

func resourceDBaaSRedisDatastoreV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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
	if d.HasChange("node_count") || d.HasChange("flavor_id") {
		err := resizeRedisDatastore(ctx, d, dbaasClient)
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
	if d.HasChange("redis_password") {
		err := updateRedisDatastorePassword(ctx, d, dbaasClient)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDBaaSRedisDatastoreV1Read(ctx, d, meta)
}

func resourceDBaaSRedisDatastoreV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
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
		Refresh:    dbaasDatastoreV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	log.Printf("[DEBUG] waiting for datastore %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the datastore %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSRedisDatastoreV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
