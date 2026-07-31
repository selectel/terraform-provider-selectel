package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
)

const networkingRouterV2Object = "networking router"

func resourceNetworkingRouterV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingRouterV2Create,
		ReadContext:   resourceNetworkingRouterV2Read,
		UpdateContext: resourceNetworkingRouterV2Update,
		DeleteContext: resourceNetworkingRouterV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingRouterV2ImportState,
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
				Computed: true,
			},
			// The external gateway is limited to the two sub-attributes the
			// cloud's policy grants to a project owner. The addresses of the
			// gateway are assigned by the API and reported in external_fixed_ip.
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
			"flavor_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
			// Read-only: the policy reserves writing external_fixed_ips for the
			// Keystone admin role.
			"external_fixed_ip": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingRouterV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := router.CreateRequest{}
	if v, ok := d.GetOk("name"); ok {
		request.Name = vpcString(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOkExists("admin_state_up"); ok { //nolint:staticcheck // GetOkExists distinguishes an explicit false from an unset bool
		request.AdminStateUp = vpcBool(v.(bool))
	}
	if v, ok := d.GetOk("flavor_id"); ok {
		request.FlavorID = vpcString(v.(string))
	}
	if v, ok := d.GetOk("availability_zone_hints"); ok {
		hints := expandToStringSlice(v.([]interface{}))
		request.AvailabilityZoneHints = &hints
	}
	if gateway, ok := expandNetworkingRouterV2Gateway(d); ok {
		request.ExternalGateway = vpc.Value(gateway)
	}

	log.Print("[DEBUG] creating networking router")
	created, err := router.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingRouterV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := router.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingRouterV2Object, created.ID, err))
		}
	}

	return resourceNetworkingRouterV2Read(ctx, d, meta)
}

func resourceNetworkingRouterV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking router %s", d.Id())
	found, err := router.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingRouterV2Object)
	}

	flattenNetworkingRouterV2(d, found)

	return nil
}

func resourceNetworkingRouterV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     router.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = vpcString(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = vpcBool(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("external_network_id") || d.HasChange("enable_snat") {
		// Detaching the gateway is an explicit JSON null; attaching or changing
		// it sends the whole gateway object, which is how Neutron models it.
		if gateway, ok := expandNetworkingRouterV2Gateway(d); ok {
			request.ExternalGateway = vpc.Value(gateway)
		} else {
			request.ExternalGateway = vpc.Null[router.ExternalGatewayRequest]()
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking router %s", d.Id())
		if _, err := router.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingRouterV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := router.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingRouterV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingRouterV2Read(ctx, d, meta)
}

func resourceNetworkingRouterV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking router %s", d.Id())
	if err := router.Delete(ctx, client, d.Id()); err != nil {
		// A router with attached interfaces is refused with a conflict, which
		// stays a diagnostic error: detaching them silently is not this
		// resource's business.
		return networkingCheckDeletedOnDelete(d, err, networkingRouterV2Object, func() error {
			_, confirmErr := router.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingRouterV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// expandNetworkingRouterV2Gateway builds the external gateway from the
// configuration. It reports false when no gateway is configured, which on
// update means the gateway is being detached.
func expandNetworkingRouterV2Gateway(d *schema.ResourceData) (router.ExternalGatewayRequest, bool) {
	networkID, ok := d.GetOk("external_network_id")
	if !ok {
		return router.ExternalGatewayRequest{}, false
	}

	gateway := router.ExternalGatewayRequest{NetworkID: networkID.(string)}
	if v, ok := d.GetOkExists("enable_snat"); ok { //nolint:staticcheck // GetOkExists distinguishes an explicit false from an unset bool
		gateway.EnableSNAT = vpcBool(v.(bool))
	}

	return gateway, true
}

// flattenNetworkingRouterV2 writes an API router into the state.
func flattenNetworkingRouterV2(d *schema.ResourceData, found *router.Router) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("admin_state_up", found.AdminStateUp)
	d.Set("status", found.Status)
	d.Set("flavor_id", derefString(found.FlavorID))
	d.Set("blocked", found.Blocked)

	if found.ExternalGateway != nil {
		d.Set("external_network_id", found.ExternalGateway.NetworkID)
		d.Set("enable_snat", found.ExternalGateway.EnableSNAT)
		if err := d.Set("external_fixed_ip", flattenNetworkingRouterV2FixedIPs(found.ExternalGateway.ExternalFixedIPs)); err != nil {
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
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}

func flattenNetworkingRouterV2FixedIPs(fixedIPs []router.ExternalFixedIP) []interface{} {
	result := make([]interface{}, 0, len(fixedIPs))
	for _, fixedIP := range fixedIPs {
		result = append(result, map[string]interface{}{
			"subnet_id":  fixedIP.SubnetID,
			"ip_address": fixedIP.IPAddress,
		})
	}

	return result
}
