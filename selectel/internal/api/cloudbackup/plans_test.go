package cloudbackup

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestServiceClient_Plans(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"plans": [{
				"id": "plan-id-1",
				"name": "test-plan",
				"description": "test description",
				"backup_mode": "full",
				"full_backups_amount": 5,
				"schedule_pattern": "0 0 * * *",
				"schedule_type": "cron",
				"status": "started",
				"created_at": "2023-01-01T00:00:00Z",
				"resources": [{
					"id": "resource-id-1",
					"name": "resource-name-1",
					"type": "volume"
				}]
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plans, respRes, err := client.Plans(context.Background(), &PlansQuery{Name: "test-plan", VolumeName: "test-volume"})

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantPlans := []*Plan{
			{
				ID:                "plan-id-1",
				Name:              "test-plan",
				Description:       "test description",
				BackupMode:        "full",
				FullBackupsAmount: 5,
				SchedulePattern:   "0 0 * * *",
				ScheduleType:      "cron",
				Status:            "started",
				CreatedAt:         "2023-01-01T00:00:00Z",
				Resources: []*PlanResource{
					{
						ID:   "resource-id-1",
						Name: "resource-name-1",
						Type: "volume",
					},
				},
			},
		}
		require.Equal(t, wantPlans, plans)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plans, respRes, err := client.Plans(context.Background(), nil)

		// Analyse
		require.Error(t, err)
		require.Nil(t, plans)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		plans, respRes, err := client.Plans(context.Background(), nil)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, plans)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plans, respRes, err := client.Plans(context.Background(), nil)

		// Analyse
		require.Error(t, err)
		require.Nil(t, plans)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_Plan(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"id": "plan-id-1",
			"name": "test-plan",
			"description": "test description",
			"backup_mode": "full",
			"full_backups_amount": 5,
			"schedule_pattern": "0 0 * * *",
			"schedule_type": "cron",
			"status": "started",
			"created_at": "2023-01-01T00:00:00Z",
			"resources": [{
				"id": "resource-id-1",
				"name": "resource-name-1",
				"type": "volume"
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.Plan(context.Background(), "plan-id-1")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantPlan := &Plan{
			ID:                "plan-id-1",
			Name:              "test-plan",
			Description:       "test description",
			BackupMode:        "full",
			FullBackupsAmount: 5,
			SchedulePattern:   "0 0 * * *",
			ScheduleType:      "cron",
			Status:            "started",
			CreatedAt:         "2023-01-01T00:00:00Z",
			Resources: []*PlanResource{
				{
					ID:   "resource-id-1",
					Name: "resource-name-1",
					Type: "volume",
				},
			},
		}
		require.Equal(t, wantPlan, plan)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.Plan(context.Background(), "plan-id-1")

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		plan, respRes, err := client.Plan(context.Background(), "plan-id-1")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, plan)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.Plan(context.Background(), "plan-id-1")

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_PlanCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"id": "plan-id-1",
			"name": "test-plan",
			"description": "test description",
			"backup_mode": "full",
			"full_backups_amount": 5,
			"schedule_pattern": "0 0 * * *",
			"schedule_type": "cron",
			"status": "started",
			"created_at": "2023-01-01T00:00:00Z",
			"resources": [{
				"id": "resource-id-1",
				"name": "resource-name-1",
				"type": "volume"
			}]
		}`
		fakeResp := httptest.NewFakeResponse(201, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		createReq := &Plan{
			Name:              "test-plan",
			Description:       "test description",
			BackupMode:        "full",
			FullBackupsAmount: 5,
			SchedulePattern:   "0 0 * * *",
			ScheduleType:      "cron",
			Resources: []*PlanResource{
				{
					ID: "resource-id-1",
				},
			},
		}

		// Execute
		plan, respRes, err := client.PlanCreate(context.Background(), createReq)

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 201, respRes.StatusCode)
		wantPlan := &Plan{
			ID:                "plan-id-1",
			Name:              "test-plan",
			Description:       "test description",
			BackupMode:        "full",
			FullBackupsAmount: 5,
			SchedulePattern:   "0 0 * * *",
			ScheduleType:      "cron",
			Status:            "started",
			CreatedAt:         "2023-01-01T00:00:00Z",
			Resources: []*PlanResource{
				{
					ID:   "resource-id-1",
					Name: "resource-name-1",
					Type: "volume",
				},
			},
		}
		require.Equal(t, wantPlan, plan)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(201, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.PlanCreate(context.Background(), &Plan{})

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.NotNil(t, respRes)
		require.Equal(t, 201, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		plan, respRes, err := client.PlanCreate(context.Background(), &Plan{})

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, plan)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.PlanCreate(context.Background(), &Plan{})

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_PlanUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"id": "plan-id-1",
			"name": "updated-plan-name",
			"description": "updated description",
			"backup_mode": "full",
			"full_backups_amount": 10,
			"schedule_pattern": "0 2 * * *",
			"schedule_type": "cron",
			"status": "started",
			"created_at": "2023-01-01T00:00:00Z",
			"resources": [{
				"id": "resource-id-1",
				"name": "resource-name-1",
				"type": "volume"
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		updateReq := &Plan{
			Name:              "updated-plan-name",
			Description:       "updated description",
			FullBackupsAmount: 10,
			SchedulePattern:   "0 2 * * *",
		}

		// Execute
		plan, respRes, err := client.PlanUpdate(context.Background(), "plan-id-1", updateReq)

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantPlan := &Plan{
			ID:                "plan-id-1",
			Name:              "updated-plan-name",
			Description:       "updated description",
			BackupMode:        "full",
			FullBackupsAmount: 10,
			SchedulePattern:   "0 2 * * *",
			ScheduleType:      "cron",
			Status:            "started",
			CreatedAt:         "2023-01-01T00:00:00Z",
			Resources: []*PlanResource{
				{
					ID:   "resource-id-1",
					Name: "resource-name-1",
					Type: "volume",
				},
			},
		}
		require.Equal(t, wantPlan, plan)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.PlanUpdate(context.Background(), "plan-id-1", &Plan{})

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		plan, respRes, err := client.PlanUpdate(context.Background(), "plan-id-1", &Plan{})

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, plan)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plan, respRes, err := client.PlanUpdate(context.Background(), "plan-id-1", &Plan{})

		// Analyse
		require.Error(t, err)
		require.Nil(t, plan)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_PlanDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		fakeResp := httptest.NewFakeResponse(204, "") //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		respRes, err := client.PlanDelete(context.Background(), "plan-id-1")

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 204, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		respRes, err := client.PlanDelete(context.Background(), "plan-id-1")

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

		// Execute
		respRes, err := client.PlanDelete(context.Background(), "plan-id-1")

		// Analyse
		require.Error(t, err)
		require.Nil(t, respRes)
	})
}
