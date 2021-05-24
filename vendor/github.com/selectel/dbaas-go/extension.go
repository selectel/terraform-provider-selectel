package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Extension is the API response for the extension.
type Extension struct {
	ID                   string `json:"id"`
	AvailableExtensionID string `json:"available_extension_id"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
	DatastoreID          string `json:"datastore_id"`
	DatabaseID           string `json:"database_id"`
	Status               Status `json:"status"`
}

// ExtensionCreateOpts represents options for the extension Create request.
type ExtensionCreateOpts struct {
	AvailableExtensionID string `json:"available_extension_id"`
	DatastoreID          string `json:"datastore_id"`
	DatabaseID           string `json:"database_id"`
}

// ExtensionQueryParams represents available query parameters for extension.
type ExtensionQueryParams struct {
	ID                   string `json:"id,omitempty"`
	AvailableExtensionID string `json:"available_extension_id,omitempty"`
	DatastoreID          string `json:"datastore_id,omitempty"`
	DatabaseID           string `json:"database_id,omitempty"`
	Status               Status `json:"status,omitempty"`
}

// Extensions returns all extensions.
func (api *API) Extensions(ctx context.Context, params *ExtensionQueryParams) ([]Extension, error) {
	uri, err := setQueryParams("/extensions", params)
	if err != nil {
		return []Extension{}, err
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Extension{}, err
	}

	var result struct {
		Extensions []Extension `json:"extensions"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Extension{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Extensions, nil
}

// Extension returns a extension based on the ID.
func (api *API) Extension(ctx context.Context, extensionID string) (Extension, error) {
	uri := fmt.Sprintf("/extensions/%s", extensionID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Extension{}, err
	}

	var result struct {
		Extension Extension `json:"extension"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Extension{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Extension, nil
}

// CreateExtension creates a new extension.
func (api *API) CreateExtension(ctx context.Context, opts ExtensionCreateOpts) (Extension, error) {
	uri := "/extensions"
	createExtensionOpts := struct {
		Extension ExtensionCreateOpts `json:"extension"`
	}{
		Extension: opts,
	}
	requestBody, err := json.Marshal(createExtensionOpts)
	if err != nil {
		return Extension{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPost, uri, requestBody)
	if err != nil {
		return Extension{}, err
	}

	var result struct {
		Extension Extension `json:"extension"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Extension{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Extension, nil
}

// DeleteExtension deletes an existing extension.
func (api *API) DeleteExtension(ctx context.Context, extensionID string) error {
	uri := fmt.Sprintf("/extensions/%s", extensionID)

	_, err := api.makeRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}
