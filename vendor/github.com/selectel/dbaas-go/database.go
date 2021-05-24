package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Database is the API response for the databases.
type Database struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	LcCollate   string `json:"lc_collate"`
	LcCtype     string `json:"lc_ctype"`
	DatastoreID string `json:"datastore_id"`
	Status      Status `json:"status"`
}

// DatabaseCreateOpts represents options for the database Create request.
type DatabaseCreateOpts struct {
	DatastoreID string `json:"datastore_id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id,omitempty"`
	LcCollate   string `json:"lc_collate,omitempty"`
	LcCtype     string `json:"lc_ctype,omitempty"`
}

// DatabaseUpdateOpts represents options for the database Update request.
type DatabaseUpdateOpts struct {
	OwnerID string `json:"owner_id"`
}

// DatabaseQueryParams represents available query parameters for database.
type DatabaseQueryParams struct {
	ID          string `json:"id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Name        string `json:"name,omitempty"`
	OwnerID     string `json:"owner_id,omitempty"`
	LcCollate   string `json:"lc_collate,omitempty"`
	LcCtype     string `json:"lc_ctype,omitempty"`
	DatastoreID string `json:"datastore_id,omitempty"`
	Status      Status `json:"status,omitempty"`
}

// Databases returns all databases.
func (api *API) Databases(ctx context.Context, params *DatabaseQueryParams) ([]Database, error) {
	uri, err := setQueryParams("/databases", params)
	if err != nil {
		return []Database{}, err
	}

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Database{}, err
	}

	var result struct {
		Databases []Database `json:"databases"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []Database{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Databases, nil
}

// Database returns a database based on the ID.
func (api *API) Database(ctx context.Context, databaseID string) (Database, error) {
	uri := fmt.Sprintf("/databases/%s", databaseID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Database{}, err
	}

	var result struct {
		Database Database `json:"database"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Database{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Database, nil
}

// CreateDatabase creates a new database.
func (api *API) CreateDatabase(ctx context.Context, opts DatabaseCreateOpts) (Database, error) {
	uri := "/databases"
	createDatabaseOpts := struct {
		Database DatabaseCreateOpts `json:"database"`
	}{
		Database: opts,
	}
	requestBody, err := json.Marshal(createDatabaseOpts)
	if err != nil {
		return Database{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPost, uri, requestBody)
	if err != nil {
		return Database{}, err
	}

	var result struct {
		Database Database `json:"database"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Database{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Database, nil
}

// UpdateDatabase updates an existing database.
func (api *API) UpdateDatabase(ctx context.Context, databaseID string, opts DatabaseUpdateOpts) (Database, error) {
	uri := fmt.Sprintf("/databases/%s", databaseID)
	updateDatabaseOpts := struct {
		Database DatabaseUpdateOpts `json:"database"`
	}{
		Database: opts,
	}
	requestBody, err := json.Marshal(updateDatabaseOpts)
	if err != nil {
		return Database{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPut, uri, requestBody)
	if err != nil {
		return Database{}, err
	}

	var result struct {
		Database Database `json:"database"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Database{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Database, nil
}

// DeleteDatabase deletes an existing database.
func (api *API) DeleteDatabase(ctx context.Context, databaseID string) error {
	uri := fmt.Sprintf("/databases/%s", databaseID)

	_, err := api.makeRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}
