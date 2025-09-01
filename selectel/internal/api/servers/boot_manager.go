package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OperatingSystemsQuery struct {
	LocationID string `json:"location_id"`
	ServiceID  string `json:"service_id"`
}

func (os OperatingSystemsQuery) queryParamsRaw() string {
	switch {
	case os.LocationID != "" && os.ServiceID != "":
		return "?service_uuid=" + os.ServiceID + "&location_uuid=" + os.LocationID

	case os.LocationID != "":
		return "?location_uuid=" + os.LocationID

	case os.ServiceID != "":
		return "?service_uuid=" + os.ServiceID
	}

	return ""
}

func (client *ServiceClient) OperatingSystems(ctx context.Context, query ...OperatingSystemsQuery) (OperatingSystems, *ResponseResult, error) {
	url := client.Endpoint + "/boot/template/os/new"
	if len(query) > 0 {
		url += query[0].queryParamsRaw()
	}

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		OperatingSystems OperatingSystems `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.OperatingSystems, responseResult, nil
}

func (client *ServiceClient) OperatingSystemByResource(ctx context.Context, resourceID string) (*OperatingSystemAtResource, *ResponseResult, error) {
	url := fmt.Sprintf("%s/boot/os/%s", client.Endpoint, resourceID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		OperatingSystem *OperatingSystemAtResource `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.OperatingSystem, responseResult, nil
}

func (client *ServiceClient) LocalDrives(ctx context.Context, serviceID string) (LocalDrives, *ResponseResult, error) {
	url := fmt.Sprintf("%s/boot/partitions/local_drives?service_uuid=%s", client.Endpoint, serviceID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result LocalDrives `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

type (
	PartitionsConfig map[string]*PartitionConfigItem

	PartitionConfigItem struct {
		Type string `json:"type"`

		// For "local_drive"
		Match *PartitionConfigItemMatch `json:"match,omitempty"`

		// For "partition"
		Device   string  `json:"device,omitempty"`
		Size     float64 `json:"size,omitempty"`
		Priority *int    `json:"priority,omitempty"`

		// For "filesystem"
		FSType string `json:"fstype,omitempty"`
		Mount  string `json:"mount,omitempty"`

		// For "soft_raid"
		Members []string `json:"members,omitempty"`
		Level   string   `json:"level,omitempty"`
	}

	PartitionConfigItemMatch struct {
		Size int    `json:"size"`
		Type string `json:"type"`
	}
)

func (client *ServiceClient) PartitionsValidate(ctx context.Context, config PartitionsConfig, serviceID string) (PartitionsConfig, *ResponseResult, error) {
	url := fmt.Sprintf("%s/boot/partitions/validate?service_id=%s", client.Endpoint, serviceID)

	payload := struct {
		PartitionsConfig PartitionsConfig `json:"partitions_config"`
	}{
		PartitionsConfig: config,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		PartitionsConfig PartitionsConfig `json:"partitions_config"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.PartitionsConfig, responseResult, nil
}

type InstallNewOSPayload struct {
	OSVersion        string           `json:"version"`
	OSTemplate       string           `json:"os_template"`
	OSArch           string           `json:"arch"`
	UserSSHKey       string           `json:"user_ssh_key,omitempty"`
	UserHostname     string           `json:"userhostname"`
	Password         string           `json:"password,omitempty"`
	PartitionsConfig PartitionsConfig `json:"partitions_config,omitempty"`
	UserData         string           `json:"cloud_init_user_data,omitempty"`
}

func (client *ServiceClient) InstallNewOS(
	ctx context.Context, payload *InstallNewOSPayload, resourceID string,
) (*ResponseResult, error) {
	url := fmt.Sprintf("%s/boot/os/%s", client.Endpoint, resourceID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		return responseResult, responseResult.Err
	}

	defer func() { _ = responseResult.Body.Close() }()

	return responseResult, nil
}
