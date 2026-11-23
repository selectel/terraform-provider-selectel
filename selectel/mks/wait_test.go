package mks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollTransportErrorKeepsWaiting(t *testing.T) {
	shortPollInterval.Do(func() { pollInterval = time.Millisecond })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	calls := 0
	err := poll(ctx, func() (bool, error) {
		calls++

		return false, &net.OpError{Op: "dial", Err: fmt.Errorf("connection refused")}
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Greater(t, calls, 1)
}

func TestTaskWaiterTaskAPIError(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()

	_, err := newTaskWaiter(context.Background(), client, "gone")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error listing tasks of cluster gone")

	fake.seedCluster(testClusterID)
	fake.seedTask(testClusterID, "", "CREATE_CLUSTER", false)
	fake.fail(routeTask, http.StatusForbidden)
	err = newCreatedTaskWaiter(client, testClusterID).Wait(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting task")

	fake.fail(routeTasks, http.StatusForbidden)
	err = newCreatedTaskWaiter(client, testClusterID).WaitClusterDeleted(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error listing tasks of cluster")
}

func TestTaskWaiterSurvivesTransientAPIErrors(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()

	fake.seedCluster(testClusterID)
	fake.failFor(routeTasks, http.StatusServiceUnavailable, 1)
	fake.failFor(routeTask, http.StatusTooManyRequests, 1)
	fake.seedTask(testClusterID, "", "CREATE_CLUSTER", false)
	err := newCreatedTaskWaiter(client, testClusterID).Wait(context.Background())
	require.NoError(t, err)

	fake.failFor(routeTasks, http.StatusTooManyRequests, 1)
	fake.failFor(routeTask, http.StatusServiceUnavailable, 1)
	fake.removeCluster(testClusterID)
	err = newCreatedTaskWaiter(client, testClusterID).WaitClusterDeleted(context.Background())
	require.NoError(t, err)
}
