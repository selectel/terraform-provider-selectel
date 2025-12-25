package selectel

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/globalrouter"
)

func resourceGlobalRouterDedicatedNetworkV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalRouterDedicatedNetworkV1Create,
		ReadContext:   resourceGlobalRouterDedicatedNetworkV1Read,
		UpdateContext: resourceGlobalRouterDedicatedNetworkV1Update,
		DeleteContext: resourceGlobalRouterDedicatedNetworkV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Human-readable name of the network",
			},
			"router_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the parent router",
				ForceNew:    true,
			},
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the zone",
				ForceNew:    true,
			},
			"vlan": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "VLAN of the network in the dedicated networks",
				ForceNew:    true,
			},
			"inner_vlan": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Inner VLAN of the network in the dedicated networks",
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "List of the resource tags",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource creation time",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource last update time",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource status",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource owner account UUID",
			},
			"netops_vlan_uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
			"sv_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
		},
	}
}

func resourceGlobalRouterDedicatedNetworkV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var createOpts globalrouter.DedicatedNetworkCreateRequest

	routerID := d.Get("router_id").(string)
	createOpts.RouterID = routerID
	zoneID := d.Get("zone_id").(string)
	createOpts.ZoneID = zoneID
	vlan := d.Get("vlan").(int)
	createOpts.Vlan = vlan
	name := d.Get("name").(string)
	createOpts.Name = name

	// optional args
	if v, ok := d.GetOk("inner_vlan"); ok {
		innerVlan := v.(int)
		createOpts.InnerVlan = innerVlan
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.(*schema.Set).List()
		createOpts.Tags = expandToStringSlice(tags)
	}

	createdRouter, _, err := client.DedicatedNetworkCreate(ctx, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGlobalRouterDedicatedNetwork, err))
	}

	d.SetId(createdRouter.ID)

	diagErr = waiters.WaitForNetworkV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutCreate))
	if diagErr != nil {
		return diagErr
	}

	return resourceGlobalRouterDedicatedNetworkV1Read(ctx, d, meta)
}

func resourceGlobalRouterDedicatedNetworkV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	res, _, err := client.Network(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterDedicatedNetwork, d.Id(), err))
	}

	if res == nil {
		return diag.FromErr(fmt.Errorf("can't find network %q", d.Id()))
	}

	d.Set("name", res.Name)
	d.Set("router_id", res.RouterID)
	d.Set("zone_id", res.ZoneID)
	d.Set("vlan", res.Vlan)
	d.Set("inner_vlan", res.InnerVlan)
	d.Set("tags", res.Tags)
	d.Set("created_at", res.CreatedAt)
	d.Set("updated_at", res.UpdatedAt)
	d.Set("status", res.Status)
	d.Set("account_id", res.AccountID)
	d.Set("netops_vlan_uuid", res.NetopsVlanUUID)
	d.Set("sv_network_id", res.SvNetworkID)

	return nil
}

func resourceGlobalRouterDedicatedNetworkV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		updateOpts globalrouter.NetworkUpdateRequest
		hasChange  bool
	)

	if d.HasChange("name") {
		hasChange = true
		name := d.Get("name").(string)
		updateOpts.Name = &name
	}

	if d.HasChange("tags") {
		hasChange = true
		if v, ok := d.GetOk("tags"); ok {
			tags := v.(*schema.Set).List()
			tagsToUpdate := expandToStringSlice(tags)
			updateOpts.Tags = &tagsToUpdate
		} else {
			updateOpts.Tags = &[]string{}
		}
	}

	if hasChange {
		_, _, err := client.NetworkUpdate(ctx, d.Id(), &updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectGlobalRouterDedicatedNetwork, d.Id(), err))
		}

		diagErr = waiters.WaitForNetworkV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if diagErr != nil {
			return diagErr
		}

		return resourceGlobalRouterDedicatedNetworkV1Read(ctx, d, meta)
	}

	return nil
}

func resourceGlobalRouterDedicatedNetworkV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	resourceID := d.Id()

	_, err := client.NetworkDisconnect(ctx, resourceID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGlobalRouterDedicatedNetwork, d.Id(), err))
	}

	diagErr = waiters.WaitForNetworkV1Deleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
	if diagErr != nil {
		return diagErr
	}

	d.SetId("")

	return nil
}
