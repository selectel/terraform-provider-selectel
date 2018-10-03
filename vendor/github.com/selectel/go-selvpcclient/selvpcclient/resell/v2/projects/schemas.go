package projects

import (
	"encoding/json"

	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/quotas"
)

// Project represents a single Identity service project.
type Project struct {
	// ID is a unique id of a project.
	ID string `json:"-"`

	// Name is a human-readable name of a project.
	Name string `json:"-"`

	// URL is a public url of a project that is set by the admin API.
	URL string `json:"-"`

	// Enabled shows if project is active or it was disabled by the admin API.
	Enabled bool `json:"-"`

	// CustomURL is a public url of a project that can be set by a user.
	CustomURL string `json:"-"`

	// Theme represents project theme settings.
	Theme Theme `json:"-"`

	// Quotas contains information about project quotas.
	Quotas []quotas.Quota `json:"-"`
}

// UnmarshalJSON implements custom unmarshalling method for the Project type.
func (result *Project) UnmarshalJSON(b []byte) error {
	// Populate temporary structure with resource quotas represented as maps.
	var s struct {
		ID        string                                  `json:"id"`
		Name      string                                  `json:"name"`
		URL       string                                  `json:"url"`
		Enabled   bool                                    `json:"enabled"`
		CustomURL string                                  `json:"custom_url"`
		Theme     Theme                                   `json:"theme"`
		Quotas    map[string][]quotas.ResourceQuotaEntity `json:"quotas"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	// Populate the result with the unmarshalled data.
	*result = Project{
		ID:        s.ID,
		Name:      s.Name,
		URL:       s.URL,
		Enabled:   s.Enabled,
		CustomURL: s.CustomURL,
		Theme:     s.Theme,
	}

	if len(s.Quotas) != 0 {
		// Convert resource quota maps to the slice of Quota types.
		// Here we're allocating memory in advance because we already know the length
		// of a result slice from the JSON bytearray.
		resourceQuotasSlice := make([]quotas.Quota, len(s.Quotas))
		i := 0
		for resourceName, resourceQuotas := range s.Quotas {
			resourceQuotasSlice[i] = quotas.Quota{
				Name:                   resourceName,
				ResourceQuotasEntities: resourceQuotas,
			}
			i++
		}

		// Add the unmarshalled quotas slice to the result.
		result.Quotas = resourceQuotasSlice
	}

	return nil
}

// Theme represents theme settings for a single project.
type Theme struct {
	// Color is a hex string with a custom background color.
	Color string `json:"color"`

	// Logo contains url for the project custom header logotype.
	Logo string `json:"logo"`
}
