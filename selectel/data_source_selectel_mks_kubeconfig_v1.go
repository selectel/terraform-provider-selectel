package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			},
			"raw_config": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"server": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"cluster_ca_cert": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_cert": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceMKSKubeconfigV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mksClient, diagErr := getMKSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

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
