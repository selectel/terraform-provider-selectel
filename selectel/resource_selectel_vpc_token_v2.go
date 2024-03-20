package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVPCTokenV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCTokenV2Create,
		ReadContext:   resourceVPCTokenV2Read,
		DeleteContext: resourceVPCTokenV2Delete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"account_name"},
				Optional:      true,
				ForceNew:      true,
			},
			"account_name": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"project_id"},
				Optional:      true,
				ForceNew:      true,
			},
		},
	}
}

func resourceVPCTokenV2Create(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_token_v2"))
}

func resourceVPCTokenV2Read(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_token_v2"))
}

func resourceVPCTokenV2Delete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_token_v2"))
}
