package selectel

import (
	"context"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadiff"
)

const (
	vpcNetworkV2Object = "VPC network"
)

var vpcNetworkV2DNSDomainRegexp = regexp.MustCompile(`^$|\.$`)

var vpcNetworkDocs = resourceDocs{Name: "network"}

func resourceVPCNetworkV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkV2Create,
		ReadContext:   resourceVPCNetworkV2Read,
		UpdateContext: resourceVPCNetworkV2Update,
		DeleteContext: resourceVPCNetworkV2Delete,
		CustomizeDiff: schemadiff.RecomputeOnChange("all_tags", "tags"),
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(vpcV2DefaultDeleteTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkV2ImportState,
		},
		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return vpcNetworkDocs.identitySchemaByID(exampleResourceID, networkIDFromControlPanel, networkRegionFromControlPanel)
			},
		},
		Description: "Creates and manages a private network using public API v2. " +
			privateNetworksMoreInfo + "\n\n" +
			upgradeGuideNote,
		Schema: vpcNetworkDocs.withDocsHints(map[string]*schema.Schema{
			"id":         vpcNetworkDocs.idResourceSchema(),
			"project_id": vpcNetworkDocs.projectIDResourceSchema(),
			"region":     vpcNetworkDocs.regionResourceSchema(),
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the network.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the network.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Specifies if the network is administratively up.",
			},
			"dns_domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					vpcNetworkV2DNSDomainRegexp,
					"fully-qualified (unambiguous) DNS domain names must have a dot at the end",
				),
				Description: "DNS domain of the network. " + privateDNSLearnMore,
			},
			"availability_zone_hints": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments requested for allocating the network. " + regionLearnMore,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the network.",
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the network.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the network.",
			},
			"subnets": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of the subnets within the network.",
			},
			"availability_zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Pool segments where the network is actually allocated.",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network can be accessed by any project.",
			},
			"external": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network has an external routing facility.",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum transmission unit of the network.",
			},
			"port_security_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if port security is enabled by default for the ports of the network.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network is administratively blocked.",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the network is public.",
			},
			"is_dns_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the private DNS service is enabled for the network.",
			},
		}),
	}
}

func resourceVPCNetworkV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := network.CreateRequest{
		AdminStateUp: new(d.Get("admin_state_up").(bool)),
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = new(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = new(v.(string))
	}
	if v, ok := d.GetOk("dns_domain"); ok {
		request.DNSDomain = new(v.(string))
	}
	if v, ok := d.GetOk("availability_zone_hints"); ok {
		hints := convert.ToStringSlice(v.([]any))
		request.AvailabilityZoneHints = &hints
	}

	log.Print("[DEBUG] creating VPC network")
	created, err := network.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcNetworkV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := convert.ToStringSlice(tags.(*schema.Set).List())
		if _, err := network.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, created.ID, err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting VPC network %s", d.Id())
	found, err := network.Get(ctx, client, d.Id())
	if err != nil {
		return vpcCheckDeleted(d, err, vpcNetworkV2Object)
	}

	flattenVPCNetworkV2(d, found)
	setProjectRegionIdentity(d)

	return nil
}

func resourceVPCNetworkV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     network.UpdateRequest
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
	if d.HasChange("dns_domain") {
		request.DNSDomain = new(d.Get("dns_domain").(string))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating VPC network %s", d.Id())
		if _, err := network.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandVPCUpdateTagsV2(d)
		if _, err := network.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, d.Id(), err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC network %s", d.Id())
	err = deleteVPCV2WithRetry(ctx, vpcNetworkV2Object, d.Id(),
		vpcV2StatusPollInterval, d.Timeout(schema.TimeoutDelete), vpcV2ConflictIsTransient,
		func(ctx context.Context) error { return network.Delete(ctx, client, d.Id()) })
	if err != nil {
		return diag.FromErr(errDeletingObject(vpcNetworkV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCNetworkV2ImportState(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func flattenVPCNetworkV2(d *schema.ResourceData, found *network.Network) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("admin_state_up", found.AdminStateUp)
	d.Set("dns_domain", found.DNSDomain)
	d.Set("status", found.Status)
	d.Set("shared", found.Shared)
	d.Set("external", found.RouterExternal)
	d.Set("mtu", found.MTU)
	d.Set("port_security_enabled", found.PortSecurityEnabled)
	d.Set("blocked", found.Blocked)
	d.Set("is_public", found.IsPublic)
	d.Set("is_dns_enabled", found.IsDNSEnabled)

	if err := d.Set("availability_zone_hints", found.AvailabilityZoneHints); err != nil {
		log.Print(errSettingComplexAttr("availability_zone_hints", err))
	}
	if err := d.Set("availability_zones", found.AvailabilityZones); err != nil {
		log.Print(errSettingComplexAttr("availability_zones", err))
	}
	if err := d.Set("subnets", found.Subnets); err != nil {
		log.Print(errSettingComplexAttr("subnets", err))
	}
	flattenVPCTagsV2(d, found.Tags)
}
