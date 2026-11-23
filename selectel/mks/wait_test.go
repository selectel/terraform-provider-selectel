package mks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskWaiterTaskAPIError(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()

	_, err := newTaskWaiter(context.Background(), client, "gone", time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error listing tasks of cluster gone")

	fake.seedCluster(testClusterID)
	fake.seedTask(testClusterID, "", "CREATE_CLUSTER", false)
	fake.fail(routeTask, http.StatusForbidden)
	err = newCreatedTaskWaiter(client, testClusterID, time.Millisecond).Wait(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting task")

	fake.fail(routeTasks, http.StatusForbidden)
	err = newCreatedTaskWaiter(client, testClusterID, time.Millisecond).WaitClusterDeleted(context.Background())
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
	err := newCreatedTaskWaiter(client, testClusterID, time.Millisecond).Wait(context.Background())
	require.NoError(t, err)

	fake.failFor(routeTasks, http.StatusTooManyRequests, 1)
	fake.failFor(routeTask, http.StatusServiceUnavailable, 1)
	fake.removeCluster(testClusterID)
	err = newCreatedTaskWaiter(client, testClusterID, time.Millisecond).WaitClusterDeleted(context.Background())
	require.NoError(t, err)

	fake.server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err = newCreatedTaskWaiter(client, testClusterID, time.Millisecond).Wait(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
