package mks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/selectel/mks-go/v2/pkg/mksclient"
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

func TestWaitActive(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()
	fake.seedCluster(testClusterID)

	got, err := waitActive(context.Background(), client, testClusterID, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, testClusterID, got.Id)

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "ERROR" })
	_, err = waitActive(context.Background(), client, testClusterID, time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster "+testClusterID+" is in ERROR")

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "PENDING_UPDATE" })
	fake.settleAfter(testClusterID, 3, "ACTIVE")
	getsBefore := fake.callCount(routeCluster)
	_, err = waitActive(context.Background(), client, testClusterID, time.Millisecond)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, fake.callCount(routeCluster)-getsBefore, 3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "MAINTENANCE" })
	got, err = waitActive(ctx, client, testClusterID, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, mksclient.ClusterDetailedStatusMAINTENANCE, got.Status)

	fake.updateCluster(testClusterID, func(c *mksclient.ClusterDetailed) { c.Status = "PENDING_UPDATE" })
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = waitActive(ctx, client, testClusterID, time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "last state: cluster "+testClusterID+" is in PENDING_UPDATE")
}
