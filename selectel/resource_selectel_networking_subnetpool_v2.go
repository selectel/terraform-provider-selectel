package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/subnetpool"
)

const networkingSubnetPoolV2Object = "networking subnetpool"

func resourceNetworkingSubnetPoolV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSubnetPoolV2Create,
		ReadContext:   resourceNetworkingSubnetPoolV2Read,
		UpdateContext: resourceNetworkingSubnetPoolV2Update,
		DeleteContext: resourceNetworkingSubnetPoolV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingSubnetPoolV2ImportState,
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
				Required: true,
			},
			"prefixes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_quota": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"default_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"min_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"max_prefixlen": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"address_scope_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			// ip_version is derived by Neutron from prefixes and is not accepted
			// on create, so it is read-only here.
			"ip_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// shared and is_default are writable only with admin privileges.
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
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
			"revision_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingSubnetPoolV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := subnetpool.CreateRequest{
		Name:     vpcString(d.Get("name").(string)),
		Prefixes: expandToStringSlice(d.Get("prefixes").([]interface{})),
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOk("default_quota"); ok {
		request.DefaultQuota = vpcInt(v.(int))
	}
	if v, ok := d.GetOk("default_prefixlen"); ok {
		request.DefaultPrefixLen = vpcInt(v.(int))
	}
	if v, ok := d.GetOk("min_prefixlen"); ok {
		request.MinPrefixLen = vpcInt(v.(int))
	}
	if v, ok := d.GetOk("max_prefixlen"); ok {
		request.MaxPrefixLen = vpcInt(v.(int))
	}
	if v, ok := d.GetOk("address_scope_id"); ok {
		request.AddressScopeID = vpcString(v.(string))
	}

	log.Print("[DEBUG] creating networking subnetpool")
	created, err := subnetpool.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingSubnetPoolV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := subnetpool.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetPoolV2Object, created.ID, err))
		}
	}

	return resourceNetworkingSubnetPoolV2Read(ctx, d, meta)
}

func resourceNetworkingSubnetPoolV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking subnetpool %s", d.Id())
	found, err := subnetpool.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingSubnetPoolV2Object)
	}

	flattenNetworkingSubnetPoolV2(d, found)

	return nil
}

func resourceNetworkingSubnetPoolV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     subnetpool.UpdateRequest
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
	if d.HasChange("prefixes") {
		prefixes := expandToStringSlice(d.Get("prefixes").([]interface{}))
		request.Prefixes = &prefixes
		needsUpdate = true
	}
	if d.HasChange("default_quota") {
		request.DefaultQuota = vpcInt(d.Get("default_quota").(int))
		needsUpdate = true
	}
	if d.HasChange("default_prefixlen") {
		request.DefaultPrefixLen = vpcInt(d.Get("default_prefixlen").(int))
		needsUpdate = true
	}
	if d.HasChange("min_prefixlen") {
		request.MinPrefixLen = vpcInt(d.Get("min_prefixlen").(int))
		needsUpdate = true
	}
	if d.HasChange("max_prefixlen") {
		request.MaxPrefixLen = vpcInt(d.Get("max_prefixlen").(int))
		needsUpdate = true
	}
	if d.HasChange("address_scope_id") {
		// An empty value detaches the pool from its address scope, which the
		// API expresses as an explicit JSON null.
		if scopeID := d.Get("address_scope_id").(string); scopeID == "" {
			request.AddressScopeID = vpc.Null[string]()
		} else {
			request.AddressScopeID = vpc.Value(scopeID)
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking subnetpool %s", d.Id())
		if _, err := subnetpool.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetPoolV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := subnetpool.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingSubnetPoolV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingSubnetPoolV2Read(ctx, d, meta)
}

func resourceNetworkingSubnetPoolV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking subnetpool %s", d.Id())
	if err := subnetpool.Delete(ctx, client, d.Id()); err != nil {
		// A pool still used by subnets is refused with a conflict, which stays
		// an error and keeps the resource in the state.
		return networkingCheckDeletedOnDelete(d, err, networkingSubnetPoolV2Object, func() error {
			_, confirmErr := subnetpool.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingSubnetPoolV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// flattenNetworkingSubnetPoolV2 writes an API subnet pool into the state.
func flattenNetworkingSubnetPoolV2(d *schema.ResourceData, found *subnetpool.SubnetPool) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("default_quota", derefInt(found.DefaultQuota))
	d.Set("default_prefixlen", found.DefaultPrefixLen)
	d.Set("min_prefixlen", found.MinPrefixLen)
	d.Set("max_prefixlen", found.MaxPrefixLen)
	d.Set("address_scope_id", derefString(found.AddressScopeID))
	d.Set("ip_version", found.IPVersion)
	d.Set("shared", found.Shared)
	d.Set("is_default", found.IsDefault)
	d.Set("created_at", found.CreatedAt)
	d.Set("updated_at", found.UpdatedAt)
	d.Set("revision_number", found.RevisionNumber)

	if err := d.Set("prefixes", found.Prefixes); err != nil {
		log.Print(errSettingComplexAttr("prefixes", err))
	}
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}
