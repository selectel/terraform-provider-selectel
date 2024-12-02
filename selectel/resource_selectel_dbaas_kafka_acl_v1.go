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
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func resourceDBaaSKafkaACLV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDBaaSACLV1Create,
		ReadContext:   resourceDBaaSACLV1Read,
		UpdateContext: resourceDBaaSACLV1Update,
		DeleteContext: resourceDBaaSACLV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDBaaSACLV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: resourceDBaaSKafkaACKV1Schema(),
	}
}

func resourceDBaaSACLV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	aclCreateOpts := dbaas.ACLCreateOpts{
		DatastoreID: d.Get("datastore_id").(string),
		UserID:      d.Get("user_id").(string),
		Pattern:     d.Get("pattern").(string),
		PatternType: d.Get("pattern_type").(string),
		AllowRead:   d.Get("allow_read").(bool),
		AllowWrite:  d.Get("allow_write").(bool),
	}

	log.Print(msgCreate(objectACL, aclCreateOpts))
	acl, err := dbaasClient.CreateACL(ctx, aclCreateOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectACL, err))
	}

	log.Printf("[DEBUG] waiting for acl %s to become 'ACTIVE'", acl.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSACLV1ActiveState(ctx, dbaasClient, acl.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectACL, err))
	}

	d.SetId(acl.ID)

	return resourceDBaaSACLV1Read(ctx, d, meta)
}

func resourceDBaaSACLV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectACL, d.Id()))
	acl, err := dbaasClient.ACL(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectACL, d.Id(), err))
	}
	d.Set("datastore_id", acl.DatastoreID)
	if acl.Pattern != "" {
		d.Set("pattern", acl.Pattern)
	}
	d.Set("pattern_type", acl.PatternType)
	d.Set("allow_read", acl.AllowRead)
	d.Set("allow_write", acl.AllowWrite)
	d.Set("status", acl.Status)

	return nil
}

func resourceDBaaSACLV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("allow_read") || d.HasChange("allow_write") {
		updateOpts := dbaas.ACLUpdateOpts{
			AllowRead:  d.Get("allow_read").(bool),
			AllowWrite: d.Get("allow_write").(bool),
		}

		log.Print(msgUpdate(objectACL, d.Id(), updateOpts))
		_, err := dbaasClient.UpdateACL(ctx, d.Id(), updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectACL, d.Id(), err))
		}

		log.Printf("[DEBUG] waiting for acl %s to become 'ACTIVE'", d.Id())
		timeout := d.Timeout(schema.TimeoutCreate)
		err = waiters.WaitForDBaaSACLV1ActiveState(ctx, dbaasClient, d.Id(), timeout)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectACL, d.Id(), err))
		}
	}

	return resourceDBaaSACLV1Read(ctx, d, meta)
}

func resourceDBaaSACLV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectACL, d.Id()))
	err := dbaasClient.DeleteACL(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectACL, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.Itoa(http.StatusOK)},
		Target:     []string{strconv.Itoa(http.StatusNotFound)},
		Refresh:    waiters.DBaaSACLV1DeleteStateRefreshFunc(ctx, dbaasClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	log.Printf("[DEBUG] waiting for acl %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for the acl %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceDBaaSACLV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
