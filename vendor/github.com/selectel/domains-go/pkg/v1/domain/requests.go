package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	v1 "github.com/selectel/domains-go/pkg/v1"
)

// GetByID returns a single domain by its id.
func GetByID(ctx context.Context, client *v1.ServiceClient, domainID int) (*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, strconv.Itoa(domainID)}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a domain from the response body.
	domain := &View{}
	err = responseResult.ExtractResult(domain)
	if err != nil {
		return nil, responseResult, err
	}

	return domain, responseResult, nil
}

// GetByName returns a single domain by its domain name.
func GetByName(ctx context.Context, client *v1.ServiceClient, domainName string) (*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, domainName}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a domain from the response body.
	domain := &View{}
	err = responseResult.ExtractResult(domain)
	if err != nil {
		return nil, responseResult, err
	}

	return domain, responseResult, nil
}

// List gets a list of all domains.
func List(ctx context.Context, client *v1.ServiceClient) ([]*View, *v1.ResponseResult, error) {
	url := client.Endpoint + "/"
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract domains from the response body.
	var domains []*View
	err = responseResult.ExtractResult(&domains)
	if err != nil {
		return nil, responseResult, err
	}

	return domains, responseResult, nil
}

// Create requests a creation of a new domain.
func Create(ctx context.Context, client *v1.ServiceClient, opts *CreateOpts) (*View, *v1.ResponseResult, error) {
	requestBody, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}

	url := client.Endpoint + "/"
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	if opts.BindZone != "" {
		result := struct {
			Domain *View `json:"domain"`
		}{}

		// Extract domain from the response body.
		err = responseResult.ExtractResult(&result)
		if err != nil {
			return nil, responseResult, err
		}
		return result.Domain, responseResult, nil
	}

	// Extract domain from the response body.
	domain := &View{}
	err = responseResult.ExtractResult(domain)
	if err != nil {
		return nil, responseResult, err
	}

	return domain, responseResult, nil
}

// Delete deletes a single domain by its id.
func Delete(ctx context.Context, client *v1.ServiceClient, domainID int) (*v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, strconv.Itoa(domainID)}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}
