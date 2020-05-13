package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
)

func resourceMKSNodegroupV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceMKSNodegroupV1Create,
		Read:   resourceMKSNodegroupV1Read,
		Update: resourceMKSNodegroupV1Update,
		Delete: resourceMKSNodegroupV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceMKSNodegroupV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nodes_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"keypair_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"affinity_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cpus": {
				Type:          schema.TypeInt,
				ConflictsWith: []string{"flavor_id"},
				Optional:      true,
				ForceNew:      true,
			},
			"ram_mb": {
				Type:          schema.TypeInt,
				ConflictsWith: []string{"flavor_id"},
				Optional:      true,
				ForceNew:      true,
			},
			"volume_gb": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"volume_type": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"local_volume"},
				Optional:      true,
				ForceNew:      true,
			},
			"local_volume": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMKSNodegroupV1Create(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	selMutexKV.Lock(clusterID)
	defer selMutexKV.Unlock(clusterID)

	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return errCreatingObject(objectToken, err)
	}

	region := d.Get("region").(string)
	endpoint := getMKSClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	// Get a list of all nodegroups in the cluster.
	allNodegroups, _, err := nodegroup.List(ctx, mksClient, clusterID)
	if err != nil {
		return errGettingObject("all nodegroups in the cluster", clusterID, err)
	}

	// Prepare a map with known nodegroup IDs.
	nodegroupIDs := make(map[string]struct{})
	for _, ng := range allNodegroups {
		if _, ok := nodegroupIDs[ng.ID]; !ok {
			nodegroupIDs[ng.ID] = struct{}{}
		}
	}

	// Prepare nodegroup create options.
	createOpts := &nodegroup.CreateOpts{
		Count:            d.Get("nodes_count").(int),
		FlavorID:         d.Get("flavor_id").(string),
		CPUs:             d.Get("cpus").(int),
		RAMMB:            d.Get("ram_mb").(int),
		VolumeGB:         d.Get("volume_gb").(int),
		VolumeType:       d.Get("volume_type").(string),
		LocalVolume:      d.Get("local_volume").(bool),
		KeypairName:      d.Get("keypair_name").(string),
		AffinityPolicy:   d.Get("affinity_policy").(string),
		AvailabilityZone: d.Get("availability_zone").(string),
	}

	labels := d.Get("labels").(map[string]interface{})
	createOpts.Labels = expandMKSNodegroupV1Labels(labels)

	log.Print(msgCreate(objectNodegroup, createOpts))
	_, err = nodegroup.Create(ctx, mksClient, clusterID, createOpts)
	if err != nil {
		return errCreatingObject(objectNodegroup, err)
	}

	log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", clusterID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForMKSClusterV1ActiveState(ctx, mksClient, clusterID, timeout)
	if err != nil {
		return errCreatingObject(objectNodegroup, err)
	}

	// Get a list of all nodegroups in the cluster and find a new nodegroup.
	allNodegroups, _, err = nodegroup.List(ctx, mksClient, clusterID)
	if err != nil {
		return errGettingObject("all nodegroups in the cluster", clusterID, err)
	}

	var nodegroupID string
	for _, ng := range allNodegroups {
		if _, ok := nodegroupIDs[ng.ID]; !ok {
			nodegroupID = ng.ID
		}
	}

	// The ID must be a combination of the cluster and nodegroup ID
	// since a cluster ID is required to retrieve a nodegroup ID.
	id := fmt.Sprintf("%s/%s", clusterID, nodegroupID)
	d.SetId(id)

	return resourceMKSNodegroupV1Read(d, meta)
}

func resourceMKSNodegroupV1Read(d *schema.ResourceData, meta interface{}) error {
	clusterID, nodegroupID, err := mksNodegroupV1ParseID(d.Id())
	if err != nil {
		d.SetId("")
		return errGettingObject(objectNodegroup, d.Id(), err)
	}

	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return errCreatingObject(objectToken, err)
	}

	region := d.Get("region").(string)
	endpoint := getMKSClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	log.Print(msgGet(objectNodegroup, d.Id()))
	mksNodegroup, response, err := nodegroup.Get(ctx, mksClient, clusterID, nodegroupID)
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return errGettingObject(objectNodegroup, d.Id(), err)
	}

	d.Set("cluster_id", mksNodegroup.ClusterID)
	d.Set("flavor_id", mksNodegroup.FlavorID)
	d.Set("volume_gb", mksNodegroup.VolumeGB)
	d.Set("volume_type", mksNodegroup.VolumeType)
	d.Set("local_volume", mksNodegroup.LocalVolume)
	d.Set("availability_zone", mksNodegroup.AvailabilityZone)

	if err := d.Set("labels", mksNodegroup.Labels); err != nil {
		log.Print(errSettingComplexAttr("labels", err))
	}

	nodes := flattenMKSNodegroupV1Nodes(mksNodegroup.Nodes)
	if err := d.Set("nodes", nodes); err != nil {
		log.Print(errSettingComplexAttr("nodes", err))
	}

	return nil
}

func resourceMKSNodegroupV1Update(d *schema.ResourceData, meta interface{}) error {
	clusterID, nodegroupID, err := mksNodegroupV1ParseID(d.Id())
	if err != nil {
		d.SetId("")
		return errUpdatingObject(objectNodegroup, d.Id(), err)
	}

	selMutexKV.Lock(clusterID)
	defer selMutexKV.Unlock(clusterID)

	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return errCreatingObject(objectToken, err)
	}

	region := d.Get("region").(string)
	endpoint := getMKSClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	if d.HasChange("labels") {
		labels := d.Get("labels").(map[string]interface{})
		updateOpts := nodegroup.UpdateOpts{
			Labels: expandMKSNodegroupV1Labels(labels),
		}

		log.Print(msgUpdate(objectNodegroup, d.Id(), updateOpts))
		_, err := nodegroup.Update(ctx, mksClient, clusterID, nodegroupID, &updateOpts)
		if err != nil {
			return errUpdatingObject(objectNodegroup, d.Id(), err)
		}

		log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", clusterID)
		timeout := d.Timeout(schema.TimeoutUpdate)
		err = waitForMKSClusterV1ActiveState(ctx, mksClient, clusterID, timeout)
		if err != nil {
			return errUpdatingObject(objectNodegroup, d.Id(), err)
		}
	}

	if d.HasChange("nodes_count") {
		resizeOpts := nodegroup.ResizeOpts{
			Desired: d.Get("nodes_count").(int),
		}

		log.Print(msgUpdate(objectNodegroup, d.Id(), resizeOpts))
		_, err = nodegroup.Resize(ctx, mksClient, clusterID, nodegroupID, &resizeOpts)
		if err != nil {
			return errUpdatingObject(objectNodegroup, d.Id(), err)
		}

		log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", clusterID)
		timeout := d.Timeout(schema.TimeoutUpdate)
		err = waitForMKSClusterV1ActiveState(ctx, mksClient, clusterID, timeout)
		if err != nil {
			return errUpdatingObject(objectNodegroup, d.Id(), err)
		}
	}

	return resourceMKSNodegroupV1Read(d, meta)
}

func resourceMKSNodegroupV1Delete(d *schema.ResourceData, meta interface{}) error {
	clusterID, nodegroupID, err := mksNodegroupV1ParseID(d.Id())
	if err != nil {
		d.SetId("")
		return errDeletingObject(objectNodegroup, d.Id(), err)
	}

	selMutexKV.Lock(clusterID)
	defer selMutexKV.Unlock(clusterID)

	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return errCreatingObject(objectToken, err)
	}

	region := d.Get("region").(string)
	endpoint := getMKSClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	log.Print(msgDelete(objectNodegroup, d.Id()))
	_, err = nodegroup.Delete(ctx, mksClient, clusterID, nodegroupID)
	if err != nil {
		return errDeletingObject(objectNodegroup, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", clusterID)
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waitForMKSClusterV1ActiveState(ctx, mksClient, clusterID, timeout)
	if err != nil {
		return errDeletingObject(objectNodegroup, d.Id(), err)
	}

	return nil
}

func resourceMKSNodegroupV1ImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("SEL_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return nil, errors.New("SEL_REGION must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	d.Set("region", config.Region)

	return []*schema.ResourceData{d}, nil
}
