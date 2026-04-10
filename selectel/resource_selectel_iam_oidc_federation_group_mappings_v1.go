package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/federations/oidc/groupmappings"
)

func resourceIAMOIDCFederationGroupMappingsV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents OIDC Federation group mappings in IAM API",
		CreateContext: resourceIAMOIDCFederationGroupMappingsV1Create,
		ReadContext:   resourceIAMOIDCFederationGroupMappingsV1Read,
		UpdateContext: resourceIAMOIDCFederationGroupMappingsV1Update,
		DeleteContext: resourceIAMOIDCFederationGroupMappingsV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Federation ID to manage group mappings for.",
			},
			"group_mapping": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of mappings between internal and external groups.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"internal_group_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Internal IAM group ID.",
						},
						"external_group_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "External identity provider group ID.",
						},
					},
				},
			},
		},
	}
}

func resourceIAMOIDCFederationGroupMappingsV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Get("federation_id").(string)
	request := expandOIDCFederationGroupMappings(d)

	log.Print(msgCreate(objectOIDCFederationGroupMappings, fmt.Sprintf("federation_id: %s, group_mappings: %+v", federationID, request.GroupMappings)))

	err := iamClient.OIDCFederations.GroupMappings.Update(ctx, federationID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectOIDCFederationGroupMappings, err))
	}

	d.SetId(federationID)

	return resourceIAMOIDCFederationGroupMappingsV1Read(ctx, d, meta)
}

func resourceIAMOIDCFederationGroupMappingsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Id()
	if federationID == "" {
		if v, ok := d.GetOk("federation_id"); ok {
			federationID = v.(string)
		}
	}

	log.Print(msgGet(objectOIDCFederationGroupMappings, federationID))

	mappings, err := iamClient.OIDCFederations.GroupMappings.List(ctx, federationID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrFederationNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errGettingObject(objectOIDCFederationGroupMappings, federationID, err))
	}

	d.Set("federation_id", federationID)
	d.Set("group_mapping", flattenOIDCFederationGroupMappings(mappings.GroupMappings))

	return nil
}

func resourceIAMOIDCFederationGroupMappingsV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Get("federation_id").(string)
	if federationID == "" {
		federationID = d.Id()
	}

	request := expandOIDCFederationGroupMappings(d)

	log.Print(msgUpdate(objectOIDCFederationGroupMappings, federationID, fmt.Sprintf("group_mappings: %+v", request.GroupMappings)))

	err := iamClient.OIDCFederations.GroupMappings.Update(ctx, federationID, request)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectOIDCFederationGroupMappings, federationID, err))
	}

	d.SetId(federationID)

	return resourceIAMOIDCFederationGroupMappingsV1Read(ctx, d, meta)
}

func resourceIAMOIDCFederationGroupMappingsV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Id()
	if federationID == "" {
		if v, ok := d.GetOk("federation_id"); ok {
			federationID = v.(string)
		}
	}

	log.Print(msgDelete(objectOIDCFederationGroupMappings, federationID))

	emptyRequest := groupmappings.GroupMappingsRequest{
		GroupMappings: []groupmappings.GroupMapping{},
	}

	err := iamClient.OIDCFederations.GroupMappings.Update(ctx, federationID, emptyRequest)
	if err != nil && !errors.Is(err, iamerrors.ErrFederationNotFound) {
		return diag.FromErr(errDeletingObject(objectOIDCFederationGroupMappings, federationID, err))
	}

	d.SetId("")

	return nil
}

func expandOIDCFederationGroupMappings(d *schema.ResourceData) groupmappings.GroupMappingsRequest {
	rawMappings := d.Get("group_mapping").([]interface{})

	mappings := make([]groupmappings.GroupMapping, 0, len(rawMappings))

	for _, raw := range rawMappings {
		mappingData := raw.(map[string]interface{})

		mappings = append(mappings, groupmappings.GroupMapping{
			InternalGroupID: mappingData["internal_group_id"].(string),
			ExternalGroupID: mappingData["external_group_id"].(string),
		})
	}

	return groupmappings.GroupMappingsRequest{
		GroupMappings: mappings,
	}
}

func flattenOIDCFederationGroupMappings(mappings []groupmappings.GroupMapping) []interface{} {
	result := make([]interface{}, 0, len(mappings))

	for _, mapping := range mappings {
		result = append(result, map[string]interface{}{
			"internal_group_id": mapping.InternalGroupID,
			"external_group_id": mapping.ExternalGroupID,
		})
	}

	return result
}
