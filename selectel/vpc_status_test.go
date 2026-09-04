package selectel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func vpcV2StatusSequence(statuses ...string) vpcV2StatusRefreshFunc {
	i := 0

	return func(context.Context) (string, error) {
		if i >= len(statuses) {
			return statuses[len(statuses)-1], nil
		}
		status := statuses[i]
		i++

		return status, nil
	}
}

func TestWaitForVPCV2Status(t *testing.T) {
	const (
		interval = time.Millisecond
		timeout  = 50 * time.Millisecond
	)
	pending := []string{vpcV2StatusDown}

	t.Run("already settled", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusActive))
		require.NoError(t, err)
		assert.Equal(t, vpcV2StatusActive, status)
	})

	t.Run("pending then active", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, vpcV2StatusDown, vpcV2StatusActive))
		require.NoError(t, err)
		assert.Equal(t, vpcV2StatusActive, status)
	})

	t.Run("unknown status counts as settled", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, "N/A"))
		require.NoError(t, err)
		assert.Equal(t, "N/A", status)
	})

	t.Run("error status", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, vpcV2StatusError))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "object id is in ERROR status")
		assert.Equal(t, vpcV2StatusError, status)
	})

	t.Run("timeout keeps last status", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown))
		require.ErrorIs(t, err, errVPCV2StatusTimeout)
		assert.Equal(t, vpcV2StatusDown, status)
	})

	t.Run("refresh error", func(t *testing.T) {
		refreshErr := errors.New("boom")
		_, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, func(context.Context) (string, error) { return "", refreshErr })
		require.ErrorIs(t, err, refreshErr)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := waitForVPCV2Status(ctx, "object", "id", pending,
			time.Hour, time.Hour, vpcV2StatusSequence(vpcV2StatusDown))
		require.ErrorIs(t, err, context.Canceled)
	})
}
