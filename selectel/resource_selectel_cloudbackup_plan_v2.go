package selectel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	cloudbackup "github.com/selectel/cloudbackup-go/pkg/v2"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/cloudbackup"
)

func resourceCloudBackupPlanV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudBackupPlanV2Create,
		ReadContext:   resourceCloudBackupPlanV2Read,
		UpdateContext: resourceCloudBackupPlanV2Update,
		DeleteContext: resourceCloudBackupPlanV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCloudBackupPlanV2ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project identifier in UUID format",
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Human-readable name of the plan",
			},
			"backup_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "full",
				ValidateFunc: validation.StringInSlice([]string{"full", "frequency"}, true),
				Description:  `Backup mode used for this plan. Allowed values: "full", "frequency"`,
			},
			"full_backups_amount": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Maximum number of backups to save in a full plan or full backups in a frequency plan",
			},
			"schedule_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"crontab", "calendar"}, true),
				Description:  `Backup scheduling type. Allowed values: "calendar", "crontab"`,
			},
			"schedule_pattern": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Backup scheduling pattern",
			},
			"resources": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of resources included in the plan",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "UUID of the backed up resource",
									},
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the backed up resource",
									},
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the backed up resource",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceCloudBackupPlanV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	resources, diagErr := readCloudBackupPlanV2Resource(d)
	if diagErr != nil {
		return diagErr
	}

	var (
		name               = d.Get("name").(string)
		backupMode, _      = d.Get("backup_mode").(string)
		fullBackupsAmount  = d.Get("full_backups_amount").(int)
		scheduleType, _    = d.Get("schedule_type").(string)
		schedulePattern, _ = d.Get("schedule_pattern").(string)
	)

	plan := cloudbackup.Plan{
		BackupMode:        backupMode,
		FullBackupsAmount: fullBackupsAmount,
		Name:              name,
		Resources:         resources,
		SchedulePattern:   schedulePattern,
		ScheduleType:      scheduleType,
	}

	createdPlan, _, err := client.PlanCreate(ctx, &plan)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectCloudBackupPlan, err))
	}

	d.SetId(createdPlan.ID)

	diagErr = waiters.WaitForPlanV2StartedState(ctx, client, d.Id(), d.Timeout(schema.TimeoutCreate))
	if diagErr != nil {
		return diagErr
	}

	return resourceCloudBackupPlanV2Read(ctx, d, meta)
}

func readCloudBackupPlanV2Resource(d *schema.ResourceData) ([]*cloudbackup.PlanResource, diag.Diagnostics) {
	rRaw := d.Get("resources")

	rListRaw, ok := rRaw.([]interface{})
	if !ok {
		return nil, diag.Errorf("resources field has unexpected type")
	}

	resourcesRawMap, ok := rListRaw[0].(map[string]interface{})
	if !ok {
		return nil, diag.Errorf("resources list has unexpected type")
	}

	resourcesRaw, ok := resourcesRawMap["resource"].([]interface{})
	if !ok {
		return nil, diag.Errorf("resources.resource field has unexpected type")
	}

	resources := make([]*cloudbackup.PlanResource, 0, len(resourcesRaw))
	for idx, r := range resourcesRaw {
		item, ok := r.(map[string]interface{})
		if !ok {
			return nil, diag.Errorf("resources[%d] has unexpected type", idx)
		}

		id, okID := item["id"].(string)
		if !okID {
			return nil, diag.Errorf("resources[%d].id has unexpected type", idx)
		}

		name, okName := item["name"].(string)
		if !okName {
			return nil, diag.Errorf("resources[%d].name has unexpected type", idx)
		}

		typeVal, okType := item["type"].(string)
		if !okType {
			return nil, diag.Errorf("resources[%d].type has unexpected type", idx)
		}

		resources = append(resources, &cloudbackup.PlanResource{
			ID:   id,
			Name: name,
			Type: typeVal,
		})
	}

	return resources, nil
}

func resourceCloudBackupPlanV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	res, _, err := client.Plan(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectCloudBackupPlan, d.Id(), err))
	}

	if res == nil {
		return diag.FromErr(fmt.Errorf("can't find plan %q", d.Id()))
	}

	d.Set("name", res.Name)
	d.Set("backup_mode", res.BackupMode)
	d.Set("full_backups_amount", res.FullBackupsAmount)
	d.Set("schedule_type", res.ScheduleType)
	d.Set("schedule_pattern", res.SchedulePattern)

	resources := make([]map[string]interface{}, 0, len(res.Resources))
	for _, r := range res.Resources {
		resources = append(resources, map[string]interface{}{
			"id":   r.ID,
			"name": r.Name,
			"type": r.Type,
		})
	}

	d.Set("resources", []interface{}{map[string]interface{}{
		"resource": resources,
	}})

	return nil
}

func resourceCloudBackupPlanV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	resources, diagErr := readCloudBackupPlanV2Resource(d)
	if diagErr != nil {
		return diagErr
	}

	d.Partial(true)

	backupModeChanges := d.HasChange("backup_mode")

	if backupModeChanges {
		return diag.Errorf("backup_mode cannot be changed for existing plan")
	}

	d.Partial(false)

	var (
		name               = d.Get("name").(string)
		fullBackupsAmount  = d.Get("full_backups_amount").(int)
		scheduleType, _    = d.Get("schedule_type").(string)
		schedulePattern, _ = d.Get("schedule_pattern").(string)
	)

	plan := cloudbackup.PlanUpdateRequest{
		FullBackupsAmount: fullBackupsAmount,
		Name:              name,
		Resources:         resources,
		SchedulePattern:   schedulePattern,
		ScheduleType:      scheduleType,
	}

	_, _, err := client.PlanUpdate(ctx, d.Id(), &plan)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectCloudBackupPlan, d.Id(), err))
	}

	diagErr = waiters.WaitForPlanV2StartedState(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if diagErr != nil {
		return diagErr
	}

	return resourceCloudBackupPlanV2Read(ctx, d, meta)
}

func resourceCloudBackupPlanV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	planID := d.Id()

	_, err := client.PlanDelete(ctx, planID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectCloudBackupPlan, d.Id(), err))
	}

	diagErr = waiters.WaitForPlanV2Deleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
	if diagErr != nil {
		return diagErr
	}

	d.SetId("")

	return nil
}

func resourceCloudBackupPlanV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("project_id must be set for the resource import")
	}

	_ = d.Set("project_id", config.ProjectID)

	if config.Region == "" {
		return nil, errors.New("region must be set for the resource import")
	}

	_ = d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}
