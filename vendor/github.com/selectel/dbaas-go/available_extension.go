package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AvailableExtension is the API response for the available extensions.
type AvailableExtension struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	DatastoreTypeIDs []string `json:"datastore_type_ids"`
	DependencyIDs    []string `json:"dependency_ids"`
}

// AvailableExtensions returns all available extensions.
func (api *API) AvailableExtensions(ctx context.Context) ([]AvailableExtension, error) {
	uri := "/available-extensions"

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []AvailableExtension{}, err
	}

	var result struct {
		AvailableExtensions []AvailableExtension `json:"available-extensions"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []AvailableExtension{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.AvailableExtensions, nil
}

// AvailableExtension returns an available extension based on the ID.
func (api *API) AvailableExtension(ctx context.Context, availableExtensionID string) (AvailableExtension, error) {
	uri := fmt.Sprintf("/available-extensions/%s", availableExtensionID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AvailableExtension{}, err
	}

	var result struct {
		AvailableExtension AvailableExtension `json:"available-extension"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return AvailableExtension{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.AvailableExtension, nil
}
