package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/federations/oidc"
)

func resourceIAMOIDCFederationV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents an OIDC Federation in IAM API",
		CreateContext: resourceIAMOIDCFederationV1Create,
		ReadContext:   resourceIAMOIDCFederationV1Read,
		UpdateContext: resourceIAMOIDCFederationV1Update,
		DeleteContext: resourceIAMOIDCFederationV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Federation.",
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Alias of the Federation.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Description of the Federation.",
			},
			"issuer": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the credential provider.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client ID for OIDC authentication.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Client secret for OIDC authentication.",
			},
			"auth_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Authorization endpoint URL.",
			},
			"token_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Token endpoint URL.",
			},
			"jwks_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JWKS endpoint URL.",
			},
			"auto_users_creation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable automatic creation of users for this Federation.",
			},
			"enable_group_mappings": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable group mappings for this Federation.",
			},
			"session_max_age_hours": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Session lifetime.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Account ID",
			},
		},
	}
}

func resourceIAMOIDCFederationV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	opts := oidc.CreateRequest{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Alias:              d.Get("alias").(string),
		Issuer:             d.Get("issuer").(string),
		ClientID:           d.Get("client_id").(string),
		ClientSecret:       d.Get("client_secret").(string),
		AuthURL:            d.Get("auth_url").(string),
		TokenURL:           d.Get("token_url").(string),
		JWKSURL:            d.Get("jwks_url").(string),
		SessionMaxAgeHours: d.Get("session_max_age_hours").(int),
		AutoUsersCreation:  d.Get("auto_users_creation").(bool),
		EnableGroupMapping: d.Get("enable_group_mappings").(bool),
	}
	log.Print(msgCreate(objectOIDCFederation, opts))

	federation, err := iamClient.OIDCFederations.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectOIDCFederation, err))
	}

	d.SetId(federation.ID)

	return resourceIAMOIDCFederationV1Read(ctx, d, meta)
}

func resourceIAMOIDCFederationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectOIDCFederation, d.Id()))
	federation, err := iamClient.OIDCFederations.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectOIDCFederation, d.Id(), err))
	}

	d.Set("account_id", federation.AccountID)
	d.Set("name", federation.Name)
	d.Set("alias", federation.Alias)
	d.Set("description", federation.Description)
	d.Set("issuer", federation.Issuer)
	d.Set("client_id", federation.ClientID)
	d.Set("client_secret", federation.ClientSecret)
	d.Set("auth_url", federation.AuthURL)
	d.Set("token_url", federation.TokenURL)
	d.Set("jwks_url", federation.JWKSURL)
	d.Set("auto_users_creation", federation.AutoUsersCreation)
	d.Set("enable_group_mappings", federation.EnableGroupMapping)
	d.Set("session_max_age_hours", federation.SessionMaxAgeHours)

	return nil
}

func resourceIAMOIDCFederationV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	description := d.Get("description").(string)
	autoUsersCreation := d.Get("auto_users_creation").(bool)
	enableGroupMappings := d.Get("enable_group_mappings").(bool)

	opts := oidc.UpdateRequest{
		Name:               d.Get("name").(string),
		Description:        &description,
		Alias:              d.Get("alias").(string),
		Issuer:             d.Get("issuer").(string),
		ClientID:           d.Get("client_id").(string),
		ClientSecret:       d.Get("client_secret").(string),
		AuthURL:            d.Get("auth_url").(string),
		TokenURL:           d.Get("token_url").(string),
		JWKSURL:            d.Get("jwks_url").(string),
		SessionMaxAgeHours: d.Get("session_max_age_hours").(int),
		AutoUsersCreation:  &autoUsersCreation,
		EnableGroupMapping: &enableGroupMappings,
	}

	log.Print(msgUpdate(objectOIDCFederation, d.Id(), opts))
	err := iamClient.OIDCFederations.Update(ctx, d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectOIDCFederation, d.Id(), err))
	}

	return resourceIAMOIDCFederationV1Read(ctx, d, meta)
}

func resourceIAMOIDCFederationV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectOIDCFederation, d.Id()))
	err := iamClient.OIDCFederations.Delete(ctx, d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrFederationNotFound) {
		return diag.FromErr(errDeletingObject(objectOIDCFederation, d.Id(), err))
	}

	return nil
}
