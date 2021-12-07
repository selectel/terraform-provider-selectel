package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/cluster"
)

func dataSourceMKSKubeconfigV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMKSKubeconfigV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"raw_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_ca_cert": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_cert": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMKSKubeconfigV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	clusterID := d.Get("cluster_id").(string)

	mksCluster, _, err := cluster.Get(ctx, mksClient, clusterID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectCluster, clusterID, err))
	}

	parsedKubeconfig, _, err := cluster.GetParsedKubeconfig(ctx, mksClient, mksCluster.ID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectKubeConfig, clusterID, err))
	}

	d.SetId(clusterID)
	d.Set("raw_config", parsedKubeconfig.KubeconfigRaw)
	d.Set("server", parsedKubeconfig.Server)
	d.Set("cluster_ca_cert", parsedKubeconfig.ClusterCA)
	d.Set("client_cert", parsedKubeconfig.ClientCert)
	d.Set("client_key", parsedKubeconfig.ClientKey)
	return nil
}
