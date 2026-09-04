package selectel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/snapshot"
)

func dataSourceComputeSnapshotV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComputeSnapshotV3Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"region": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"name": {
				// The selected snapshot's name is also available as a search criterion.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"status": {
				// The selected snapshot's status is also available as a search criterion.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"volume_id": {
				// The selected snapshot's volume ID is also available as a search criterion.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceComputeSnapshotV3Read(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	client, diagnostics := getBlockStorageClient(d, meta)
	if diagnostics.HasError() {
		return diagnostics
	}

	return readComputeSnapshotV3(ctx, d, client)
}

func readComputeSnapshotV3(
	ctx context.Context,
	d *schema.ResourceData,
	client *blockstorage.Client,
) diag.Diagnostics {
	listOpts := snapshot.ListOpts{}
	if name, ok := d.GetOk("name"); ok {
		listOpts.Name = name.(string)
	}
	if status, ok := d.GetOk("status"); ok {
		listOpts.Status = status.(string)
	}
	if volumeID, ok := d.GetOk("volume_id"); ok {
		listOpts.VolumeID = volumeID.(string)
	}

	snapshots, err := snapshot.List(ctx, client, listOpts)
	if err != nil {
		return blockStorageOperationDiagnostics("read the complete snapshot list", err)
	}

	if len(snapshots) == 0 {
		return diag.Errorf("no Block Storage snapshots matched the configured search criteria; " +
			"check project_id, region, access permissions, and any name, status, or volume_id filters")
	}

	selected := &snapshots[0]
	if len(snapshots) > 1 {
		if !d.Get("most_recent").(bool) {
			return diag.Errorf(
				"found %d Block Storage snapshots matching the configured search criteria; "+
					"set most_recent to true or use more specific search criteria",
				len(snapshots),
			)
		}

		for i := 1; i < len(snapshots); i++ {
			if snapshots[i].CreatedAt.After(selected.CreatedAt) {
				selected = &snapshots[i]
			}
		}
	}

	if err := setComputeSnapshotV3(d, selected); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setComputeSnapshotV3(d *schema.ResourceData, selected *snapshot.View) error {
	if err := d.Set("name", selected.Name); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot name: %w", err)
	}
	if err := d.Set("description", selected.Description); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot description: %w", err)
	}
	if err := d.Set("volume_id", selected.VolumeID); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot volume_id: %w", err)
	}
	if err := d.Set("status", selected.Status); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot status: %w", err)
	}
	if err := d.Set("size", selected.Size); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot size: %w", err)
	}
	if err := d.Set("metadata", selected.Metadata); err != nil {
		return fmt.Errorf("failed to set Block Storage snapshot metadata: %w", err)
	}

	d.SetId(selected.ID)

	return nil
}
