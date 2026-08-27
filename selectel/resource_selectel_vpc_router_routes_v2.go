package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/sliceutil"
)

const vpcRouterRoutesV2Object = "VPC router routes"

func resourceVPCRouterRoutesV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRouterRoutesV2Create,
		ReadContext:   resourceVPCRouterRoutesV2Read,
		UpdateContext: resourceVPCRouterRoutesV2Update,
		DeleteContext: resourceVPCRouterRoutesV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRouterRoutesV2ImportState,
		},
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
			"router_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"routes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"next_hop": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
					},
				},
			},
		},
	}
}

func resourceVPCRouterRoutesV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	routerID := d.Get("router_id").(string)
	desired := expandVPCRouterRoutesV2(d.Get("routes"))

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		return diag.FromErr(errGettingObject(vpcRouterV2Object, routerID, err))
	}

	// Avoid deleting pre-existing routes during initial creation.
	routes := sliceutil.Union(found.Routes, desired)

	if len(routes) != len(found.Routes) {
		log.Printf("[DEBUG] adding %d route(s) to VPC router %s", len(routes)-len(found.Routes), routerID)
		if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
			Routes: &routes,
		}); err != nil {
			return diag.FromErr(errCreatingObject(vpcRouterRoutesV2Object, err))
		}
	}

	d.SetId(routerID)

	return resourceVPCRouterRoutesV2Read(ctx, d, meta)
}

func resourceVPCRouterRoutesV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting routes of VPC router %s", d.Id())
	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcRouterV2Object)
	}

	d.Set("router_id", found.ID)
	if err := d.Set("routes", flattenVPCRouterRoutesV2(found.Routes)); err != nil {
		log.Print(errSettingComplexAttr("routes", err))
	}

	return nil
}

func resourceVPCRouterRoutesV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.HasChange("routes") {
		return resourceVPCRouterRoutesV2Read(ctx, d, meta)
	}

	routerID := d.Get("router_id").(string)
	desired := expandVPCRouterRoutesV2(d.Get("routes"))

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	log.Printf("[DEBUG] replacing routes of VPC router %s with %d entrie(s)", routerID, len(desired))
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &desired,
	}); err != nil {
		return diag.FromErr(errUpdatingObject(vpcRouterRoutesV2Object, routerID, err))
	}

	return resourceVPCRouterRoutesV2Read(ctx, d, meta)
}

func resourceVPCRouterRoutesV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	managed := expandVPCRouterRoutesV2(d.Get("routes"))
	if len(managed) == 0 {
		return nil
	}

	routerID := d.Get("router_id").(string)

	selMutexKV.Lock(routerID)
	defer selMutexKV.Unlock(routerID)

	found, err := router.Get(ctx, client, routerID)
	if err != nil {
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return nil
		}

		return diag.FromErr(errGettingObject(vpcRouterV2Object, routerID, err))
	}

	kept := sliceutil.Subtract(found.Routes, managed)

	if len(kept) == len(found.Routes) {
		return nil
	}

	log.Printf("[DEBUG] removing %d managed route(s) from VPC router %s", len(found.Routes)-len(kept), routerID)
	if _, err := router.Update(ctx, client, routerID, router.UpdateRequest{
		Routes: &kept,
	}); err != nil {
		return diag.FromErr(errDeletingObject(vpcRouterRoutesV2Object, routerID, err))
	}

	return nil
}

func resourceVPCRouterRoutesV2ImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(vpcRouterV2Object, d.Id(), err)
	}

	d.Set("router_id", found.ID)
	if err := d.Set("routes", flattenVPCRouterRoutesV2(found.Routes)); err != nil {
		log.Print(errSettingComplexAttr("routes", err))
	}

	return []*schema.ResourceData{d}, nil
}

func expandVPCRouterRoutesV2(value any) []router.Route {
	set, ok := value.(*schema.Set)
	if !ok {
		return []router.Route{}
	}

	routes := make([]router.Route, 0, set.Len())
	for _, raw := range set.List() {
		entry := raw.(map[string]any)
		routes = append(routes, router.Route{
			Destination: entry["destination_cidr"].(string),
			NextHop:     entry["next_hop"].(string),
		})
	}

	return routes
}

func flattenVPCRouterRoutesV2(routes []router.Route) []any {
	result := make([]any, 0, len(routes))
	for _, route := range routes {
		result = append(result, map[string]any{
			"destination_cidr": route.Destination,
			"next_hop":         route.NextHop,
		})
	}

	return result
}
