package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/go-selvpcclient/v2/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
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
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
					uz1Region,
					nl1Region,
				}, false),
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
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectToken, err))
	}

	region := d.Get("region").(string)
	endpoint := getMKSClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

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
