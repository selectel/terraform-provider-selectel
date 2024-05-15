package selectel

import (
	"context"
	"unicode"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVPCUserV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCUserV2Create,
		ReadContext:   resourceVPCUserV2Read,
		UpdateContext: resourceVPCUserV2Update,
		DeleteContext: resourceVPCUserV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				ValidateDiagFunc: func(i interface{}, _ cty.Path) diag.Diagnostics {
					password := i.(string)
					if len(password) < 8 {
						return diag.Errorf("password must be at least 8 characters long")
					}

					chrType := 0
					for _, r := range password {
						switch {
						case unicode.IsDigit(r):
							chrType |= 1
						case unicode.IsLower(r):
							chrType |= 2
						case unicode.IsUpper(r):
							chrType |= 4
						}
					}
					if chrType != 7 {
						return diag.Errorf("password must contain at least one digit, one lowercase and one uppercase character")
					}

					return nil
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
		},
	}
}

func resourceVPCUserV2Create(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_user_v2"))
}

func resourceVPCUserV2Read(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_user_v2"))
}

func resourceVPCUserV2Update(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_user_v2"))
}

func resourceVPCUserV2Delete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.FromErr(errResourceDeprecated("selectel_vpc_user_v2"))
}
