package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadiff"
)

const vpcRouterV2Object = "VPC router"

func resourceVPCRouterV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRouterV2Create,
		ReadContext:   resourceVPCRouterV2Read,
		UpdateContext: resourceVPCRouterV2Update,
		DeleteContext: resourceVPCRouterV2Delete,
		CustomizeDiff: schemadiff.RecomputeOnChange("all_tags", "tags"),
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(vpcV2DefaultDeleteTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCRouterV2ImportState,
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ha": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"external_network_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_snat": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"external_network_id"},
			},
			"availability_zone_hints": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"external_fixed_ip": vpcComputedFixedIPsSchema(),
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceVPCRouterV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := router.CreateRequest{
		AdminStateUp: new(d.Get("admin_state_up").(bool)),
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = new(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if schemaconfig.IsSet(d, "ha") {
		request.HA = new(d.Get("ha").(bool))
	}
	if v, ok := d.GetOk("availability_zone_hints"); ok {
		hints := convert.ToStringSlice(v.([]any))
		request.AvailabilityZoneHints = &hints
	}
	if gateway, ok := expandVPCRouterV2Gateway(d); ok {
		request.ExternalGateway = vpc.Value(gateway)
	}

	log.Print("[DEBUG] creating VPC router")
	created, err := router.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcRouterV2Object, err))
	}

	d.SetId(created.ID)

	if err := checkVPCRouterV2Created(created); err != nil {
		return diag.FromErr(err)
	}

	if tags, ok := d.GetOk("tags"); ok {
		values := convert.ToStringSlice(tags.(*schema.Set).List())
		if _, err := router.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcRouterV2Object, created.ID, err))
		}
	}

	return resourceVPCRouterV2Read(ctx, d, meta)
}

func resourceVPCRouterV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC router %s", d.Id())
	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcRouterV2Object)
	}

	flattenVPCRouterV2(d, found)

	return nil
}

func resourceVPCRouterV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     router.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = new(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = new(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = new(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("external_network_id") || d.HasChange("enable_snat") {
		if gateway, ok := expandVPCRouterV2Gateway(d); ok {
			request.ExternalGateway = vpc.Value(gateway)
		} else {
			request.ExternalGateway = vpc.Null[router.ExternalGatewayRequest]()
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating VPC router %s", d.Id())
		if _, err := router.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcRouterV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandVPCUpdateTagsV2(d)
		if _, err := router.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcRouterV2Object, d.Id(), err))
		}
	}

	if d.HasChange("ha") {
		selMutexKV.Lock(d.Id())
		defer selMutexKV.Unlock(d.Id())

		err := migrateVPCRouterV2HA(ctx, client, d.Id(), d.Get("ha").(bool), d.Get("admin_state_up").(bool))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceVPCRouterV2Read(ctx, d, meta)
}

func migrateVPCRouterV2HA(ctx context.Context, client *vpc.Client, id string, ha, adminStateUp bool) error {
	if adminStateUp {
		log.Printf("[DEBUG] disabling VPC router %s before changing ha", id)
		if _, err := router.Update(ctx, client, id, router.UpdateRequest{AdminStateUp: new(false)}); err != nil {
			return fmt.Errorf("error disabling VPC router %s before changing ha: %w", id, err)
		}
	}

	log.Printf("[DEBUG] setting ha=%t on VPC router %s", ha, id)
	_, migrateErr := router.Update(ctx, client, id, router.UpdateRequest{HA: new(ha)})
	if migrateErr != nil {
		migrateErr = fmt.Errorf("error migrating VPC router %s to ha=%t: %w", id, ha, migrateErr)
	}

	if adminStateUp {
		log.Printf("[DEBUG] re-enabling VPC router %s after changing ha", id)
		if _, err := router.Update(ctx, client, id, router.UpdateRequest{AdminStateUp: new(true)}); err != nil {
			return errors.Join(migrateErr, fmt.Errorf("error re-enabling VPC router %s after changing ha: %w", id, err))
		}
	}

	return migrateErr
}

func resourceVPCRouterV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC router %s", d.Id())
	err = deleteVPCV2WithRetry(ctx, vpcRouterV2Object, d.Id(),
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutDelete), vpcV2ConflictIsTransient,
		func(ctx context.Context) error { return router.Delete(ctx, client, d.Id()) })
	if err != nil {
		return diag.FromErr(errDeletingObject(vpcRouterV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCRouterV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func expandVPCRouterV2Gateway(d *schema.ResourceData) (router.ExternalGatewayRequest, bool) {
	networkID, ok := d.GetOk("external_network_id")
	if !ok {
		return router.ExternalGatewayRequest{}, false
	}

	gateway := router.ExternalGatewayRequest{NetworkID: networkID.(string)}
	if schemaconfig.IsSet(d, "enable_snat") {
		gateway.EnableSNAT = new(d.Get("enable_snat").(bool))
	}

	return gateway, true
}

func flattenVPCRouterV2(d *schema.ResourceData, found *router.Router) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("admin_state_up", found.AdminStateUp)
	d.Set("ha", found.HA)
	d.Set("status", found.Status)
	d.Set("blocked", found.Blocked)

	if found.ExternalGateway != nil {
		d.Set("external_network_id", found.ExternalGateway.NetworkID)
		d.Set("enable_snat", found.ExternalGateway.EnableSNAT)
		if err := d.Set("external_fixed_ip", flattenVPCRouterV2FixedIPs(found.ExternalGateway.ExternalFixedIPs)); err != nil {
			log.Print(errSettingComplexAttr("external_fixed_ip", err))
		}
	} else {
		d.Set("external_network_id", "")
		d.Set("external_fixed_ip", nil)
	}

	if err := d.Set("availability_zone_hints", found.AvailabilityZoneHints); err != nil {
		log.Print(errSettingComplexAttr("availability_zone_hints", err))
	}
	if err := d.Set("availability_zones", found.AvailabilityZones); err != nil {
		log.Print(errSettingComplexAttr("availability_zones", err))
	}
	flattenVPCTagsV2(d, found.Tags)
}

func flattenVPCRouterV2FixedIPs(fixedIPs []router.ExternalFixedIP) []any {
	result := make([]any, 0, len(fixedIPs))
	for _, fixedIP := range fixedIPs {
		result = append(result, map[string]any{
			"subnet_id":  fixedIP.SubnetID,
			"ip_address": fixedIP.IPAddress,
		})
	}

	return result
}

func checkVPCRouterV2Created(created *router.Router) error {
	if created.Status == vpcV2StatusError {
		return fmt.Errorf("%s %s is in %s status", vpcRouterV2Object, created.ID, created.Status)
	}

	return nil
}
