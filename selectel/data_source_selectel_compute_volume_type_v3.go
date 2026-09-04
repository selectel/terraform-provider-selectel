package selectel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
)

func dataSourceComputeVolumeTypeV3() *schema.Resource {
	const (
		volumeTypeIDAttribute = "volume_type_id"
		nameAttribute         = "name"
	)

	return &schema.Resource{
		ReadContext: dataSourceComputeVolumeTypeV3Read,
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
			volumeTypeIDAttribute: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{volumeTypeIDAttribute, nameAttribute},
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			nameAttribute: {
				// The API returns the name when the volume type is selected by ID.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{volumeTypeIDAttribute, nameAttribute},
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"extra_specs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"supports_custom_iops": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceComputeVolumeTypeV3Read(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	client, diagnostics := getBlockStorageClient(d, meta)
	if diagnostics.HasError() {
		return diagnostics
	}

	var selected *volumetype.View

	if volumeTypeID, ok := d.GetOk("volume_type_id"); ok {
		volumeTypeID := volumeTypeID.(string)
		var err error

		selected, _, err = volumetype.Get(ctx, client, volumeTypeID)
		if err != nil {
			if blockstorage.IsKind(err, blockstorage.KindNotFound) {
				return diag.Errorf(
					"Block Storage volume type %q does not exist or is not accessible in the selected project and region; "+
						"check volume_type_id, project_id, region, and access permissions: %v",
					volumeTypeID,
					err,
				)
			}

			return blockStorageOperationDiagnostics("read the volume type by ID", err)
		}
	} else {
		name := d.Get("name").(string)
		volumeTypes, err := volumetype.List(ctx, client, volumetype.ListOpts{})
		if err != nil {
			return blockStorageOperationDiagnostics("read the complete volume type list", err)
		}

		matches := make([]*volumetype.View, 0, 1)
		for i := range volumeTypes {
			if volumeTypes[i].Name == name {
				matches = append(matches, &volumeTypes[i])
			}
		}

		switch len(matches) {
		case 0:
			return diag.Errorf(
				"Block Storage volume type %q was not found: the name is unknown, or it is a regional "+
					"volume type name that can only be used when creating a volume together with an availability zone",
				name,
			)
		case 1:
			selected = matches[0]
		default:
			return diag.Errorf(
				"found %d Block Storage volume types named %q; use volume_type_id for an unambiguous lookup",
				len(matches),
				name,
			)
		}
	}

	qosLimits, _, err := volumetype.ListQoSLimits(ctx, client)
	if err != nil {
		return blockStorageOperationDiagnostics("read volume type QoS capabilities", err)
	}

	if err := setComputeVolumeTypeV3(d, selected, supportsCustomIOPS(selected.ID, qosLimits)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setComputeVolumeTypeV3(
	d *schema.ResourceData,
	selected *volumetype.View,
	supportsCustomIOPS bool,
) error {
	if err := d.Set("name", selected.Name); err != nil {
		return fmt.Errorf("failed to set Block Storage volume type name: %w", err)
	}
	if err := d.Set("description", selected.Description); err != nil {
		return fmt.Errorf("failed to set Block Storage volume type description: %w", err)
	}
	if err := d.Set("is_public", selected.IsPublic); err != nil {
		return fmt.Errorf("failed to set Block Storage volume type visibility: %w", err)
	}
	if err := d.Set("extra_specs", selected.ExtraSpecs); err != nil {
		return fmt.Errorf("failed to set Block Storage volume type extra_specs: %w", err)
	}
	if err := d.Set("supports_custom_iops", supportsCustomIOPS); err != nil {
		return fmt.Errorf("failed to set Block Storage volume type custom IOPS capability: %w", err)
	}

	d.SetId(selected.ID)

	return nil
}

func supportsCustomIOPS(volumeTypeID string, limits []volumetype.QoSLimitsView) bool {
	for i := range limits {
		if limits[i].VolumeTypeID == volumeTypeID && limits[i].AllowUserQoS && limits[i].FullQoSDiskType {
			return true
		}
	}

	return false
}
