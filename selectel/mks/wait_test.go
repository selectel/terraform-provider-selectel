package mks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollTransientErrors(t *testing.T) {
	shortPollInterval.Do(func() { pollInterval = time.Millisecond })
	t.Run("a rate limit and a 5xx keep waiting until the check succeeds", func(t *testing.T) {
		calls := 0
		err := poll(context.Background(), func() (bool, error) {
			calls++
			if calls == 1 {
				return false, mksError(http.StatusTooManyRequests)
			}
			if calls == 2 {
				return false, mksError(http.StatusInternalServerError)
			}

			return true, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, calls)
	})
	t.Run("a 4xx other than 429 ends the wait", func(t *testing.T) {
		calls := 0
		err := poll(context.Background(), func() (bool, error) {
			calls++

			return false, mksError(http.StatusConflict)
		})
		require.Error(t, err)
		assert.Equal(t, 1, calls)
	})
	t.Run("a transport error keeps waiting until the deadline passes", func(t *testing.T) {
		deadline := time.Now().Add(20 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		calls := 0
		err := poll(ctx, func() (bool, error) {
			calls++

			return false, &net.OpError{Op: "dial", Err: fmt.Errorf("connection refused")}
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Greater(t, calls, 1)
	})
}

func TestIsTransient(t *testing.T) {
	for _, tt := range []struct {
		name  string
		err   error
		want  bool
	}{
		{name: "rate limit", err: mksError(http.StatusTooManyRequests), want: true},
		{name: "internal server error", err: mksError(http.StatusInternalServerError), want: true},
		{name: "conflict", err: mksError(http.StatusConflict), want: false},
		{name: "not found", err: mksError(http.StatusNotFound), want: false},
		{name: "timeout", err: context.DeadlineExceeded, want: true},
		{name: "canceled", err: context.Canceled, want: true},
		{name: "transport", err: &net.OpError{Op: "dial", Err: fmt.Errorf("refused")}, want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTransient(tt.err))
		})
	}
}

func mksError(status int) error {
	return mksclient.MKSError{StatusCode: status, Message: "fake error"}
}
