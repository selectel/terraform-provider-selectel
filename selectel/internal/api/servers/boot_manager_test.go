package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestServiceClient_OperatingSystems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [{
				"uuid": "1",
				"os_name": "Ubuntu",
				"os_value": "ubuntu-20.04",
				"arch": "x86_64",
				"version_value": "20.04",
				"script_allowed": true,
				"is_ssh_key_allowed": true,
				"partitioning": true,
				"template_version": "v1.0",
				"default_partitions": [{
					"type": "partition",
					"device": "/dev/sda1",
					"size": 10240,
					"fstype": "ext4",
					"mount": "/"
				}]
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		ops, respRes, err := client.OperatingSystems(context.Background(), OperatingSystemsQuery{
			LocationID: "locid",
			ServiceID:  "serviceid",
		})

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantOps := OperatingSystems{
			&OperatingSystem{
				UUID:            "1",
				Name:            "Ubuntu",
				OSValue:         "ubuntu-20.04",
				Arch:            "x86_64",
				VersionValue:    "20.04",
				ScriptAllowed:   true,
				IsSSHKeyAllowed: true,
				Partitioning:    true,
				TemplateVersion: "v1.0",
				DefaultPartitions: []*PartitionConfigItem{
					{
						Type:   "partition",
						Device: "/dev/sda1",
						Size:   10240,
						FSType: "ext4",
						Mount:  "/",
					},
				},
			},
		}
		require.Equal(t, wantOps, ops)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		ops, respRes, err := client.OperatingSystems(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, ops)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		ops, respRes, err := client.OperatingSystems(context.Background())

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, ops)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		ops, respRes, err := client.OperatingSystems(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, ops)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_LocalDrives(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": {
				"drive1": {
					"type": "SSD", 
					"match": {
						"size": 500, 
						"type": "SSD"
					}
				}
			}
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		drives, respRes, err := client.LocalDrives(context.Background(), "serviceid")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantDrives := LocalDrives{
			"drive1": &LocalDrive{
				Type: "SSD",
				Match: &LocalDriveMatch{
					Size: 500,
					Type: "SSD",
				},
			},
		}
		require.Equal(t, wantDrives, drives)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		drives, respRes, err := client.LocalDrives(context.Background(), "serviceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, drives)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		drives, respRes, err := client.LocalDrives(context.Background(), "serviceid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, drives)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		drives, respRes, err := client.LocalDrives(context.Background(), "serviceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, drives)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_PartitionsValidate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"partitions_config": {
				"partition1": {
					"type": "local_drive",
					"match": {
						"size": 500,
						"type": "SSD"
					}
				}
			}
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		config := PartitionsConfig{
			"partition1": &PartitionConfigItem{
				Type: "local_drive",
				Match: &PartitionConfigItemMatch{
					Size: 500,
					Type: "SSD",
				},
			},
		}

		// Execute
		validatedConfig, respRes, err := client.PartitionsValidate(context.Background(), config, "serviceid")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		require.Equal(t, config, validatedConfig)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		config := PartitionsConfig{}

		// Execute
		validatedConfig, respRes, err := client.PartitionsValidate(context.Background(), config, "serviceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, validatedConfig)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		config := PartitionsConfig{}

		// Execute
		validatedConfig, respRes, err := client.PartitionsValidate(context.Background(), config, "serviceid")

		// Validate
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, validatedConfig)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		config := PartitionsConfig{}

		// Execute
		validatedConfig, respRes, err := client.PartitionsValidate(context.Background(), config, "serviceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, validatedConfig)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_ReinstallOS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": {}
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		payload := &InstallNewOSPayload{
			OSVersion:    "20.04",
			OSTemplate:   "ubuntu",
			OSArch:       "x86_64",
			UserSSHKey:   "ssh-rsa AAAAB3...",
			UserHostname: "test-host",
			Password:     "password123",
		}

		// Execute
		respRes, err := client.InstallNewOS(context.Background(), payload, "resourceid")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		payload := &InstallNewOSPayload{
			OSVersion:  "20.04",
			OSTemplate: "ubuntu",
			OSArch:     "x86_64",
		}

		// Execute
		respRes, err := client.InstallNewOS(context.Background(), payload, "resourceid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		payload := &InstallNewOSPayload{
			OSVersion:  "20.04",
			OSTemplate: "ubuntu",
			OSArch:     "x86_64",
		}

		// Execute
		respRes, err := client.InstallNewOS(context.Background(), payload, "resourceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_OperatingSystemByResource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": {
				"user_ssh_key": "ssh-rsa AAAAB3...",
				"userhostname": "test-host",
				"cloud_init_user_data": "echo Hello",
				"password": "password123",
				"os_template": "ubuntu",
				"arch": "x86_64",
				"version": "20.04",
				"reinstall": 1
			}
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		os, respRes, err := client.OperatingSystemByResource(context.Background(), "resourceid")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantOS := &OperatingSystemAtResource{
			UserSSHKey:   "ssh-rsa AAAAB3...",
			UserHostName: "test-host",
			UserData:     "echo Hello",
			Password:     "password123",
			OSValue:      "ubuntu",
			Arch:         "x86_64",
			Version:      "20.04",
			Reinstall:    1,
		}
		require.Equal(t, wantOS, os)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		os, respRes, err := client.OperatingSystemByResource(context.Background(), "resourceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, os)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		os, respRes, err := client.OperatingSystemByResource(context.Background(), "resourceid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, os)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		os, respRes, err := client.OperatingSystemByResource(context.Background(), "resourceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, os)
		require.Nil(t, respRes)
	})
}

func TestOperatingSystemsQuery_queryParamsRaw(t *testing.T) {
	tests := []struct {
		name   string
		query  OperatingSystemsQuery
		expect string
	}{
		{
			name:   "BothLocationAndServiceID",
			query:  OperatingSystemsQuery{LocationID: "locid", ServiceID: "serviceid"},
			expect: "?service_uuid=serviceid&location_uuid=locid",
		},
		{
			name:   "OnlyLocationID",
			query:  OperatingSystemsQuery{LocationID: "locid"},
			expect: "?location_uuid=locid",
		},
		{
			name:   "OnlyServiceID",
			query:  OperatingSystemsQuery{ServiceID: "serviceid"},
			expect: "?service_uuid=serviceid",
		},
		{
			name:   "NoParams",
			query:  OperatingSystemsQuery{},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.query.queryParamsRaw()
			require.Equal(t, tt.expect, result)
		})
	}
}
