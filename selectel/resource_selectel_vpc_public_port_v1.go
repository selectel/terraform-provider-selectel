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
		Importer:      &schema.ResourceImporter{StateContext: resourceVPCPublicPortV1ImportState},
		Schema: map[string]*schema.Schema{
			"region":         {Type: schema.TypeString, Required: true, ForceNew: true},
			"project_id":     {Type: schema.TypeString, Required: true, ForceNew: true},
			"network_id":     {Type: schema.TypeString, Computed: true},
			"ip_address":     {Type: schema.TypeString, Computed: true},
			"description":    {Type: schema.TypeString, Optional: true},
			"subnet":         {Type: schema.TypeString, Computed: true},
			"gateway":        {Type: schema.TypeString, Computed: true},
			"admin_state_up": {Type: schema.TypeBool, Optional: true, Computed: true},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				MaxItems: 20,
			},
		},
	}
}

func resourceVPCPublicPortV1Create(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client, diagErr := getPublicNetAPIClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	dto := &publicnetapi.PortCreateDTO{ProjectID: d.Get("project_id").(string)}

	if descriptionVal, ok := d.GetOk("description"); ok {
		description := descriptionVal.(string)
		dto.Description = &description
	}

	if adminStateUpVal, ok := d.GetOk("admin_state_up"); ok {
		adminStateUp := adminStateUpVal.(bool)
		dto.AdminStateUp = &adminStateUp
	}

	if securityGroupIDsVal, ok := d.GetOk("security_group_ids"); ok {
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

func resourceVPCPublicPortV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceVPCPublicPortV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		dto.SecurityGroupIDs = expandStringList(d.Get("security_group_ids").([]any))
	}

	log.Print(msgUpdate(objectPublicPort, d.Id(), dto))

	port, err := client.UpdatePort(ctx, d.Id(), &dto)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectPublicPort, d.Id(), err))
	}

	fillVPCPublicPortData(port, d)

	return nil
}

func resourceVPCPublicPortV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	meta interface{},
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
