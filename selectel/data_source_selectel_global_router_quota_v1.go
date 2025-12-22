package selectel

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

func dataSourceGlobalRouterQuotaV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalRouterQuotaV1Read,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"scope_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceGlobalRouterQuotaV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}
	quotaName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterQuota, quotaName, errors.New("'name' should have type string")),
		)
	}
	scope, ok := d.Get("scope").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterQuota, quotaName, errors.New("'scope' should have type string")),
		)
	}
	scopeValue, ok := d.Get("scope_value").(string)
	if !ok {
		return diag.FromErr(
			errGettingObject(objectGlobalRouterQuota, quotaName, errors.New("'scope_value' should have type string")),
		)
	}

	quota, err := getQuotaByParams(ctx, client, quotaName, scope, scopeValue)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterQuota, quotaName, err))
	}

	err = setGRQuotaToResourceData(d, quota)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterQuota, quotaName, err))
	}

	return nil
}

func setGRQuotaToResourceData(d *schema.ResourceData, quota *globalrouter.Quota) error {
	d.SetId(quota.ID)
	d.Set("name", quota.Name)
	d.Set("scope", quota.Scope)
	d.Set("scope_value", quota.ScopeValue)
	d.Set("limit", quota.Limit)

	return nil
}
