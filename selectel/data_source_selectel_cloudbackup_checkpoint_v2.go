package selectel

import (
	"context"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/cloudbackup"
)

func dataSourceCloudBackupCheckpointV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudBackupCheckpointV2Read,
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
						"plan_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"volume_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			// computed
			"checkpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"plan_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"checkpoint_items": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"backup_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"chain_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"checkpoint_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"created_at": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"backup_created_at": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"is_incremental": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"status": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"resource": {
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
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudBackupCheckpointV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getScheduledBackupClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	filter := expandCloudBlockStorageBackupCheckpointsSearchFilter(d)

	log.Printf("[DEBUG] Getting %s '%#v'", objectCloudBackupCheckpoint, filter)

	checkpoints, _, err := client.Checkpoints(ctx, &cloudbackup.CheckpointsQuery{
		PlanName:   filter.planName,
		VolumeName: filter.volumeName,
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(objectCloudBackupCheckpoint, err))
	}

	osFlatten := flattenCloudBlockStorageBackupCheckpoints(checkpoints)
	if err := d.Set("checkpoints", osFlatten); err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(checkpoints))
	for _, e := range checkpoints {
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

type cloudBackupCheckpointsFilter struct {
	planName   string
	volumeName string
}

func expandCloudBlockStorageBackupCheckpointsSearchFilter(d *schema.ResourceData) cloudBackupCheckpointsFilter {
	filter := cloudBackupCheckpointsFilter{}

	filterSet, ok := d.Get("filter").(*schema.Set)
	if !ok {
		return filter
	}

	if filterSet.Len() == 0 {
		return filter
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	planName, ok := resourceFilterMap["plan_name"]
	if ok {
		filter.planName = planName.(string)
	}

	volumeName, ok := resourceFilterMap["volume_name"]
	if ok {
		filter.volumeName = volumeName.(string)
	}

	return filter
}

func flattenCloudBlockStorageBackupCheckpoints(list []*cloudbackup.Checkpoint) []interface{} {
	res := make([]interface{}, len(list))
	for i, e := range list {
		sMap := make(map[string]interface{})
		sMap["id"] = e.ID
		sMap["plan_id"] = e.PlanID
		sMap["created_at"] = e.CreatedAt
		sMap["status"] = e.Status

		items := make([]interface{}, len(e.CheckpointItems))
		for j, item := range e.CheckpointItems {
			itemMap := make(map[string]interface{})
			itemMap["id"] = item.ID
			itemMap["backup_id"] = item.BackupID
			itemMap["chain_id"] = item.ChainID
			itemMap["checkpoint_id"] = item.CheckpointID
			itemMap["created_at"] = item.CreatedAt
			itemMap["backup_created_at"] = item.BackupCreatedAt
			itemMap["is_incremental"] = item.IsIncremental
			itemMap["status"] = item.Status

			resource := make([]interface{}, 1)
			resourceMap := make(map[string]interface{})
			resourceMap["id"] = item.Resource.ID
			resourceMap["name"] = item.Resource.Name
			resourceMap["type"] = item.Resource.Type
			resource[0] = resourceMap
			itemMap["resource"] = resource

			items[j] = itemMap
		}
		sMap["checkpoint_items"] = items

		res[i] = sMap
	}

	return res
}
