package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/secretsmanager-go/service/certs"
)

func resourceSecretsManagerCertificateV1() *schema.Resource {
	return &schema.Resource{
		Description: "represents a Certificate — entity from SecretsManager service",

		CreateContext: resourceSecretsManagerCertificateV1Create,
		ReadContext:   resourceSecretsManagerCertificateV1Read,
		UpdateContext: resourceSecretsManagerCertificateV1Update,
		DeleteContext: resourceSecretsManagerCertificateV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSecretsManagerCertificateV1ImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "name of the certificate",
				Type:        schema.TypeString,
				Required:    true,
			},
			"certificates": {
				Description: "ca_chain list of public certs for certificate",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Description: "certificate — that must begin with -----BEGIN CERTIFICATE----- and end with -----END CERTIFICATE-----",
					Type:        schema.TypeString,
				},
				Required: true,
			},
			"private_key": {
				Description: "that should start with -----BEGIN PRIVATE KEY----- and end with -----END PRIVATE KEY-----",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"project_id": {
				Description: "id of a project where secret is used",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dns_names": {
				Description: "computed list of Subject Alternative Names",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"id": {
				Description: "computed id of a certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"issued_by": {
				Description: "information that is incorporated into certificate",
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"locality": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"serial_number": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"street_address": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
					},
				},
				Computed: true,
			},
			"serial": {
				Description: "number written in the certificate that was chosen by the CA which issued the certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"validity": {
				Description: "validity of a certificate in terms of notBefore and notAfter timestamps.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"basic_constraints": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"not_after": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_before": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"version": {
				Description: "of the certificate",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceSecretsManagerCertificateV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	name := d.Get("name").(string)
	rawCertificates := d.Get("certificates").([]interface{})
	certificates := convertToStringSlice(rawCertificates)

	privateKey := d.Get("private_key").(string)

	cert := certs.CreateCertificateRequest{
		Name: name,
		Pem: certs.Pem{
			Certificates: certificates,
			PrivateKey:   privateKey,
		},
	}

	log.Print(msgCreate(objectCertificate, cert.Name))

	createdCert, errCr := cl.Certificates.Create(ctx, cert)
	if errCr != nil {
		return diag.FromErr(errCreatingObject(objectCertificate, errCr))
	}

	d.SetId(createdCert.ID)

	return resourceSecretsManagerCertificateV1Read(ctx, d, meta)
}

func resourceSecretsManagerCertificateV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectCertificate, d.Id()))
	cert, errGet := cl.Certificates.Get(ctx, d.Get("id").(string))
	if errGet != nil {
		return diag.FromErr(errGettingObject(objectCertificate, d.Id(), errGet))
	}

	d.Set("dns_names", cert.DNSNames)
	d.Set("issued_by", convertSMIssuedByToList(cert.IssuedBy))
	d.Set("serial", cert.Serial)
	d.Set("validity", convertSMValidityToList(cert.Validity))
	d.Set("version", cert.Version)
	// Set fields in case called from resourceSecretsManagerCertificateV1ImportState.
	d.Set("name", cert.Name)
	d.Set("id", cert.ID)

	return nil
}

func resourceSecretsManagerCertificateV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)

		log.Print(msgUpdate(objectCertificate, d.Id(), newName))

		errUpd := cl.Certificates.UpdateName(ctx, d.Id(), newName)
		if errUpd != nil {
			return diag.FromErr(errUpdatingObject(objectCertificate, d.Id(), errUpd))
		}
	}

	if d.HasChange("certificates") || d.HasChange("private_key") {
		rawCertificates := d.Get("certificates").([]interface{})
		certificates := convertToStringSlice(rawCertificates)

		upd := certs.UpdateCertificateVersionRequest{
			Pem: certs.Pem{
				Certificates: certificates,
				PrivateKey:   d.Get("private_key").(string),
			},
		}

		log.Print(msgUpdate(objectCertificate, d.Id(), "updated pem"))

		errUpd := cl.Certificates.UpdateVersion(ctx, d.Id(), upd)
		if errUpd != nil {
			return diag.FromErr(errUpdatingObject(objectCertificate, d.Id(), errUpd))
		}
	}

	return resourceSecretsManagerCertificateV1Read(ctx, d, meta)
}

func resourceSecretsManagerCertificateV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectCertificate, d.Id()))

	errDel := cl.Certificates.Delete(ctx, d.Id())
	if errDel != nil {
		return diag.FromErr(errDeletingObject(objectCertificate, d.Id(), errDel))
	}

	return nil
}

// resourceSecretsManagerCertificateV1ImportState —  helper used in Importer: &schema.ResourceImporter
// to avoid difficulties occurred with required INFRA_PROJECT_ID env in
// resourceSecretsManagerCertificateV1Read when uising schema.ImportStatePassthroughContext.
func resourceSecretsManagerCertificateV1ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("INFRA_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)

	log.Print(msgImport(objectCertificate, d.Id()))
	resourceSecretsManagerCertificateV1Read(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

// convertSMIssuedByToList — helper for setting "issued_by" attribute with nested structure.
func convertSMIssuedByToList(ib certs.IssuedBy) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"country":        convertToInterfaceSlice(ib.Country),
			"locality":       convertToInterfaceSlice(ib.Locality),
			"serial_number":  ib.SerialNumber,
			"street_address": convertToInterfaceSlice(ib.StreetAddress),
		},
	}
}

// convertValidityToList — helper for setting "validity" attribute with nested structure.
func convertSMValidityToList(validity certs.Validity) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"basic_constraints": validity.BasicConstraints,
			"not_after":         validity.NotAfter,
			"not_before":        validity.NotBefore,
		},
	}
}
