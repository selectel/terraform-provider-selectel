package selectel

import (
	"context"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cloudbackup "github.com/selectel/cloudbackup-go/pkg/v2"
)

func dataSourceCloudBackupPlanV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudBackupPlanV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"volume_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"status": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			// computed
			"plans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"full_backups_amount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"schedule_pattern": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schedule_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudBackupPlanV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandCloudBackupPlansSearchFilter(d)

	log.Printf("[DEBUG] Getting %s '%#v'", objectCloudBackupPlan, filter)

	plans, _, err := client.Plans(ctx, &cloudbackup.PlansQuery{
		Name:       filter.name,
		VolumeName: filter.volumeName,
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(objectCloudBackupPlan, err))
	}

	filteredPlans := filterCloudBackupPlans(plans, filter)

	osFlatten := flattenCloudBackupPlans(filteredPlans)
	if err := d.Set("plans", osFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(filteredPlans))
	for _, e := range filteredPlans {
		ids = append(ids, e.ID)
	}

	slices.Sort(ids)

	checksum, err := stringListChecksum(ids)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(checksum)

	return nil
}

type cloudBackupPlansFilter struct {
	name       string
	volumeName string
	status     string
}

func expandCloudBackupPlansSearchFilter(d *schema.ResourceData) cloudBackupPlansFilter {
	filter := cloudBackupPlansFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok {
		return filter
	}

	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	volumeName, ok := resourceFilterMap["volume_name"]
	if ok {
		filter.volumeName = volumeName.(string)
	}

	status, ok := resourceFilterMap["status"]
	if ok {
		filter.status = status.(string)
	}

	return filter
}

func filterCloudBackupPlans(list []*cloudbackup.Plan, filter cloudBackupPlansFilter) []*cloudbackup.Plan {
	filtered := make([]*cloudbackup.Plan, 0, len(list))
	for _, entry := range list {
		if filter.status != "" && entry.Status != filter.status {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func flattenCloudBackupPlans(list []*cloudbackup.Plan) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["backup_mode"] = e.BackupMode
		sMap["created_at"] = e.CreatedAt
		sMap["description"] = e.Description
		sMap["id"] = e.ID
		sMap["full_backups_amount"] = e.FullBackupsAmount
		sMap["name"] = e.Name

		resources := make([]interface{}, len(e.Resources))
		for j, r := range e.Resources {
			rMap := make(map[string]interface{})
			rMap["id"] = r.ID
			rMap["name"] = r.Name
			rMap["type"] = r.Type
			resources[j] = rMap
		}
		sMap["resources"] = resources

		sMap["schedule_pattern"] = e.SchedulePattern
		sMap["schedule_type"] = e.ScheduleType
		sMap["status"] = e.Status

		res[i] = sMap
	}

	return res
}
