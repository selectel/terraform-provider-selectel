package mks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/selectel/mks-go/v2/pkg/task"
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

func mksError(status int) error {
	return &mksclient.MKSError{StatusCode: status, Message: "fake error"}
}

func TestWaitDeletableUnsettledStatus(t *testing.T) {
	shortPollInterval.Do(func() { pollInterval = time.Millisecond })
	deadline := time.Now().Add(20 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	err := waitDeletable(ctx, func() ([]statusValue, error) {
		return []statusValue{{object: objectCluster, status: "PENDING_DELETE", accepted: []string{"ACTIVE"}}}, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster status PENDING_DELETE does not allow the delete yet")
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

	fake.fail(routeTask, 0)
	fake.fail(routeTasks, http.StatusForbidden)
	_, err = newTaskWaiter(context.Background(), client, testClusterID)
	require.Error(t, err)

	err = newCreatedTaskWaiter(client, testClusterID).WaitClusterDeleted(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error listing tasks of cluster")

	fake.fail(routeTasks, 0)
	fake.removeCluster(testClusterID)
	err = newCreatedTaskWaiter(client, testClusterID).WaitClusterDeleted(context.Background())
	assert.NoError(t, err)
}

func TestTaskWaiterAdvanceStatuses(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()
	ctx := context.Background()

	fake.seedCluster(testClusterID)
	fake.seedTask(testClusterID, "", "NEW", false)
	w := newCreatedTaskWaiter(client, testClusterID)
	err := w.Wait(ctx)
	assert.NoError(t, err)

	fake.seedTask(testClusterID, "", "ANOTHER", false)
	fake.failTasks("ANOTHER", true)
	w2 := newCreatedTaskWaiter(client, testClusterID)
	err = w2.Wait(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ended in ERROR")

	err = w.WaitClusterDeleted(ctx)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "task") || strings.Contains(err.Error(), "timeout"), err)
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

func TestRealSDKErrorIsTransient(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()
	fake.seedCluster(testClusterID)
	fake.fail(routeTask, http.StatusTooManyRequests)
	_, err := task.Get(context.Background(), client, testClusterID, "missing")
	require.Error(t, err)
	assert.True(t, isTransient(err))
}
