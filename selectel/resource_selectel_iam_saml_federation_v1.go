package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/federations/saml"
)

func resourceIAMSAMLFederationV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a SAML Federation in IAM API",
		CreateContext: resourceIAMSAMLFederationV1Create,
		ReadContext:   resourceIAMSAMLFederationV1Read,
		UpdateContext: resourceIAMSAMLFederationV1Update,
		DeleteContext: resourceIAMSAMLFederationV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Federation.",
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
			"sso_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Single sign-on endpoint URL.",
			},
			"sign_authn_requests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should sign authentication requests.",
			},
			"session_max_age_hours": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Session lifetime.",
			},
			"force_authn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable forced authentication at every login.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Account ID",
			},
		},
	}
}

func resourceIAMSAMLFederationV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	opts := saml.CreateRequest{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Issuer:             d.Get("issuer").(string),
		SSOUrl:             d.Get("sso_url").(string),
		SignAuthnRequests:  d.Get("sign_authn_requests").(bool),
		ForceAuthn:         d.Get("force_authn").(bool),
		SessionMaxAgeHours: d.Get("session_max_age_hours").(int),
	}
	log.Print(msgCreate(objectSAMLFederation, opts))

	federation, err := iamClient.SAMLFederations.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectSAMLFederation, err))
	}

	d.SetId(federation.ID)

	return resourceIAMSAMLFederationV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectSAMLFederation, d.Id()))
	federation, err := iamClient.SAMLFederations.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectSAMLFederation, d.Id(), err))
	}

	d.Set("account_id", federation.AccountID)
	d.Set("name", federation.Name)
	d.Set("description", federation.Description)
	d.Set("issuer", federation.Issuer)
	d.Set("sso_url", federation.SSOUrl)
	d.Set("sign_authn_requests", federation.SignAuthnRequests)
	d.Set("force_authn", federation.ForceAuthn)
	d.Set("session_max_age_hours", federation.SessionMaxAgeHours)

	return nil
}

func resourceIAMSAMLFederationV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	signAuthnRequests := d.Get("sign_authn_requests").(bool)
	forceAuthn := d.Get("force_authn").(bool)
	description := d.Get("description").(string)

	opts := saml.UpdateRequest{
		Name:               d.Get("name").(string),
		Description:        &description,
		Issuer:             d.Get("issuer").(string),
		SSOUrl:             d.Get("sso_url").(string),
		SignAuthnRequests:  &signAuthnRequests,
		ForceAuthn:         &forceAuthn,
		SessionMaxAgeHours: d.Get("session_max_age_hours").(int),
	}

	log.Print(msgUpdate(objectSAMLFederation, d.Id(), opts))
	err := iamClient.SAMLFederations.Update(ctx, d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectSAMLFederation, d.Id(), err))
	}

	return resourceIAMSAMLFederationV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectSAMLFederation, d.Id()))
	err := iamClient.SAMLFederations.Delete(ctx, d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrFederationNotFound) {
		return diag.FromErr(errDeletingObject(objectSAMLFederation, d.Id(), err))
	}

	return nil
}
