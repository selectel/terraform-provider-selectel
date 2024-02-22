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

func getIAMServiceUserRolesFromList(roles []interface{}) []serviceusers.Role {
	if len(roles) == 0 {
		return nil
	}
	result := make([]serviceusers.Role, len(roles))
	for i := range roles {
		role := serviceusers.Role{}
		obj := roles[i].(map[string]interface{})
		role.RoleName = serviceusers.RoleName(strings.ToLower(obj["role_name"].(string)))
		role.Scope = serviceusers.Scope(strings.ToLower(obj["scope"].(string)))
		if v, ok := obj["project_id"]; ok {
			role.ProjectID = v.(string)
		}
		result[i] = role
	}

	return result
}

func getIAMUserRolesFromList(roles []interface{}) []users.Role {
	if len(roles) == 0 {
		return nil
	}
	result := make([]users.Role, len(roles))
	for i := range roles {
		role := users.Role{}
		obj := roles[i].(map[string]interface{})
		role.RoleName = users.RoleName(strings.ToLower(obj["role_name"].(string)))
		role.Scope = users.Scope(strings.ToLower(obj["scope"].(string)))
		if v, ok := obj["project_id"]; ok {
			role.ProjectID = v.(string)
		}
		result[i] = role
	}

	return result
}

func getIAMUserFederationFromSet(federationSet *schema.Set) (*users.Federation, error) {
	if federationSet.Len() == 0 {
		return nil, nil
	}
	var idRaw, externalIDRaw interface{}
	var ok bool

	resourceFederationMap := federationSet.List()[0].(map[string]interface{})
	if idRaw, ok = resourceFederationMap["id"]; !ok {
		return nil, errors.New("federation.id value isn't provided")
	}
	if externalIDRaw, ok = resourceFederationMap["external_id"]; !ok {
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

func containsServiceUsersRole(s []serviceusers.Role, e serviceusers.Role) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func containsUsersRole(s []users.Role, e users.Role) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
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

func flattenIAMUserFederation(federation *users.Federation) []interface{} {
	if federation == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"id":          federation.ID,
			"external_id": federation.ExternalID,
		},
	}
}
