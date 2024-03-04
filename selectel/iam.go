package selectel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go"
	"github.com/selectel/iam-go/service/roles"
	"github.com/selectel/iam-go/service/users"
)

const (
	importIAMFieldValueFailed = "IMPORT_FAILED"
)

func getIAMClient(meta interface{}) (*iam.Client, diag.Diagnostics) {
	config := meta.(*Config)

	selvcpclient, err := config.GetSelVPCClient()
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get selvpc client for iam: %w", err))
	}

	iamClient, err := iam.New(
		iam.WithAuthOpts(&iam.AuthOpts{
			KeystoneToken: selvcpclient.GetXAuthToken(),
		}),
	)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't create iam client: %w", err))
	}

	return iamClient, nil
}

func convertIAMMapToUserFederation(federationMap map[string]interface{}) (*users.Federation, error) {
	if len(federationMap) == 0 {
		return nil, nil
	}
	var idRaw, externalIDRaw interface{}
	var ok bool

	if idRaw, ok = federationMap["id"]; !ok {
		return nil, errors.New("federation.id value isn't provided")
	}
	if externalIDRaw, ok = federationMap["external_id"]; !ok {
		return nil, errors.New("federation.external_id value isn't provided")
	}

	id := idRaw.(string)
	externalID := externalIDRaw.(string)

	federation := &users.Federation{
		ExternalID: externalID,
		ID:         id,
	}

	return federation, nil
}

func convertIAMSetToRoles(rolesSet *schema.Set) ([]roles.Role, error) {
	rolesList := rolesSet.List()

	output := make([]roles.Role, len(rolesList))
	var roleNameRaw, scopeRaw, projectIDRaw interface{}
	var ok bool

	for i := range rolesList {
		var roleName, scope, projectID string
		resourceRoleMap := rolesList[i].(map[string]interface{})

		if roleNameRaw, ok = resourceRoleMap["role_name"]; !ok {
			return nil, errors.New("role_name value isn't provided")
		}
		if scopeRaw, ok = resourceRoleMap["scope"]; !ok {
			return nil, errors.New("scope value isn't provided")
		}
		if projectIDRaw, ok = resourceRoleMap["project_id"]; ok {
			projectID = projectIDRaw.(string)
		}

		roleName = roleNameRaw.(string)
		scope = scopeRaw.(string)

		output[i] = roles.Role{
			RoleName:  roles.Name(strings.ToLower(roleName)),
			Scope:     roles.Scope(strings.ToLower(scope)),
			ProjectID: projectID,
		}
	}

	return output, nil
}

func convertIAMRolesToSet(roles []roles.Role) []interface{} {
	result := make([]interface{}, 0)
	for _, role := range roles {
		result = append(result, map[string]interface{}{
			"role_name":  role.RoleName,
			"scope":      role.Scope,
			"project_id": role.ProjectID,
		})
	}

	return result
}

func flattenIAMUserFederation(federation *users.Federation) map[string]interface{} {
	if federation == nil {
		return nil
	}

	return map[string]interface{}{
		"id":          federation.ID,
		"external_id": federation.ExternalID,
	}
}
