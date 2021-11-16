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

func dataSourceMksClusterKubeconfigV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMksClusterKubeconfigV1Read,
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
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"kubeconfig": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMksClusterKubeconfigV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	clusterId := d.Get("cluster_id").(string)

	mksCluster, _, err := cluster.Get(ctx, mksClient, clusterId)
	if err != nil {
		return diag.FromErr(errGettingObject(objectCluster, clusterId, err))
	}

	kubeconfig, _, err := cluster.GetKubeconfig(ctx, mksClient, mksCluster.ID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectKubeConfig, clusterId, err))
	}

	d.SetId(clusterId)
	d.Set("kubeconfig", string(kubeconfig))
	return nil
}
