package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/mks-go/pkg/v1/kubeversion"
)

func dataSourceMKSKubeVersionsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMKSKubeVersionsV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"latest_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceMKSKubeVersionsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mksClient, diagErr := getMKSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	mksKubeVersions, _, err := kubeversion.List(ctx, mksClient)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectKubeVersions, err))
	}

	latestVersion, err := parseMKSKubeVersionsV1Latest(mksKubeVersions)
	if err != nil {
		return diag.FromErr(err)
	}
	defaultVersion := parseMKSKubeVersionsV1Default(mksKubeVersions)
	versions := flattenMKSKubeVersionsV1(mksKubeVersions)
	if err := d.Set("versions", versions); err != nil {
		return diag.FromErr(err)
	}

	checksum, err := stringListChecksum(versions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("latest_version", latestVersion)
	d.Set("default_version", defaultVersion)
	d.SetId(checksum)

	return nil
}
