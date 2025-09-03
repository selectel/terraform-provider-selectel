package selectel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/dbaas-go"
)

func resourceDBaaSDatastoreSecurityGroupV1() *schema.Resource {
	return &schema.Resource{
		UpdateContext: resourceDBaaSDatastoreSecurityGroupV1Update,
		ReadContext:   resourceDBaaSDatastoreSecurityGroupV1Read,
		DeleteContext: resourceDBaaSDatastoreSecurityGroupV1Delete,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsUUID,
				},
			},
		},
	}
}

func resourceDBaaSDatastoreSecurityGroupV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	datastore, err := dbaasClient.Datastore(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectDatastore, d.Id(), err))
	}
	d.Set("project_id", datastore.ProjectID)
	d.Set("security_groups", datastore.SecurityGroups)
	return nil
}

func resourceDBaaSDatastoreSecurityGroupV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	securityGroupsSet, ok := d.Get("security_groups").(*schema.Set)
	if !ok {
		err := fmt.Errorf("is not a list of string")
		return diag.FromErr(errParseDatastoreV1SecurityGroups(err))
	}

	securityGroups, err := resourceDBaaSDatastoreV1SecurityGroupsFromSet(securityGroupsSet)
	if err != nil {
		return diag.FromErr(errParseDatastoreV1SecurityGroups(err))
	}

	securityGroupsOpts := dbaas.DatastoreSecurityGroupOpts{
		SecurityGroups: securityGroups,
	}

	log.Printf("[DEBUG] updating datastore %q security groups %v", d.Id(), securityGroupsOpts)
	_, error_str := dbaasClient.UpdateSecurityGroup(ctx, d.Id(), securityGroupsOpts)
	if error_str != nil {
		return diag.FromErr(error_str)
	}

	return nil
}

func resourceDBaaSDatastoreSecurityGroupV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Diagnostics{}
}
