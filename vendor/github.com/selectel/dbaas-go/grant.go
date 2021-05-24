package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GrantCreateOpts represents options for the grant Create request.
type GrantCreateOpts struct {
	DatastoreID string `json:"datastore_id"`
	DatabaseID  string `json:"database_id"`
	UserID      string `json:"user_id"`
}

// Grant is the API response for the grants.
type Grant struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	ProjectID   string `json:"project_id"`
	DatastoreID string `json:"datastore_id"`
	DatabaseID  string `json:"database_id"`
	UserID      string `json:"user_id"`
	Status      Status `json:"status"`
}

// Grant returns a grant based on the ID.
func (api *API) Grant(ctx context.Context, grantID string) (Grant, error) {
	uri := fmt.Sprintf("/grants/%s", grantID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Grant{}, err
	}

	var result struct {
		Grant Grant `json:"grant"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Grant{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Grant, nil
}

// Grants returns all grants.
func (api *API) Grants(ctx context.Context) ([]Grant, error) {
	uri := "/grants"

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Grant{}, err
	}

	var result struct {
		Grants []Grant `json:"grants"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Grant{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Grants, nil
}

// CreateGrant creates a new grant.
func (api *API) CreateGrant(ctx context.Context, opts GrantCreateOpts) (Grant, error) {
	uri := "/grants"
	createGrantOpts := struct {
		Grant GrantCreateOpts `json:"grant"`
	}{
		Grant: opts,
	}
	requestBody, err := json.Marshal(createGrantOpts)
	if err != nil {
		return Grant{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPost, uri, requestBody)
	if err != nil {
		return Grant{}, err
	}

	var result struct {
		Grant Grant `json:"grant"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Grant{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Grant, nil
}

// DeleteGrant deletes an existing grant.
func (api *API) DeleteGrant(ctx context.Context, grantID string) error {
	uri := fmt.Sprintf("/grants/%s", grantID)

	_, err := api.makeRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}
