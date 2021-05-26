package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// FlavorResponse is the API response for the flavors.
type FlavorResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Vcpus       int    `json:"vcpus"`
	RAM         int    `json:"ram"`
	Disk        int    `json:"disk"`
}

// Flavors returns all flavors.
func (api *API) Flavors(ctx context.Context) ([]FlavorResponse, error) {
	uri := "/flavors"

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []FlavorResponse{}, err
	}

	var result struct {
		Flavors []FlavorResponse `json:"flavors"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []FlavorResponse{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Flavors, nil
}

// Flavor returns a flavor based on the ID.
func (api *API) Flavor(ctx context.Context, flavorID string) (FlavorResponse, error) {
	uri := fmt.Sprintf("/flavors/%s", flavorID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return FlavorResponse{}, err
	}

	var result struct {
		Flavor FlavorResponse `json:"flavor"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return FlavorResponse{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Flavor, nil
}
