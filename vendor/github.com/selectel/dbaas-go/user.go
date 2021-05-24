package dbaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UserCreateOpts represents options for the user Create request.
type UserCreateOpts struct {
	Name        string `json:"name"`
	Password    string `json:"password"`
	DatastoreID string `json:"datastore_id"`
}

// UserUpdateOpts represents options for the user Update request.
type UserUpdateOpts struct {
	Password string `json:"password"`
}

// User is the API response for the users.
type User struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	ProjectID   string `json:"project_id"`
	DatastoreID string `json:"datastore_id"`
	Name        string `json:"name"`
	Status      Status `json:"status"`
}

// User returns a user based on the ID.
func (api *API) User(ctx context.Context, userID string) (User, error) {
	uri := fmt.Sprintf("/users/%s", userID)

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return User{}, err
	}

	var result struct {
		User User `json:"user"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return User{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.User, nil
}

// Users returns all users.
func (api *API) Users(ctx context.Context) ([]User, error) {
	uri := "/users"

	resp, err := api.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []User{}, err
	}

	var result struct {
		Users []User `json:"users"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return []User{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.Users, nil
}

// CreateUser creates a new user.
func (api *API) CreateUser(ctx context.Context, opts UserCreateOpts) (User, error) {
	uri := "/users"
	createUserOpts := struct {
		User UserCreateOpts `json:"user"`
	}{
		User: opts,
	}
	requestBody, err := json.Marshal(createUserOpts)
	if err != nil {
		return User{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPost, uri, requestBody)
	if err != nil {
		return User{}, err
	}

	var result struct {
		User User `json:"user"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return User{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.User, nil
}

// DeleteUser deletes an existing user.
func (api *API) DeleteUser(ctx context.Context, userID string) error {
	uri := fmt.Sprintf("/users/%s", userID)

	_, err := api.makeRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser updates an existing user.
func (api *API) UpdateUser(ctx context.Context, userID string, opts UserUpdateOpts) (User, error) {
	uri := fmt.Sprintf("/users/%s", userID)
	updateUserOpts := struct {
		User UserUpdateOpts `json:"user"`
	}{
		User: opts,
	}
	requestBody, err := json.Marshal(updateUserOpts)
	if err != nil {
		return User{}, fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	resp, err := api.makeRequest(ctx, http.MethodPut, uri, requestBody)
	if err != nil {
		return User{}, err
	}

	var result struct {
		User User `json:"user"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return User{}, fmt.Errorf("Error during Unmarshal, %w", err)
	}

	return result.User, nil
}
