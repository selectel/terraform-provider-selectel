package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
)

func resourceVPCPublicPortV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPublicPortV1Create,
		ReadContext:   resourceVPCPublicPortV1Read,
		UpdateContext: resourceVPCPublicPortV1Update,
		DeleteContext: resourceVPCPublicPortV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPublicPortV1ImportState,
		},
		Description: "Creates and manages a direct public IP address (public port) in VPC using public API v1. " +
			"For more information about direct public IP address, see the " +
			"[official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/direct-public-ip-addresses).",
		Schema: withDocsHints("public port", map[string]*schema.Schema{
			"region":     regionSchema("public port", "ru-6"),
			"project_id": projectIDSchema(),
			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the service network to which the public port is attached.",
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
				Description: "Public port description.",
			},
			"subnet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CIDR of the subnet in which the public port IP address is allocated.",
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
				Description: "Enables (`true`) or disables (`false`) the public port administratively.",
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: 20,
				Description: "List of OpenStack security group identifiers to associate with the public port. " +
					"Learn more about the [openstack_networking_secgroup_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_secgroup_v2) resource in the official OpenStack documentation. " +
					"The default value is the identifier of the default security group in the project.",
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

	securityGroupIDsVal := d.Get("security_group_ids")
	if securityGroupIDsVal != nil {
		dto.SecurityGroupIDs = expandStringList(securityGroupIDsVal.([]any))
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
	config := meta.(*Config)

	if config.ProjectID == "" {
		return nil, fmt.Errorf("project_id must be set in your environment")
	}

	if config.Region == "" {
		return nil, fmt.Errorf("region must be set in your environment")
	}

	_ = d.Set("project_id", config.ProjectID)
	_ = d.Set("region", config.Region)

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
