package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/craas-go/pkg/svc"
	tokenv2 "github.com/selectel/craas-go/pkg/v2/token"
)

func resourceCRaaSTokenV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCRaaSTokenV2Create,
		ReadContext:   resourceCRaaSTokenV2Read,
		UpdateContext: resourceCRaaSTokenV2Update,
		DeleteContext: resourceCRaaSTokenV2Delete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"mode_rw": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"all_registries": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"registry_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: false,
			},
			"is_set": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"expires_at": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				Optional:     true,
				ForceNew:     false,
			},
			"username": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
		},
	}
}

func resourceCRaaSTokenV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClientV2(d, meta)
	if diagErr != nil {
		return diagErr
	}

	name := d.Get("name").(string)
	modeRW := d.Get("mode_rw").(bool)
	allRegistries := d.Get("all_registries").(bool)
	isSet := d.Get("is_set").(bool)
	registries := d.Get("registry_ids").([]interface{})
	registriesIDs := make([]string, len(registries))
	for i, v := range registries {
		registriesIDs[i] = v.(string)
	}
	expiresAt := d.Get("expires_at").(string)
	expires, err := time.Parse("2006-01-02T00:00:00Z", expiresAt)
	if err != nil {
		return diag.FromErr(err)
	}
	createOpts := &tokenv2.TokenV2{
		Name: name,
		Scope: tokenv2.Scope{
			ModeRW:        modeRW,
			AllRegistries: allRegistries,
			RegistryIDs:   registriesIDs,
		},
		Expiration: tokenv2.Expiration{
			IsSet:     isSet,
			ExpiresAt: expires,
		},
	}

	log.Print(msgCreate(objectRegistryToken, createOpts))
	dockerCfg := false
	newToken, res, err := tokenv2.Create(ctx, craasClient, createOpts, &dockerCfg)
	if res != nil && res.Err != nil {
		return diag.FromErr(errCreatingObject(objectRegistryToken, res.Err))
	}
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRegistryToken, err))
	}
	d.SetId(newToken.ID)
	d.Set("token", newToken.Token)

	return resourceCRaaSTokenV2Read(ctx, d, meta)
}

func resourceCRaaSTokenV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClientV2(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectRegistryToken, d.Id()))
	tokenObj, response, err := tokenv2.GetByID(ctx, craasClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectRegistryToken, d.Id(), err))
	}

	if remove, reason := shouldRemoveCRaaSTokenV2FromState(tokenObj); remove {
		log.Printf("[WARN] CRaaS token %s: %s, removing from state", d.Id(), reason)
		d.SetId("")
		return nil
	}

	d.Set("username", craasV1TokenUsername)
	d.Set("token", d.Get("token").(string))

	return nil
}

func resourceCRaaSTokenV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClientV2(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectRegistryToken, d.Id()))
	response, err := tokenv2.Delete(ctx, craasClient, d.Id())
	if err != nil {
		if isCRaaSTokenV2DeleteNotFound(response) {
			return nil
		}

		return diag.FromErr(errDeletingObject(objectRegistryToken, d.Id(), err))
	}

	return nil
}

func resourceCRaaSTokenV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClientV2(d, meta)
	if diagErr != nil {
		return diagErr
	}
	var (
		sc  tokenv2.Scope
		exp tokenv2.Expiration
	)
	name := d.Get("name").(string)
	sc.AllRegistries = d.Get("all_registries").(bool)
	registries := d.Get("registry_ids").([]interface{})
	sc.RegistryIDs = make([]string, len(registries))
	for i, v := range registries {
		sc.RegistryIDs[i] = v.(string)
	}
	sc.ModeRW = d.Get("mode_rw").(bool)
	exp.IsSet = d.Get("is_set").(bool)
	expiresAt := d.Get("expires_at").(string)
	expires, err := time.Parse("2006-01-02T00:00:00Z", expiresAt)
	if err != nil {
		return diag.FromErr(err)
	}
	exp.ExpiresAt = expires

	log.Print(msgUpdate(objectRegistryToken, d.Id(), sc))
	log.Print(msgUpdate(objectRegistryToken, d.Id(), exp))
	_, res, err := tokenv2.Patch(ctx, craasClient, d.Id(), name, sc, exp)
	if res != nil && res.Err != nil {
		return diag.FromErr(errUpdatingObject(objectRegistryToken, d.Id(), res.Err))
	}
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectRegistryToken, d.Id(), err))
	}

	return resourceCRaaSTokenV2Read(ctx, d, meta)
}

func shouldRemoveCRaaSTokenV2FromState(tokenObj *tokenv2.TokenV2) (bool, string) {
	return shouldRemoveCRaaSTokenV2FromStateAt(tokenObj, time.Now())
}

func shouldRemoveCRaaSTokenV2FromStateAt(tokenObj *tokenv2.TokenV2, now time.Time) (bool, string) {
	if tokenObj.Status != tokenv2.StatusActive {
		return true, fmt.Sprintf("has non-active status %q", tokenObj.Status)
	}

	if tokenObj.Expiration.IsSet && tokenObj.Expiration.ExpiresAt.Before(now) {
		return true, fmt.Sprintf("expired at %s", tokenObj.Expiration.ExpiresAt.Format(time.RFC3339))
	}

	return false, ""
}

func isCRaaSTokenV2DeleteNotFound(response *svc.ResponseResult) bool {
	return response != nil && response.StatusCode == http.StatusNotFound
}
