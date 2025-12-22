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

func resourceGlobalRouterRouterV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalRouterRouterV1Create,
		ReadContext:   resourceGlobalRouterRouterV1Read,
		UpdateContext: resourceGlobalRouterRouterV1Update,
		DeleteContext: resourceGlobalRouterRouterV1Delete,
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
				Description: "Human-readable name of the router",
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
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Field for internal usage. If set to false, all router's networks are disabled",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource owner account UUID",
			},
			"project_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of a project in cloud",
			},
			"netops_router_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
			"leak_uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID for linked routers",
			},
			"prefix_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Prefix pool UUID",
			},
			"vpn_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Field for internal usage",
			},
		},
	}
}

func resourceGlobalRouterRouterV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var createOpts globalrouter.RouterCreateRequest

	name := d.Get("name").(string)
	createOpts.Name = name

	if v, ok := d.GetOk("tags"); ok {
		tags := v.(*schema.Set).List()
		createOpts.Tags = expandToStringSlice(tags)
	}

	createdRouter, _, err := client.RouterCreate(ctx, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGlobalRouterRouter, err))
	}

	d.SetId(createdRouter.ID)

	diagErr = waiters.WaitForRouterV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutCreate))
	if diagErr != nil {
		return diagErr
	}

	return resourceGlobalRouterRouterV1Read(ctx, d, meta)
}

func resourceGlobalRouterRouterV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	res, _, err := client.Router(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterRouter, d.Id(), err))
	}

	if res == nil {
		return diag.FromErr(fmt.Errorf("can't find router %q", d.Id()))
	}

	d.Set("name", res.Name)
	d.Set("tags", res.Tags)
	d.Set("created_at", res.CreatedAt)
	d.Set("updated_at", res.UpdatedAt)
	d.Set("enabled", res.Enabled)
	d.Set("status", res.Status)
	d.Set("account_id", res.AccountID)
	d.Set("project_id", res.ProjectID)
	d.Set("leak_uuid", res.LeakUUID)
	d.Set("netops_router_id", res.NetopsRouterID)
	d.Set("prefix_pool_id", res.PrefixPoolID)
	d.Set("vpn_id", res.VpnID)

	return nil
}

func resourceGlobalRouterRouterV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		updateOpts globalrouter.RouterUpdateRequest
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
		_, _, err := client.RouterUpdate(ctx, d.Id(), &updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectGlobalRouterRouter, d.Id(), err))
		}

		diagErr = waiters.WaitForRouterV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if diagErr != nil {
			return diagErr
		}

		return resourceGlobalRouterRouterV1Read(ctx, d, meta)
	}

	return nil
}

func resourceGlobalRouterRouterV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	resourceID := d.Id()

	_, err := client.RouterDelete(ctx, resourceID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGlobalRouterRouter, d.Id(), err))
	}

	diagErr = waiters.WaitForRouterV1Deleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
	if diagErr != nil {
		return diagErr
	}

	d.SetId("")

	return nil
}
