package selectel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

func dataSourceComputeVolumeV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComputeVolumeV3Read,
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
			"volume_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name", "status", "metadata"},
				ValidateFunc:  validation.StringIsNotWhiteSpace,
			},
			"name": {
				// The selected volume's name is also available as a search criterion.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"status": {
				// The selected volume's status is also available as a search criterion.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"metadata": {
				// The selected volume's metadata is also available as a search criterion.
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: validateComputeVolumeV3Metadata,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootable": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_vol_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceComputeVolumeV3Read(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	client, diagnostics := getBlockStorageClient(d, meta)
	if diagnostics.HasError() {
		return diagnostics
	}

	return readComputeVolumeLookupV3(ctx, d, client)
}

func readComputeVolumeLookupV3(
	ctx context.Context,
	d *schema.ResourceData,
	client *blockstorage.Client,
) diag.Diagnostics {
	var selected *volume.View

	if volumeID, ok := d.GetOk("volume_id"); ok {
		var err error

		selected, _, err = volume.Get(ctx, client, volumeID.(string))
		if err != nil {
			if blockstorage.IsKind(err, blockstorage.KindNotFound) {
				return diag.Errorf(
					"Block Storage volume %q does not exist or is not accessible to the current role: %v",
					volumeID,
					err,
				)
			}

			return blockStorageOperationDiagnostics("read the volume by ID", err)
		}
	} else {
		volumes, err := volume.ListDetail(ctx, client, volume.ListOpts{})
		if err != nil {
			return blockStorageOperationDiagnostics("read the complete volume list", err)
		}

		matches := make([]*volume.View, 0, 1)
		for i := range volumes {
			if matchesComputeVolumeLookupV3(d, &volumes[i]) {
				matches = append(matches, &volumes[i])
			}
		}

		switch len(matches) {
		case 0:
			return diag.Errorf("no Block Storage volumes matched the configured search criteria; " +
				"check project_id, region, access permissions, and any name, status, or metadata filters. " +
				"To look up a known volume_id, remove the conflicting filters")
		case 1:
			selected = matches[0]
		default:
			return diag.Errorf(
				"found %d Block Storage volumes matching the configured search criteria; "+
					"use volume_id for an unambiguous lookup",
				len(matches),
			)
		}
	}

	if err := setComputeVolumeLookupV3(d, selected); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func matchesComputeVolumeLookupV3(d *schema.ResourceData, candidate *volume.View) bool {
	if name, ok := d.GetOk("name"); ok && candidate.Name != name.(string) {
		return false
	}
	if status, ok := d.GetOk("status"); ok && candidate.Status != status.(string) {
		return false
	}
	if metadata, ok := d.GetOk("metadata"); ok {
		for key, value := range expandComputeVolumeV3StringMap(metadata) {
			candidateValue, exists := candidate.Metadata[key]
			if !exists || candidateValue != value {
				return false
			}
		}
	}

	return true
}

func setComputeVolumeLookupV3(d *schema.ResourceData, selected *volume.View) error {
	if err := setComputeVolumeV3(d, selected); err != nil {
		return err
	}
	if err := d.Set("snapshot_id", selected.SnapshotID); err != nil {
		return fmt.Errorf("failed to set Block Storage volume snapshot_id: %w", err)
	}
	if err := d.Set("source_vol_id", selected.SourceVolID); err != nil {
		return fmt.Errorf("failed to set Block Storage volume source_vol_id: %w", err)
	}
	if err := d.Set("status", selected.Status); err != nil {
		return fmt.Errorf("failed to set Block Storage volume status: %w", err)
	}
	if err := d.Set("bootable", selected.Bootable); err != nil {
		return fmt.Errorf("failed to set Block Storage volume bootable: %w", err)
	}

	d.SetId(selected.ID)

	return nil
}
