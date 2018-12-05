package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "users"

// List gets a list of users in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient) ([]*User, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract users from the response body.
	var result struct {
		Users []*User `json:"users"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Users, responseResult, nil
}

// Create requests a creation of the user.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, createOpts UserOpts) (*User, *selvpcclient.ResponseResult, error) {
	// Nest create options into the parent "user" JSON structure.
	type createUser struct {
		Options UserOpts `json:"user"`
	}
	createUserOpts := &createUser{Options: createOpts}
	requestBody, err := json.Marshal(createUserOpts)
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

	// Extract a user from the response body.
	var result struct {
		User *User `json:"user"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.User, responseResult, nil
}

// Update requests an update of the user referenced by its id.
func Update(ctx context.Context, client *selvpcclient.ServiceClient, id string, updateOpts UserOpts) (*User, *selvpcclient.ResponseResult, error) {
	// Nest update options into the parent "user" JSON structure.
	type updateUser struct {
		Options UserOpts `json:"user"`
	}
	updateUserOpts := &updateUser{Options: updateOpts}
	requestBody, err := json.Marshal(updateUserOpts)
	if err != nil {
		return nil, nil, err
	}

	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPatch, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract an user from the response body.
	var result struct {
		User *User `json:"user"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.User, responseResult, nil
}

// Delete deletes a single user by its id.
func Delete(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}
	return responseResult, err
}
