package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/federations/saml/certificates"
)

func resourceIAMSAMLFederationCertificateV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a SAML Federation Certificate in IAM API",
		CreateContext: resourceIAMSAMLFederationCertificateV1Create,
		ReadContext:   resourceIAMSAMLFederationCertificateV1Read,
		UpdateContext: resourceIAMSAMLFederationCertificateV1Update,
		DeleteContext: resourceIAMSAMLFederationCertificateV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceIAMS3SAMLFederationCertificateV1ImportState,
		},
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Federation ID to create Certificate for.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Certificate.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Description of the Certificate.",
			},
			"data": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Certificate issued on the provider side. It must begin with -----BEGIN CERTIFICATE----- and end with -----END CERTIFICATE-----.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Account ID.",
			},
			"not_before": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Certificate lifetime left bound.",
			},
			"not_after": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Certificate lifetime right bound.",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Fingerprint.",
			},
		},
	}
}

func resourceIAMSAMLFederationCertificateV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	opts := certificates.CreateRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Data:        d.Get("data").(string),
	}
	log.Print(msgCreate(objectSAMLFederationCertificate, opts))

	certificate, err := iamClient.SAMLFederations.Certificates.Create(ctx, d.Get("federation_id").(string), opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectSAMLFederationCertificate, err))
	}

	d.SetId(certificate.ID)

	return resourceIAMSAMLFederationCertificateV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationCertificateV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectSAMLFederationCertificate, d.Id()))
	certificate, err := iamClient.SAMLFederations.Certificates.Get(ctx, d.Get("federation_id").(string), d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectSAMLFederationCertificate, d.Id(), err))
	}

	d.Set("account_id", certificate.AccountID)
	d.Set("name", certificate.Name)
	d.Set("description", certificate.Description)
	d.Set("not_before", certificate.NotBefore)
	d.Set("not_after", certificate.NotAfter)
	d.Set("fingerprint", certificate.Fingerprint)
	d.Set("data", certificate.Data)

	return nil
}

func resourceIAMSAMLFederationCertificateV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	desc := d.Get("description").(string)

	opts := certificates.UpdateRequest{
		Name:        d.Get("name").(string),
		Description: &desc,
	}

	log.Print(msgUpdate(objectSAMLFederationCertificate, d.Id(), opts))
	_, err := iamClient.SAMLFederations.Certificates.Update(ctx, d.Get("federation_id").(string), d.Id(), opts)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectSAMLFederationCertificate, d.Id(), err))
	}

	return resourceIAMSAMLFederationCertificateV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationCertificateV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectSAMLFederationCertificate, d.Id()))
	err := iamClient.SAMLFederations.Certificates.Delete(ctx, d.Get("federation_id").(string), d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrFederationCertificateNotFound) {
		return diag.FromErr(errDeletingObject(objectSAMLFederationCertificate, d.Id(), err))
	}

	return nil
}

func resourceIAMS3SAMLFederationCertificateV1ImportState(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	var v string
	if v = os.Getenv("OS_SAML_FEDERATION_ID"); v == "" {
		return nil, fmt.Errorf("no OS_SAML_FEDERATION_ID environment variable was found, provide one to use import")
	}

	d.Set("federation_id", v)

	return []*schema.ResourceData{d}, nil
}
