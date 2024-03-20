package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"unicode"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient/resell/v2/users"
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

func resourceVPCUserV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for user object: %w", err))
	}

	opts := users.UserOpts{
		Name:     d.Get("name").(string),
		Password: d.Get("password").(string),
	}

	log.Print(msgCreate(objectUser, opts))
	user, _, err := users.Create(selvpcClient, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectUser, err))
	}

	d.SetId(user.ID)

	return resourceVPCUserV2Read(ctx, d, meta)
}

func resourceVPCUserV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for user object: %w", err))
	}

	log.Print(msgGet(objectUser, d.Id()))
	user, response, err := users.Get(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectUser, d.Id(), err))
	}

	d.Set("name", user.Name)
	d.Set("enabled", user.Enabled)

	return nil
}

func resourceVPCUserV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for user object: %w", err))
	}

	enabled := d.Get("enabled").(bool)
	opts := users.UserOpts{
		Name:     d.Get("name").(string),
		Password: d.Get("password").(string),
		Enabled:  &enabled,
	}

	log.Print(msgUpdate(objectUser, d.Id(), opts))
	_, _, err = users.Update(selvpcClient, d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectUser, d.Id(), err))
	}

	return resourceVPCUserV2Read(ctx, d, meta)
}

func resourceVPCUserV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for user object: %w", err))
	}

	log.Print(msgDelete(objectUser, d.Id()))
	response, err := users.Delete(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectUser, d.Id(), err))
	}

	return nil
}
