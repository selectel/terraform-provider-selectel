package projects

import (
	"encoding/json"

	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/quotas"
)

// CreateOpts represents options for the project Create request.
type CreateOpts struct {
	// Name sets the name for a new project.
	Name string `json:"-"`

	// Quotas sets quotas for a new project.
	Quotas []quotas.QuotaOpts `json:"-"`

	// AutoQuotas allows to automatically set quotas for the new project.
	// Quota values will be calculated in the Resell V2 service.
	AutoQuotas bool `json:"-"`
}

// MarshalJSON implements custom marshalling method for the CreateOpts.
func (opts *CreateOpts) MarshalJSON() ([]byte, error) {
	// Return create options with only name and auto_quotas parameters if quotas
	// parameter hadn't been provided.
	if len(opts.Quotas) == 0 {
		return json.Marshal(&struct {
			Name       string `json:"name"`
			AutoQuotas bool   `json:"auto_quotas"`
		}{
			Name:       opts.Name,
			AutoQuotas: opts.AutoQuotas,
		})
	}

	// Convert opts's quotas update options slice to a map that has resource
	// names as keys and resource quotas update options as values.
	quotasMap := make(map[string][]quotas.ResourceQuotaOpts, len(opts.Quotas))
	for _, quota := range opts.Quotas {
		quotasMap[quota.Name] = quota.ResourceQuotasOpts
	}

	return json.Marshal(&struct {
		Name       string                                `json:"name"`
		AutoQuotas bool                                  `json:"auto_quotas"`
		Quotas     map[string][]quotas.ResourceQuotaOpts `json:"quotas"`
	}{
		Name:       opts.Name,
		AutoQuotas: opts.AutoQuotas,
		Quotas:     quotasMap,
	})
}

// UpdateOpts represents options for the project Update request.
type UpdateOpts struct {
	// Name represents the name of a project.
	Name string `json:"name,omitempty"`

	// CustomURL is a public url of a project that can be set by a user.
	CustomURL *string `json:"custom_url,omitempty"`

	// Theme represents project theme settings.
	Theme *ThemeUpdateOpts `json:"theme,omitempty"`
}

// ThemeUpdateOpts represents project theme options for the Update request.
type ThemeUpdateOpts struct {
	// Color is a hex string with a custom background color.
	Color *string `json:"color,omitempty"`

	// Logo contains url for the project custom header logotype.
	Logo *string `json:"logo,omitempty"`
}
