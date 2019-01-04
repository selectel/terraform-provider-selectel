package keypairs

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "keypairs"

// List gets a list of keypairs in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient) ([]*Keypair, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract keypairs. from the response body.
	var result struct {
		Keypairs []*Keypair `json:"keypairs"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Keypairs, responseResult, nil
}

// Create requests a creation of the keypar with the specified options.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, createOpts KeypairOpts) ([]*Keypair, *selvpcclient.ResponseResult, error) {
	// Nest create opts into additional body.
	type nestedCreateOpts struct {
		Keypair KeypairOpts `json:"keypair"`
	}
	var createKeypairOpts = nestedCreateOpts{
		Keypair: createOpts,
	}
	requestBody, err := json.Marshal(&createKeypairOpts)
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

	// Extract a keypair from the response body.
	var result struct {
		Keypair []*Keypair `json:"keypair"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Keypair, responseResult, nil
}

// Delete deletes a single keypair by its name and user ID.
func Delete(ctx context.Context, client *selvpcclient.ServiceClient, name, userID string) (*selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, name, "users", userID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}
	return responseResult, err
}
