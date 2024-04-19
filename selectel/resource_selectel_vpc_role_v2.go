package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVPCRoleV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCRoleV2Create,
		ReadContext:   resourceVPCRoleV2Read,
		DeleteContext: resourceVPCRoleV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCRoleV2Create(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_role_v2"))
}

func resourceVPCRoleV2Read(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_role_v2"))
}

func resourceVPCRoleV2Delete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_role_v2"))
}
