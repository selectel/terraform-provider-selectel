package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/federations/saml/groupmappings"
)

func resourceIAMSAMLFederationGroupMappingsV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents SAML Federation group mappings in IAM API",
		CreateContext: resourceIAMSAMLFederationGroupMappingsV1Create,
		ReadContext:   resourceIAMSAMLFederationGroupMappingsV1Read,
		UpdateContext: resourceIAMSAMLFederationGroupMappingsV1Update,
		DeleteContext: resourceIAMSAMLFederationGroupMappingsV1Delete,
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

func resourceIAMSAMLFederationGroupMappingsV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Get("federation_id").(string)
	request := expandSAMLFederationGroupMappings(d)

	log.Print(msgCreate(objectSAMLFederationGroupMappings, fmt.Sprintf("federation_id: %s, group_mappings: %+v", federationID, request.GroupMappings)))

	err := iamClient.SAMLFederations.GroupMappings.Update(ctx, federationID, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectSAMLFederationGroupMappings, err))
	}

	d.SetId(federationID)

	return resourceIAMSAMLFederationGroupMappingsV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationGroupMappingsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	log.Print(msgGet(objectSAMLFederationGroupMappings, federationID))

	mappings, err := iamClient.SAMLFederations.GroupMappings.List(ctx, federationID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrFederationNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errGettingObject(objectSAMLFederationGroupMappings, federationID, err))
	}

	d.Set("federation_id", federationID)
	d.Set("group_mapping", flattenSAMLFederationGroupMappings(mappings.GroupMappings))

	return nil
}

func resourceIAMSAMLFederationGroupMappingsV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	federationID := d.Get("federation_id").(string)
	if federationID == "" {
		federationID = d.Id()
	}

	request := expandSAMLFederationGroupMappings(d)

	log.Print(msgUpdate(objectSAMLFederationGroupMappings, federationID, fmt.Sprintf("group_mappings: %+v", request.GroupMappings)))

	err := iamClient.SAMLFederations.GroupMappings.Update(ctx, federationID, request)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectSAMLFederationGroupMappings, federationID, err))
	}

	d.SetId(federationID)

	return resourceIAMSAMLFederationGroupMappingsV1Read(ctx, d, meta)
}

func resourceIAMSAMLFederationGroupMappingsV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	log.Print(msgDelete(objectSAMLFederationGroupMappings, federationID))

	emptyRequest := groupmappings.GroupMappingsRequest{
		GroupMappings: []groupmappings.GroupMapping{},
	}

	err := iamClient.SAMLFederations.GroupMappings.Update(ctx, federationID, emptyRequest)
	if err != nil && !errors.Is(err, iamerrors.ErrFederationNotFound) {
		return diag.FromErr(errDeletingObject(objectSAMLFederationGroupMappings, federationID, err))
	}

	d.SetId("")

	return nil
}

func expandSAMLFederationGroupMappings(d *schema.ResourceData) groupmappings.GroupMappingsRequest {
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

func flattenSAMLFederationGroupMappings(mappings []groupmappings.GroupMapping) []interface{} {
	result := make([]interface{}, 0, len(mappings))

	for _, mapping := range mappings {
		result = append(result, map[string]interface{}{
			"internal_group_id": mapping.InternalGroupID,
			"external_group_id": mapping.ExternalGroupID,
		})
	}

	return result
}
