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

func resourceGlobalRouterStaticRouteV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalRouterStaticRouteV1Create,
		ReadContext:   resourceGlobalRouterStaticRouteV1Read,
		UpdateContext: resourceGlobalRouterStaticRouteV1Update,
		DeleteContext: resourceGlobalRouterStaticRouteV1Delete,
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
				Description: "Human-readable name of the static route",
			},
			"router_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the parent resource",
				ForceNew:    true,
			},
			"next_hop": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Next hop address in one of subnets connected to the parent router",
				ForceNew:    true,
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target subnet CIDR",
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "List of the resource tags",
			},
			"project_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of a project in cloud (taken from the parent subnet)",
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
			"netops_static_route_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Subnet UUID where next hop address is placed",
			},
		},
	}
}

func resourceGlobalRouterStaticRouteV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var createOpts globalrouter.StaticRouteCreateRequest

	routerID := d.Get("router_id").(string)
	createOpts.RouterID = routerID
	nextHop := d.Get("next_hop").(string)
	createOpts.NextHop = nextHop
	cidr := d.Get("cidr").(string)
	createOpts.Cidr = cidr
	name := d.Get("name").(string)
	createOpts.Name = name

	// optional args
	if v, ok := d.GetOk("tags"); ok {
		tags := v.(*schema.Set).List()
		createOpts.Tags = expandToStringSlice(tags)
	}

	createdRouter, _, err := client.StaticRouteCreate(ctx, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGlobalRouterStaticRoute, err))
	}

	d.SetId(createdRouter.ID)

	diagErr = waiters.WaitForStaticRouteV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutCreate))
	if diagErr != nil {
		return diagErr
	}

	return resourceGlobalRouterStaticRouteV1Read(ctx, d, meta)
}

func resourceGlobalRouterStaticRouteV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	res, _, err := client.StaticRoute(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterStaticRoute, d.Id(), err))
	}

	if res == nil {
		return diag.FromErr(fmt.Errorf("can't find router %q", d.Id()))
	}

	d.Set("name", res.Name)
	d.Set("router_id", res.RouterID)
	d.Set("next_hop", res.NextHop)
	d.Set("cidr", res.Cidr)
	d.Set("tags", res.Tags)
	d.Set("project_id", res.ProjectID)
	d.Set("status", res.Status)
	d.Set("created_at", res.CreatedAt)
	d.Set("updated_at", res.UpdatedAt)
	d.Set("account_id", res.AccountID)
	d.Set("netops_static_route_id", res.NetopsStaticRouteID)
	d.Set("subnet_id", res.SubnetID)

	return nil
}

func resourceGlobalRouterStaticRouteV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		updateOpts globalrouter.StaticRouteUpdateRequest
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
		_, _, err := client.StaticRouteUpdate(ctx, d.Id(), &updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectGlobalRouterStaticRoute, d.Id(), err))
		}

		diagErr = waiters.WaitForStaticRouteV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if diagErr != nil {
			return diagErr
		}

		return resourceGlobalRouterStaticRouteV1Read(ctx, d, meta)
	}

	return nil
}

func resourceGlobalRouterStaticRouteV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	resourceID := d.Id()

	_, err := client.StaticRouteDelete(ctx, resourceID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGlobalRouterStaticRoute, d.Id(), err))
	}

	diagErr = waiters.WaitForStaticRouteV1Deleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
	if diagErr != nil {
		return diagErr
	}

	d.SetId("")

	return nil
}
