package selectel

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
	"github.com/selectel/iam-go"
	"github.com/selectel/iam-go/service/roles"
	"github.com/selectel/iam-go/service/users"
)

const (
	importIAMUndefined = "UNDEFINED_WHILE_IMPORTING"
)

func getIAMClient(meta interface{}) (*iam.Client, diag.Diagnostics) {
	config := meta.(*Config)

	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get selvpc client for iam: %w", err))
	}
	// This is for future implementation of Keystone Catalog
	_, err = getEndpointForIAM(selvpcClient, config.AuthRegion)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	iamClient, err := iam.New(
		iam.WithAuthOpts(&iam.AuthOpts{
			KeystoneToken: selvpcClient.GetXAuthToken(),
		}),
		// This is for future implementation of Keystone Catalog
		// iam.WithAPIUrl(endpoint.URL),
	)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't create iam client: %w", err))
	}

	return iamClient, nil
}

func getEndpointForIAM(selvpcClient *selvpcclient.Client, region string) (string, error) {
	endpoint, err := selvpcClient.Catalog.GetEndpoint(IAM, region)
	if err != nil {
		return "", fmt.Errorf("can't get endpoint to for iam: %w", err)
	}

	return endpoint.URL, nil
}

func manageRoles(oldRoles, newRoles []roles.Role) ([]roles.Role, []roles.Role) {
	rolesToUnassign := make([]roles.Role, 0)
	rolesToAssign := make([]roles.Role, 0)

	for _, oldRole := range oldRoles {
		if !slices.Contains(newRoles, oldRole) {
			rolesToUnassign = append(rolesToUnassign, oldRole)
		}
	}

	for _, newRole := range newRoles {
		if !slices.Contains(oldRoles, newRole) {
			rolesToAssign = append(rolesToAssign, newRole)
		}
	}

	return rolesToUnassign, rolesToAssign
}

func applyServiceUserRoles(ctx context.Context, d *schema.ResourceData, iamClient *iam.Client, rolesToUnassign, rolesToAssign []roles.Role) error {
	if len(rolesToAssign) != 0 {
		err := iamClient.ServiceUsers.AssignRoles(ctx, d.Id(), rolesToAssign)
		if err != nil {
			return err
		}
	}

	if len(rolesToUnassign) != 0 {
		err := iamClient.ServiceUsers.UnassignRoles(ctx, d.Id(), rolesToUnassign)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyUserRoles(ctx context.Context, d *schema.ResourceData, iamClient *iam.Client, rolesToUnassign, rolesToAssign []roles.Role) error {
	if len(rolesToAssign) != 0 {
		err := iamClient.Users.AssignRoles(ctx, d.Id(), rolesToAssign)
		if err != nil {
			return err
		}
	}

	if len(rolesToUnassign) != 0 {
		err := iamClient.Users.UnassignRoles(ctx, d.Id(), rolesToUnassign)
		if err != nil {
			return err
		}
	}

	return nil
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
			RoleName:  roles.Name(roleName),
			Scope:     roles.Scope(scope),
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

func convertIAMFederationToMap(federation *users.Federation) map[string]interface{} {
	if federation == nil {
		return nil
	}

	return map[string]interface{}{
		"id":          federation.ID,
		"external_id": federation.ExternalID,
	}
}
