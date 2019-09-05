package tokens

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "tokens"

// Create requests a creation of the Identity token.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, createOpts TokenOpts) (*Token, *selvpcclient.ResponseResult, error) {
	// Nest create options into the parent "token" JSON structure.
	type createToken struct {
		Options TokenOpts `json:"token"`
	}
	createTokenOpts := &createToken{Options: createOpts}
	requestBody, err := json.Marshal(createTokenOpts)
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

	// Extract a token from the response body.
	var result struct {
		Token *Token `json:"token"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Token, responseResult, nil
}

// Delete a user owned Identity token by its id.
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
