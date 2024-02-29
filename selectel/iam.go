package selectel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go"
	"github.com/selectel/iam-go/service/serviceusers"
	"github.com/selectel/iam-go/service/users"
)

const (
	unknownFieldValue = "UNKNOWN"
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

func getIAMUserFederationFromMap(federationMap map[string]interface{}) (*users.Federation, error) {
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

func getIAMUserRolesFromSet(rolesSet *schema.Set) ([]users.Role, error) {
	if rolesSet.Len() == 0 {
		return nil, nil
	}
	var roleNameRaw, scopeRaw, projectIDRaw interface{}
	var ok bool

	rolesList := rolesSet.List()

	roles := make([]users.Role, len(rolesList))

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

		roles[i] = users.Role{
			RoleName:  users.RoleName(strings.ToLower(roleName)),
			Scope:     users.Scope(strings.ToLower(scope)),
			ProjectID: projectID,
		}
	}

	return roles, nil
}

func getIAMServiceUserRolesFromSet(rolesSet *schema.Set) ([]serviceusers.Role, error) {
	if rolesSet.Len() == 0 {
		return nil, nil
	}
	var roleNameRaw, scopeRaw, projectIDRaw interface{}
	var ok bool

	rolesList := rolesSet.List()

	roles := make([]serviceusers.Role, len(rolesList))

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

		roles[i] = serviceusers.Role{
			RoleName:  serviceusers.RoleName(strings.ToLower(roleName)),
			Scope:     serviceusers.Scope(strings.ToLower(scope)),
			ProjectID: projectID,
		}
	}

	return roles, nil
}

func flattenIAMServiceUserRoles(roles []serviceusers.Role) []interface{} {
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

func flattenIAMUserRoles(roles []users.Role) []interface{} {
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
