package selectel

import (
	"context"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/pointer"
)

const (
	vpcNetworkV2Object = "VPC network"
)

var vpcNetworkV2DNSDomainRegexp = regexp.MustCompile(`^$|\.$`)

func resourceVPCNetworkV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkV2Create,
		ReadContext:   resourceVPCNetworkV2Read,
		UpdateContext: resourceVPCNetworkV2Update,
		DeleteContext: resourceVPCNetworkV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkV2ImportState,
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
			"dns_domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					vpcNetworkV2DNSDomainRegexp,
					"fully-qualified (unambiguous) DNS domain names must have a dot at the end",
				),
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
			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"external": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_dns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceVPCNetworkV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := network.CreateRequest{
		AdminStateUp: pointer.To(d.Get("admin_state_up").(bool)),
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = pointer.To(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = pointer.To(v.(string))
	}
	if v, ok := d.GetOk("dns_domain"); ok {
		request.DNSDomain = pointer.To(v.(string))
	}
	if v, ok := d.GetOk("availability_zone_hints"); ok {
		hints := expandToStringSlice(v.([]interface{}))
		request.AvailabilityZoneHints = &hints
	}

	log.Print("[DEBUG] creating VPC network")
	created, err := network.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(vpcNetworkV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := network.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, created.ID, err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return nil
}

func resourceVPCNetworkV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     network.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = pointer.To(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = pointer.To(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = pointer.To(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("dns_domain") {
		request.DNSDomain = pointer.To(d.Get("dns_domain").(string))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating VPC network %s", d.Id())
		if _, err := network.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := network.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(vpcNetworkV2Object, d.Id(), err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting VPC network %s", d.Id())
	if err := network.Delete(ctx, client, d.Id()); err != nil &&
		!vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		return diag.FromErr(errDeletingObject(vpcNetworkV2Object, d.Id(), err))
	}

	return nil
}

func resourceVPCNetworkV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceVPCImportProjectRegion(d, meta); err != nil {
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
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}
