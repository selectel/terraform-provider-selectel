package mks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/cluster"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/selectel/mks-go/v2/pkg/task"
)

// readAfterWaitTimeout bounds the read that saves a created object after
// its wait failed, which can be a timeout or an interrupt.
const readAfterWaitTimeout = 2 * time.Minute

// isNotFound reports whether the API answered 404.
func isNotFound(err error) bool {
	var mksErr *mksclient.MKSError

	return errors.As(err, &mksErr) && mksErr.StatusCode == http.StatusNotFound
}

// waitActive polls until the cluster is ACTIVE or in its maintenance window, and returns it.
// It fails when the cluster is in ERROR. Any other status keeps waiting.
func waitActive(ctx context.Context, client *mksv2.ServiceClient, clusterID string, interval time.Duration) (*mksclient.ClusterDetailed, error) {
	var got *mksclient.ClusterDetailed
	err := poll(ctx, interval, func() (string, bool, error) {
		var err error
		got, err = cluster.Get(ctx, client, clusterID)
		if err != nil {
			return "", false, err
		}

		if got.Status == mksclient.ClusterDetailedStatusERROR {
			return "", true, fmt.Errorf("cluster %s is in ERROR", clusterID)
		}

		ready := got.Status == mksclient.ClusterDetailedStatusACTIVE || got.Status == mksclient.ClusterDetailedStatusMAINTENANCE

		return describeCluster(got), ready, nil
	})
	if err != nil {
		return nil, err
	}

	return got, nil
}

// taskWaiter waits for the tasks one mutating API call created: the ones listed after it and absent before it.
type taskWaiter struct {
	client    *mksv2.ServiceClient
	clusterID string
	interval  time.Duration
	before    map[string]struct{}
}

// newTaskWaiter remembers the tasks of the cluster before a mutating call.
func newTaskWaiter(ctx context.Context, client *mksv2.ServiceClient, clusterID string, interval time.Duration) (*taskWaiter, error) {
	w := newCreatedTaskWaiter(client, clusterID, interval)

	tasks, err := w.list(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		w.before[t.Id] = struct{}{}
	}

	return w, nil
}

// newCreatedTaskWaiter counts every task of the cluster, for an object that did
// not exist before the call.
func newCreatedTaskWaiter(client *mksv2.ServiceClient, clusterID string, interval time.Duration) *taskWaiter {
	return &taskWaiter{
		client:    client,
		clusterID: clusterID,
		interval:  interval,
		before:    map[string]struct{}{},
	}
}

// Wait polls the new tasks until all of them are DONE. No new tasks means
// the call was synchronous.
func (w *taskWaiter) Wait(ctx context.Context) error {
	pending := []mksclient.Task(nil)

	return poll(ctx, w.interval, func() (string, bool, error) {
		var err error
		pending, err = w.advance(ctx, pending)
		if err != nil || len(pending) == 0 {
			return "", err == nil, err
		}

		return describeTask(pending[0]), false, nil
	})
}

// WaitClusterDeleted polls until the cluster is gone and fails when one of the new tasks fails first.
// A 404 on the tasks means the cluster, and everything in it, is gone.
func (w *taskWaiter) WaitClusterDeleted(ctx context.Context) error {
	pending := []mksclient.Task(nil)

	return poll(ctx, w.interval, func() (string, bool, error) {
		var err error
		pending, err = w.advance(ctx, pending)
		if isNotFound(err) {
			return "", true, nil
		}
		if err != nil {
			return "", false, err
		}

		got, err := cluster.Get(ctx, w.client, w.clusterID)
		if isNotFound(err) {
			return "", true, nil
		}
		if err != nil {
			return "", false, err
		}
		if len(pending) > 0 {
			return describeTask(pending[0]), false, nil
		}

		return describeCluster(got), false, nil
	})
}

// list returns every task of the cluster in one call.
func (w *taskWaiter) list(ctx context.Context) ([]mksclient.Task, error) {
	// Limit 0 means no limit.
	tasks, err := task.List(ctx, w.client, w.clusterID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error listing tasks of cluster %s: %w", w.clusterID, err)
	}

	return slices.DeleteFunc(tasks, func(t mksclient.Task) bool {
		return valueOrZero(t.NodegroupId) != ""
	}), nil
}

func (w *taskWaiter) newTasks(ctx context.Context) ([]mksclient.Task, error) {
	tasks, err := w.list(ctx)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(tasks, func(t mksclient.Task) bool {
		_, ok := w.before[t.Id]

		return ok
	}), nil
}

// advance refreshes the pending tasks, the new ones when pending is nil, and returns the ones still running.
// When a refresh fails, it returns the tasks it was refreshing.
func (w *taskWaiter) advance(ctx context.Context, pending []mksclient.Task) ([]mksclient.Task, error) {
	if pending == nil {
		var err error
		pending, err = w.newTasks(ctx)
		if err != nil {
			return nil, err
		}
	}

	var running []mksclient.Task
	for _, p := range pending {
		t, err := task.Get(ctx, w.client, w.clusterID, p.Id)
		if err != nil {
			return pending, fmt.Errorf("error getting task %s of cluster %s: %w", p.Id, w.clusterID, err)
		}

		switch t.Status {
		case mksclient.DONE:
		case mksclient.ERROR, mksclient.CANCELED:
			return nil, waitTaskError(t)
		case mksclient.INQUEUE, mksclient.INPROGRESS:
			running = append(running, *t)
		default:
			// A status newer than the client keeps polling.
			running = append(running, *t)
		}
	}

	return running, nil
}

func waitTaskError(t *mksclient.Task) error {
	msg := fmt.Sprintf("task %s %s of cluster %s ended in %s", t.Type, t.Id, t.ClusterId, t.Status)
	if t.ErrorDetails != nil {
		msg += fmt.Sprintf(": %s (code %d)", t.ErrorDetails.Name, t.ErrorDetails.Code)
		details := valueOrZero(t.ErrorDetails.Details)
		if details != "" {
			msg += ": " + details
		}
	}

	return errors.New(msg)
}

// callAndWait runs a mutating call on the cluster, waits for the cluster
// tasks it created and for the cluster to become ACTIVE.
func callAndWait(ctx context.Context, client *mksv2.ServiceClient, clusterID string, interval time.Duration, call func() error) error {
	waiter, err := newTaskWaiter(ctx, client, clusterID, interval)
	if err != nil {
		return err
	}

	err = call()
	if err != nil {
		return err
	}

	err = waiter.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = waitActive(ctx, client, clusterID, interval)

	return err
}

func describeCluster(c *mksclient.ClusterDetailed) string {
	return fmt.Sprintf("cluster %s is in %s", c.Id, c.Status)
}

func describeTask(t mksclient.Task) string {
	return fmt.Sprintf("task %s %s is in %s", t.Type, t.Id, t.Status)
}

// poll calls check every interval until it reports done, fails, or ctx ends. A transient failure keeps waiting;
// any other error ends the wait. The state check returns while it is not done is named when ctx ends.
func poll(ctx context.Context, interval time.Duration, check func() (string, bool, error)) error {
	last := "unknown"
	for {
		state, done, err := check()
		if done {
			return err
		}
		if err != nil && !isTransient(err) {
			return err
		}
		if err == nil {
			last = state
		} else if ctx.Err() == nil {
			last = err.Error()
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("stopped waiting, last state: %s: %w", last, ctx.Err())
		case <-time.After(interval):
		}
	}
}

// isTransient reports whether a failed check may succeed when repeated: a 4xx other than 429
// is a real answer and ends the wait, the rest does not, since the cluster keeps changing.
func isTransient(err error) bool {
	var mksErr *mksclient.MKSError
	if errors.As(err, &mksErr) {
		return mksErr.StatusCode == http.StatusTooManyRequests || mksErr.StatusCode >= 500
	}

	return errors.As(err, new(net.Error))
}

// valueOrZero dereferences an optional field of an API model.
func valueOrZero[T any](v *T) T {
	if v == nil {
		var zero T

		return zero
	}

	return *v
}
