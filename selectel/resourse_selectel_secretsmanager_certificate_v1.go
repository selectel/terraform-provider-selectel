package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/secretsmanager-go/secretsmanagererrors"
	"github.com/selectel/secretsmanager-go/service/certs"
)

func resourceSecretsManagerCertificateV1() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages a certificate in Certificates Manager using public API v1.",

		CreateContext: resourceSecretsManagerCertificateV1Create,
		ReadContext:   resourceSecretsManagerCertificateV1Read,
		UpdateContext: resourceSecretsManagerCertificateV1Update,
		DeleteContext: resourceSecretsManagerCertificateV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSecretsManagerCertificateV1ImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Certificate name. Changing only the name does not create a version.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"certificates": {
				Description: "PEM certificate chain. Changing certificates or private_key uploads both configured values as a new version; current_version controls activation.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Description: "PEM certificate beginning with -----BEGIN CERTIFICATE----- and ending with -----END CERTIFICATE-----.",
					Type:        schema.TypeString,
				},
				Required: true,
			},
			"private_key": {
				Description: "PEM-encoded private key matching the primary certificate, in PKCS#8, PKCS#1, or EC format. Changing the key uploads a new version with the configured certificates.", // nosemgrep: generic.secrets.gitleaks.private-key.private-key
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"project_id": {
				Description: "Unique identifier of the associated project. Changing this creates a new certificate and deletes the previous certificate and all its versions.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dns_names": {
				Description: "Domain names covered by the active leaf certificate. The common name is returned when Subject Alternative Names are absent.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"id": {
				Description: "Unique identifier of the certificate.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"issued_by": {
				Description: "Active certificate issuer information: country, locality, serial_number, and street_address.",
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
				Description: "Active certificate serial number assigned by its issuing Certificate Authority.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"validity": {
				Description: "Active certificate validity: not_before and not_after timestamps in RFC3339 format, and the API basic_constraints flag.",
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
				Description: "Observed active certificate version number, not the latest uploaded version.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"current_version": {
				Description:  "Desired active version. Defaults to 0, which activates new versions on content changes. Must be non-negative; only 0 or 1 is valid at creation. Positive values pin an existing version and keep new uploads inactive. Returning to 0 leaves the active version unchanged until the next content change.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

func resourceSecretsManagerCertificateV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if d.Get("current_version").(int) > 1 {
		return diag.Errorf("current_version must be 0 or 1 when creating a certificate")
	}

	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	name := d.Get("name").(string)
	rawCertificates := d.Get("certificates").([]any)
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
	d.Set("version", createdCert.Version)

	return resourceSecretsManagerCertificateV1Read(ctx, d, meta)
}

func resourceSecretsManagerCertificateV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectCertificate, d.Id()))
	cert, errGet := cl.Certificates.Get(ctx, d.Id())
	if errGet != nil {
		return diag.FromErr(errGettingObject(objectCertificate, d.Id(), errGet))
	}

	d.Set("dns_names", cert.DNSNames)
	d.Set("issued_by", convertSMIssuedByToList(cert.IssuedBy))
	d.Set("serial", cert.Serial)
	d.Set("validity", convertSMValidityToList(cert.Validity))
	d.Set("version", cert.Version)
	if d.Get("current_version").(int) > 0 {
		d.Set("current_version", cert.Version)
	}
	// Set fields in case called from resourceSecretsManagerCertificateV1ImportState.
	d.Set("name", cert.Name)
	d.Set("id", cert.ID)

	return nil
}

func resourceSecretsManagerCertificateV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		d.Partial(true)
		return diagErr
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)

		log.Print(msgUpdate(objectCertificate, d.Id(), newName))

		errUpd := cl.Certificates.UpdateName(ctx, d.Id(), newName)
		if errUpd != nil {
			d.Partial(true)
			return diag.FromErr(errUpdatingObject(objectCertificate, d.Id(), errUpd))
		}
	}

	currentVersion := d.Get("current_version").(int)
	if d.HasChange("certificates") || d.HasChange("private_key") {
		rawCertificates := d.Get("certificates").([]any)
		certificates := convertToStringSlice(rawCertificates)

		upd := certs.UpdateCertificateVersionRequest{
			Pem: certs.Pem{
				Certificates: certificates,
				PrivateKey:   d.Get("private_key").(string),
			},
		}

		log.Print(msgUpdate(objectCertificate, d.Id(), "updated pem"))

		version, errUpd := cl.Certificates.CreateVersion(ctx, d.Id(), upd, currentVersion == 0)
		if errUpd != nil {
			restoreSecretsManagerInputs(d, "certificates", "private_key", "current_version")
			return diag.FromErr(errUpdatingObject(objectCertificate, d.Id(), errUpd))
		}
		if currentVersion == 0 {
			d.Set("version", version.Version)
		}
	}

	if currentVersion > 0 {
		cert, err := cl.Certificates.Get(ctx, d.Id())
		if err != nil {
			restoreSecretsManagerInputs(d, "current_version")
			return diag.FromErr(errGettingObject(objectCertificate, d.Id(), err))
		}
		d.Set("version", cert.Version)
		d.Set("current_version", cert.Version)
		if cert.Version != int64(currentVersion) {
			activated, err := cl.Certificates.ActivateVersion(ctx, d.Id(), uint64(currentVersion))
			if errors.Is(err, secretsmanagererrors.ErrConflictStatusText) {
				refreshed := resourceSecretsManagerCertificateV1Read(ctx, d, meta)
				if !refreshed.HasError() && d.Get("version").(int) == currentVersion {
					return nil
				}
			}
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectCertificate, d.Id(), err))
			}
			d.Set("version", activated.Version)
			d.Set("current_version", activated.Version)
		}
	}

	return resourceSecretsManagerCertificateV1Read(ctx, d, meta)
}

func resourceSecretsManagerCertificateV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
func resourceSecretsManagerCertificateV1ImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
func convertSMIssuedByToList(ib certs.IssuedBy) []any {
	return []any{
		map[string]any{
			"country":        convertToInterfaceSlice(ib.Country),
			"locality":       convertToInterfaceSlice(ib.Locality),
			"serial_number":  ib.SerialNumber,
			"street_address": convertToInterfaceSlice(ib.StreetAddress),
		},
	}
}

// convertValidityToList — helper for setting "validity" attribute with nested structure.
func convertSMValidityToList(validity certs.Validity) []any {
	return []any{
		map[string]any{
			"basic_constraints": validity.BasicConstraints,
			"not_after":         validity.NotAfter,
			"not_before":        validity.NotBefore,
		},
	}
}
