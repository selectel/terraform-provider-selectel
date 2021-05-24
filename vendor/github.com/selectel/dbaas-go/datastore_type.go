package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// DatastoreType is the API response for the datastore types.
type DatastoreType struct {
	ID      string `json:"id"`
	Engine  string `json:"engine"`
	Version string `json:"version"`
}

// DatastoreTypes returns all datastore types.
func (api *API) DatastoreTypes(ctx context.Context) ([]DatastoreType, error) {
	uri := "/datastore-types"

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []DatastoreType{}, err
	}

	var result struct {
		DatastoreTypes []DatastoreType `json:"datastore-types"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []DatastoreType{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.DatastoreTypes, nil
}

// DatastoreType returns a datastore type based on the ID.
func (api *API) DatastoreType(ctx context.Context, datastoreTypeID string) (DatastoreType, error) {
	uri := fmt.Sprintf("/datastore-types/%s", datastoreTypeID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return DatastoreType{}, err
	}

	var result struct {
		DatastoreType DatastoreType `json:"datastore-type"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return DatastoreType{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.DatastoreType, nil
}
