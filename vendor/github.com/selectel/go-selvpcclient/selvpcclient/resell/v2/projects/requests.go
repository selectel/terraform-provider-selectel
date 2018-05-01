package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "projects"

// Get returns a single project by its id.
func Get(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*Project, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a project from the response body.
	var result struct {
		Project *Project `json:"project"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Project, responseResult, nil
}

// List gets a list of projects in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient) ([]*Project, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")
	responseResult, err := client.DoRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract projects from the response body.
	var result struct {
		Projects []*Project `json:"projects"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Projects, responseResult, nil
}

// Create requests a creation of the project.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, createOpts CreateOpts) (*Project, *selvpcclient.ResponseResult, error) {
	// Nest create options into the parent "project" JSON structure.
	type createProject struct {
		Options CreateOpts `json:"project"`
	}
	createProjectOpts := &createProject{Options: createOpts}
	requestBody, err := json.Marshal(createProjectOpts)
	if err != nil {
		return nil, nil, err
	}

	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")
	responseResult, err := client.DoRequest(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a project from the response body.
	var result struct {
		Project *Project `json:"project"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Project, responseResult, nil
}

// Update requests an update of the project referenced by its id.
func Update(ctx context.Context, client *selvpcclient.ServiceClient, id string, updateOpts UpdateOpts) (*Project, *selvpcclient.ResponseResult, error) {
	// Nest update options into the parent "project" JSON structure.
	type updateProject struct {
		Options UpdateOpts `json:"project"`
	}
	updateProjectOpts := &updateProject{Options: updateOpts}
	requestBody, err := json.Marshal(updateProjectOpts)
	if err != nil {
		return nil, nil, err
	}

	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, "PATCH", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a project from the response body.
	var result struct {
		Project *Project `json:"project"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Project, responseResult, nil
}

// Delete deletes a single project by its id.
func Delete(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}
	return responseResult, err
}
