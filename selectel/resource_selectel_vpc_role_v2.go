package selectel

import (
	"context"
	"fmt"
	"strings"

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

func resourceVPCRoleV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unable to parse id: '%s'", id)
	}

	return idParts[0], idParts[1], nil
}
