package selectel

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/cluster"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
)

func resourceMKSClusterV1() *schema.Resource {
	return &schema.Resource{
		Create:   resourceMKSClusterV1Create,
		Read:     resourceMKSClusterV1Read,
		Update:   resourceMKSClusterV1Update,
		Delete:   resourceMKSClusterV1Delete,
		Importer: nil,
		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf(
				"updated_at",
				func(d *schema.ResourceDiff, meta interface{}) bool {
					var hasChange bool
					if d.HasChange("maintenance_window_start") {
						hasChange = true
					}
					if d.HasChange("enable_autorepair") {
						hasChange = true
					}
					if d.HasChange("enable_patch_version_auto_upgrade") {
						hasChange = true
					}

					return hasChange
				}),
			customdiff.ComputedIf(
				"maintenance_window_end",
				func(d *schema.ResourceDiff, meta interface{}) bool {
					return d.HasChange("maintenance_window_start")
				}),
		),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kube_version": {
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
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
				}, false),
			},
			"nodegroups": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Set:      hashNodegroups,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Required: true,
						},
						"count": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"keypair_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"affinity_policy": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ram_mb": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"volume_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"local_volume": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"flavor_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"cluster_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
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
									"created_at": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"updated_at": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"hostname": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"nodegroup_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"enable_autorepair": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
			"enable_patch_version_auto_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"maintenance_window_start": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"maintenance_window_end": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_last_start": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pki_tree_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kube_api_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMKSClusterV1Create(d *schema.ResourceData, meta interface{}) error {
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
	endpoint := getClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	// Get all nodegroups from the schema and prepare cluster create options with the first nodegroup only.
	nodegroups := d.Get("nodegroups").(*schema.Set).List()
	enableAutorepair := d.Get("enable_autorepair").(bool)
	enablePatchVersionAutoUpgrade := d.Get("enable_patch_version_auto_upgrade").(bool)
	createOpts := &cluster.CreateOpts{
		Name:                          d.Get("name").(string),
		NetworkID:                     d.Get("network_id").(string),
		SubnetID:                      d.Get("subnet_id").(string),
		KubeVersion:                   d.Get("kube_version").(string),
		MaintenanceWindowStart:        d.Get("maintenance_window_start").(string),
		EnableAutorepair:              &enableAutorepair,
		EnablePatchVersionAutoUpgrade: &enablePatchVersionAutoUpgrade,
		Region:                        region,
		Nodegroups:                    []*nodegroup.CreateOpts{expandNodegroupCreateOpts(nodegroups[0])},
	}

	log.Print(msgCreate(objectCluster, createOpts))
	newCluster, _, err := cluster.Create(ctx, mksClient, createOpts)
	if err != nil {
		return errCreatingObject(objectCluster, err)
	}

	log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", newCluster.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForClusterV1ActiveState(ctx, mksClient, newCluster.ID, timeout)
	if err != nil {
		return errCreatingObject(objectCluster, err)
	}

	// Prepare a map to store nodegroup IDs that have already been saved in the nodegroup schema.
	nodegroupIDs := make(map[string]struct{})

	for i := 0; i < len(nodegroups); i++ {
		// If there is more than one nodegroup in the schema, create these nodegroups separately.
		if i > 0 {
			opts := expandNodegroupCreateOpts(nodegroups[i])

			log.Print(msgCreate(objectClusterNodegroups, opts))
			_, err := nodegroup.Create(ctx, mksClient, newCluster.ID, opts)
			if err != nil {
				return errCreatingObject(objectClusterNodegroups, err)
			}

			log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", newCluster.ID)
			timeout := d.Timeout(schema.TimeoutCreate)
			err = waitForClusterV1ActiveState(ctx, mksClient, newCluster.ID, timeout)
			if err != nil {
				return errCreatingObject(objectClusterNodegroups, err)
			}
		}

		// Get all nodegroups in the cluster, find new nodegroup ID and save it in the nodegroup schema.
		log.Print(msgGet(objectClusterNodegroups, newCluster.ID))
		allNodegroups, _, err := nodegroup.List(ctx, mksClient, newCluster.ID)
		if err != nil {
			return errGettingObject(objectClusterNodegroups, newCluster.ID, err)
		}

		for _, n := range allNodegroups {
			if _, ok := nodegroupIDs[n.ID]; !ok {
				nodegroups[i].(map[string]interface{})["id"] = n.ID
				nodegroupIDs[n.ID] = struct{}{}
			}
		}
	}

	if err := d.Set("nodegroups", nodegroups); err != nil {
		log.Print(errSettingComplexAttr("nodegroups", err))
	}
	d.SetId(newCluster.ID)

	return resourceMKSClusterV1Read(d, meta)
}

func resourceMKSClusterV1Read(d *schema.ResourceData, meta interface{}) error {
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
	endpoint := getClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	log.Print(msgGet(objectCluster, d.Id()))
	mksCluster, response, err := cluster.Get(ctx, mksClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return errGettingObject(objectCluster, d.Id(), err)
	}

	d.Set("name", mksCluster.Name)
	d.Set("status", mksCluster.Status)
	d.Set("project_id", mksCluster.ProjectID)
	d.Set("network_id", mksCluster.NetworkID)
	d.Set("subnet_id", mksCluster.SubnetID)
	d.Set("kube_api_ip", mksCluster.KubeAPIIP)
	d.Set("kube_version", mksCluster.KubeVersion)
	d.Set("region", mksCluster.Region)
	d.Set("maintenance_window_start", mksCluster.MaintenanceWindowStart)
	d.Set("maintenance_window_end", mksCluster.MaintenanceWindowEnd)
	d.Set("enable_autorepair", mksCluster.EnableAutorepair)
	d.Set("enable_patch_version_auto_upgrade", mksCluster.EnablePatchVersionAutoUpgrade)

	if mksCluster.CreatedAt != nil {
		d.Set("created_at", mksCluster.CreatedAt.String())
	}
	if mksCluster.UpdatedAt != nil {
		d.Set("updated_at", mksCluster.UpdatedAt.String())
	}
	if mksCluster.PKITreeUpdatedAt != nil {
		d.Set("pki_tree_updated_at", mksCluster.PKITreeUpdatedAt.String())
	}
	if mksCluster.MaintenanceLastStart != nil {
		d.Set("maintenance_last_start", mksCluster.MaintenanceLastStart.String())
	}

	log.Print(msgGet(objectClusterNodegroups, d.Id()))
	allNodegroups, _, err := nodegroup.List(ctx, mksClient, d.Id())
	if err != nil {
		return errGettingObject(objectClusterNodegroups, d.Id(), err)
	}

	nodegroups := flattenNodegroups(d, allNodegroups)
	if err := d.Set("nodegroups", nodegroups); err != nil {
		log.Print(errSettingComplexAttr("nodegroups", err))
	}

	return nil
}

func resourceMKSClusterV1Update(d *schema.ResourceData, meta interface{}) error {
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
	endpoint := getClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	var updateOpts cluster.UpdateOpts
	if d.HasChange("maintenance_window_start") {
		updateOpts.MaintenanceWindowStart = d.Get("maintenance_window_start").(string)
	}
	if d.HasChange("enable_autorepair") {
		v := d.Get("enable_autorepair").(bool)
		updateOpts.EnableAutorepair = &v
	}
	if d.HasChange("enable_patch_version_auto_upgrade") {
		v := d.Get("enable_patch_version_auto_upgrade").(bool)
		updateOpts.EnablePatchVersionAutoUpgrade = &v
	}

	if updateOpts != (cluster.UpdateOpts{}) {
		log.Print(msgUpdate(objectCluster, d.Id(), updateOpts))
		_, _, err := cluster.Update(ctx, mksClient, d.Id(), &updateOpts)
		if err != nil {
			return errUpdatingObject(objectCluster, d.Id(), err)
		}

		log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", d.Id())
		timeout := d.Timeout(schema.TimeoutUpdate)
		err = waitForClusterV1ActiveState(ctx, mksClient, d.Id(), timeout)
		if err != nil {
			return errUpdatingObject(objectCluster, d.Id(), err)
		}
	}

	if d.HasChange("nodegroups") {
		o, n := d.GetChange("nodegroups")
		oldNGSet := o.(*schema.Set)
		newNGSet := n.(*schema.Set)
		nodegroupsToAdd := newNGSet.Difference(oldNGSet)
		nodegroupsToRemove := oldNGSet.Difference(newNGSet)

		log.Printf("[DEBUG] Nodegroups to add: %v", nodegroupsToAdd)
		log.Printf("[DEBUG] Nodegroups to remove: %v", nodegroupsToRemove)

		if nodegroupsToAdd.Len() != 0 {
			// Prepare a map to store nodegroup IDs and populate it with values from the old nodegroup schema.
			nodegroupIDs := make(map[string]struct{})
			for _, v := range oldNGSet.List() {
				m := v.(map[string]interface{})
				if id, ok := m["id"]; ok {
					nodegroupIDs[id.(string)] = struct{}{}
				}
			}

			// Create each new nodegroup separately and save its ID in the new nodegroup schema.
			for _, newNG := range newNGSet.List() {
				for _, addNG := range nodegroupsToAdd.List() {
					if addNG.(map[string]interface{})["name"] == newNG.(map[string]interface{})["name"] {
						opts := expandNodegroupCreateOpts(addNG)

						log.Print(msgCreate(objectClusterNodegroups, opts))
						_, err := nodegroup.Create(ctx, mksClient, d.Id(), opts)
						if err != nil {
							return errCreatingObject(objectClusterNodegroups, err)
						}

						log.Printf("[DEBUG] waiting for cluster %s to become 'ACTIVE'", d.Id())
						timeout := d.Timeout(schema.TimeoutUpdate)
						err = waitForClusterV1ActiveState(ctx, mksClient, d.Id(), timeout)
						if err != nil {
							return errCreatingObject(objectClusterNodegroups, err)
						}

						log.Print(msgGet(objectClusterNodegroups, d.Id()))
						allNodegroups, _, err := nodegroup.List(ctx, mksClient, d.Id())
						if err != nil {
							return errGettingObject(objectClusterNodegroups, d.Id(), err)
						}

						for _, n := range allNodegroups {
							if _, ok := nodegroupIDs[n.ID]; !ok {
								newNG.(map[string]interface{})["id"] = n.ID
								nodegroupIDs[n.ID] = struct{}{}
							}
						}
					}
				}
			}
			d.Set("nodegroups", newNGSet)
		}

		for _, ng := range nodegroupsToRemove.List() {
			if nodegroupID, ok := ng.(map[string]interface{})["id"]; ok {
				log.Print(msgDelete(objectClusterNodegroups, d.Id()))
				_, err := nodegroup.Delete(ctx, mksClient, d.Id(), nodegroupID.(string))
				if err != nil {
					return errDeletingObject(objectClusterNodegroups, d.Id(), err)
				}

				timeout := d.Timeout(schema.TimeoutDelete)
				err = waitForClusterV1ActiveState(ctx, mksClient, d.Id(), timeout)
				if err != nil {
					return errDeletingObject(objectClusterNodegroups, d.Id(), err)
				}
			}
		}
	}

	return resourceMKSClusterV1Read(d, meta)
}

func resourceMKSClusterV1Delete(d *schema.ResourceData, meta interface{}) error {
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
	endpoint := getClusterV1Endpoint(region)
	mksClient := v1.NewMKSClientV1(token.ID, endpoint)

	log.Print(msgDelete(objectCluster, d.Id()))
	_, err = cluster.Delete(ctx, mksClient, d.Id())
	if err != nil {
		return errDeletingObject(objectCluster, d.Id(), err)
	}

	return nil
}
