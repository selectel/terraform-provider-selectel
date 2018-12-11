package roles

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "roles"

// ListProject returns all roles in the specified project.
func ListProject(ctx context.Context, client *selvpcclient.ServiceClient, id string) ([]*Role, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, "projects", id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract roles from the response body.
	var result struct {
		Roles []*Role `json:"roles"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Roles, responseResult, nil
}

// ListUser returns all roles that are associated with the specified user.
func ListUser(ctx context.Context, client *selvpcclient.ServiceClient, id string) ([]*Role, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, "users", id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract roles from the response body.
	var result struct {
		Roles []*Role `json:"roles"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Roles, responseResult, nil
}

// Create requests a creation of the single role for the specified project and user.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, createOpts RoleOpt) (*Role, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, "projects", createOpts.ProjectID, "users", createOpts.UserID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract role from the response body.
	var result struct {
		Role *Role `json:"role"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Role, responseResult, nil
}

// CreateBulk requests a creation of several roles.
func CreateBulk(ctx context.Context, client *selvpcclient.ServiceClient, createOpts RoleOpts) ([]*Role, *selvpcclient.ResponseResult, error) {
	createRolesOpts := &createOpts
	requestBody, err := json.Marshal(createRolesOpts)
	if err != nil {
		return nil, nil, err
	}

	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract role from the response body.
	var result struct {
		Roles []*Role `json:"roles"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Roles, responseResult, nil
}

// Delete requests a deletion of the single role for the specified project and user.
func Delete(ctx context.Context, client *selvpcclient.ServiceClient, deleteOpts RoleOpt) (*selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, "projects", deleteOpts.ProjectID, "users", deleteOpts.UserID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		return responseResult, responseResult.Err
	}
	return responseResult, nil
}
