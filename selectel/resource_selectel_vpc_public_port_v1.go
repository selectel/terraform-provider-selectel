package selectel

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

var publicPortDocs = resourceDocs{Name: "public port"}

func resourceVPCPublicPortV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPublicPortV1Create,
		ReadContext:   resourceVPCPublicPortV1Read,
		UpdateContext: resourceVPCPublicPortV1Update,
		DeleteContext: resourceVPCPublicPortV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPublicPortV1ImportState,
		},
		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return publicPortDocs.identitySchemaByID(exampleResourceID, publicPortIDFromControlPanel, publicPortRegionFromControlPanel)
			},
		},
		Description: "Creates and manages a direct public IP address (public port) in VPC using the v1 public API. " +
			"For more information about direct public IP addresses, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/direct-public-ip-addresses).",
		Schema: publicPortDocs.withDocsHints(map[string]*schema.Schema{
			"id":         publicPortDocs.idResourceSchema(),
			"region":     publicPortDocs.regionResourceSchema(),
			"project_id": publicPortDocs.projectIDResourceSchema(),
			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier of the service network from which the public port is allocated.",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Direct public IP address assigned to the public port.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Description of the public port.",
			},
			"subnet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CIDR of the subnet from which the public port IP address is allocated.",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of the subnet gateway.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Specifies if the port is administratively up.",
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: 20,
				Description: "Unique identifiers of security groups associated with the public port. " +
					"If not specified, the port gets the default security group of the project. " +
					secGroupIDFromResource,
			},
		}),
	}
}

func resourceVPCPublicPortV1Create(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	client, diagErr := getPublicNetAPIClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	dto := &publicnetapi.PortCreateDTO{ProjectID: d.Get("project_id").(string)}

	description := d.Get("description").(string)
	dto.Description = &description

	adminStateUp := d.Get("admin_state_up").(bool)
	dto.AdminStateUp = &adminStateUp

	if schemaconfig.IsSet(d, "security_group_ids") {
		dto.SecurityGroupIDs = expandStringList(d.Get("security_group_ids").([]any))
	}

	log.Print(msgCreate(objectPublicPort, dto))

	port, err := client.CreatePort(ctx, dto)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectPublicPort, err))
	}

	d.SetId(port.ID)
	fillVPCPublicPortData(port, d)

	return nil
}

func resourceVPCPublicPortV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diagErr := getPublicNetAPIClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectPublicPort, d.Id()))

	port, err := client.GetPort(ctx, d.Id())
	if err != nil {
		if isNotFound(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(errGettingObject(objectPublicPort, d.Id(), err))
	}

	fillVPCPublicPortData(port, d)

	return nil
}

func resourceVPCPublicPortV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if !d.HasChanges("description", "admin_state_up", "security_group_ids") {
		return nil
	}

	client, diagErr := getPublicNetAPIClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	dto := publicnetapi.PortUpdateDTO{}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		dto.Description = &description
	}

	if d.HasChange("admin_state_up") {
		adminStateUp := d.Get("admin_state_up").(bool)
		dto.AdminStateUp = &adminStateUp
	}

	if d.HasChange("security_group_ids") {
		securityGroupIDsVal := d.Get("security_group_ids")
		if securityGroupIDsVal != nil {
			dto.SecurityGroupIDs = expandStringList(securityGroupIDsVal.([]any))
		}
	}

	log.Print(msgUpdate(objectPublicPort, d.Id(), dto))

	port, err := client.UpdatePort(ctx, d.Id(), &dto)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectPublicPort, d.Id(), err))
	}

	fillVPCPublicPortData(port, d)

	return nil
}

func resourceVPCPublicPortV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diagErr := getPublicNetAPIClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectPublicPort, d.Id()))

	err := client.DeletePort(ctx, d.Id(), nil)
	if err != nil {
		if isNotFound(err) {
			return nil
		}

		return diag.FromErr(errDeletingObject(objectPublicPort, d.Id(), err))
	}

	return nil
}

func resourceVPCPublicPortV1ImportState(
	_ context.Context,
	d *schema.ResourceData,
	meta any,
) ([]*schema.ResourceData, error) {
	if err := importProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func fillVPCPublicPortData(port *publicnetapi.Port, d *schema.ResourceData) {
	_ = d.Set("project_id", port.ProjectID)
	_ = d.Set("network_id", port.NetworkID)
	_ = d.Set("ip_address", port.IPAddress)
	_ = d.Set("description", port.Description)
	_ = d.Set("subnet", port.Subnet)
	_ = d.Set("gateway", port.Gateway)
	_ = d.Set("admin_state_up", port.AdminStateUp)
	_ = d.Set("security_group_ids", port.SecurityGroupIDs)

	setProjectRegionIdentity(d)
}

func expandStringList(raw []any) []string {
	result := make([]string, len(raw))

	for i, v := range raw {
		result[i] = v.(string)
	}

	return result
}

func isNotFound(err error) bool {
	var apiErr *publicnetapi.APIErr

	return errors.As(err, &apiErr) &&
		apiErr.Code == http.StatusNotFound
}
