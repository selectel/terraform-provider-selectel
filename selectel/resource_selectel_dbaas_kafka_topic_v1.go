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
	"github.com/selectel/dbaas-go"
)

func resourceDBaaSKafkaTopicV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSTopicV1Create,
		ReadContext:   resourceDBaaSTopicV1Read,
		UpdateContext: resourceDBaaSTopicV1Update,
		DeleteContext: resourceDBaaSTopicV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSTopicV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"datastore_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"partitions": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDBaaSTopicV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	topicCreateOpts := dbaas.TopicCreateOpts{
		DatastoreID: d.Get("datastore_id").(string),
		Name:        d.Get("name").(string),
		Partitions:  uint16(d.Get("partitions").(int)),
	}

	log.Print(msgCreate(objectTopic, topicCreateOpts))
	topic, err := dbaasClient.CreateTopic(ctx, topicCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectTopic, err))
	}

	log.Printf("[DEBUG] waiting for topic %s to become 'ACTIVE'", topic.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForDBaaSTopicV1ActiveState(ctx, dbaasClient, topic.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectTopic, err))
	}

	d.SetId(topic.ID)

	return resourceDBaaSTopicV1Read(ctx, d, meta)
}

func resourceDBaaSTopicV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectTopic, d.Id()))
	topic, err := dbaasClient.Topic(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectTopic, d.Id(), err))
	}
	d.Set("datastore_id", topic.DatastoreID)
	d.Set("name", topic.Name)
	d.Set("partitions", topic.Partitions)
	d.Set("status", topic.Status)

	return nil
}

func resourceDBaaSTopicV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("partitions") {
		partitions := uint16(d.Get("partitions").(int))
		updateOpts := dbaas.TopicUpdateOpts{
			Partitions: partitions,
		}

		log.Print(msgUpdate(objectTopic, d.Id(), updateOpts))
		_, err := dbaasClient.UpdateTopic(ctx, d.Id(), updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectTopic, d.Id(), err))
		}

		log.Printf("[DEBUG] waiting for topic %s to become 'ACTIVE'", d.Id())
		timeout := d.Timeout(schema.TimeoutCreate)
		err = waitForDBaaSTopicV1ActiveState(ctx, dbaasClient, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectTopic, d.Id(), err))
		}
	}

	return resourceDBaaSTopicV1Read(ctx, d, meta)
}

func resourceDBaaSTopicV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectTopic, d.Id()))
	err := dbaasClient.DeleteTopic(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectTopic, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    dbaasTopicV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	log.Printf("[DEBUG] waiting for topic %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the topic %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSTopicV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
