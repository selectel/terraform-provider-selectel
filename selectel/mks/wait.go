package mks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/cluster"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/selectel/mks-go/v2/pkg/nodegroup"
	"github.com/selectel/mks-go/v2/pkg/task"
)

// pollInterval is how often the waiters poll the API. Tests shorten it.
var pollInterval = 10 * time.Second

// readAfterWaitTimeout bounds the read that saves a created object after
// its wait failed, which can be a timeout or an interrupt.
const readAfterWaitTimeout = 2 * time.Minute

// readAfterWaitContext outlives ctx, so a created object can still be read
// and saved when the wait for it ran out of time or was interrupted.
func readAfterWaitContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), readAfterWaitTimeout)
}

// isNotFound reports whether the API answered 404.
func isNotFound(err error) bool {
	var mksErr *mksclient.MKSError

	return errors.As(err, &mksErr) && mksErr.StatusCode == http.StatusNotFound
}

// clusterOperation is a call the API answers 409 to unless the cluster and
// its nodegroups are in the statuses it names.
type clusterOperation struct {
	name              string
	clusterStatuses   []string
	nodegroupStatuses []string
}

var (
	clusterUpdate = clusterOperation{
		name:            "update",
		clusterStatuses: []string{"ACTIVE", "PENDING_UPDATE_CONTROL_PLANE_LOGS"},
	}
	clusterUpgrade = clusterOperation{
		name:              "upgrade",
		clusterStatuses:   []string{"ACTIVE"},
		nodegroupStatuses: []string{"ACTIVE", "PENDING_CLEANUP"},
	}
	clusterDelete = clusterOperation{
		name:              "delete",
		clusterStatuses:   []string{"ACTIVE", "MAINTENANCE", "ERROR"},
		nodegroupStatuses: []string{"ACTIVE", "ERROR"},
	}
)

// statusValue is the status of an object and the statuses an operation takes.
type statusValue struct {
	object   string
	status   string
	accepted []string
}

// waitStatus polls until the cluster and its nodegroups are in statuses the
// operation takes, so it does not fail with 409 while an earlier task runs.
// An object in ERROR the operation does not take ends the wait at once.
func waitStatus(ctx context.Context, client *mksv2.ServiceClient, clusterID string, op clusterOperation) error {
	var unsettled []string
	err := poll(ctx, func() (bool, error) {
		c, err := cluster.Get(ctx, client, clusterID)
		if err != nil {
			return false, err
		}
		var nodegroups []mksclient.NodegroupListItem
		if op.nodegroupStatuses != nil {
			nodegroups, err = nodegroup.List(ctx, client, clusterID)
			if err != nil {
				return false, err
			}
		}
		statuses := []statusValue{{object: objectCluster, status: string(c.Status), accepted: op.clusterStatuses}}
		for _, ng := range nodegroups {
			statuses = append(statuses, statusValue{object: "nodegroup " + ng.Id, status: string(ng.Status), accepted: op.nodegroupStatuses})
		}

		unsettled = unsettled[:0]
		for _, s := range statuses {
			if slices.Contains(s.accepted, s.status) {
				continue
			}
			if s.status == "ERROR" {
				unsettled = nil

				return true, fmt.Errorf("%s status ERROR does not allow the %s", s.object, op.name)
			}
			unsettled = append(unsettled, s.object+" status "+s.status)
		}

		return len(unsettled) == 0, nil
	})
	if err != nil && len(unsettled) > 0 {
		return fmt.Errorf("%s does not allow the %s yet: %w", strings.Join(unsettled, ", "), op.name, err)
	}

	return err
}

// taskWaiter waits for the tasks one mutating API call created: the API
// returns no task ID, and the call's tasks are the ones listed after it
// and absent before it.
type taskWaiter struct {
	client    *mksv2.ServiceClient
	clusterID string
	before    map[string]struct{}
}

// newTaskWaiter remembers the tasks in scope before a mutating call.
func newTaskWaiter(ctx context.Context, client *mksv2.ServiceClient, clusterID string) (*taskWaiter, error) {
	w := newCreatedTaskWaiter(client, clusterID)

	tasks, err := w.list(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		w.before[t.Id] = struct{}{}
	}

	return w, nil
}

// newCreatedTaskWaiter counts every task in scope, for an object that did
// not exist before the call.
func newCreatedTaskWaiter(client *mksv2.ServiceClient, clusterID string) *taskWaiter {
	return &taskWaiter{
		client:    client,
		clusterID: clusterID,
		before:    map[string]struct{}{},
	}
}

// Wait polls the new tasks until all of them are DONE. No new tasks means
// the call was synchronous.
func (w *taskWaiter) Wait(ctx context.Context) error {
	pending := []mksclient.Task(nil)

	return poll(ctx, func() (bool, error) {
		if pending == nil {
			var err error
			pending, err = w.newTasks(ctx)
			if err != nil {
				return false, err
			}
		}
		advanced, err := w.advance(ctx, pending)
		if err != nil {
			return false, err
		}
		pending = advanced

		return len(pending) == 0, nil
	})
}

// WaitClusterDeleted polls until the cluster is gone and fails when one of
// the new tasks fails first. A 404 on the tasks means the cluster, and
// everything in it, is gone.
func (w *taskWaiter) WaitClusterDeleted(ctx context.Context) error {
	pending := []mksclient.Task(nil)

	return poll(ctx, func() (bool, error) {
		if pending == nil {
			var err error
			pending, err = w.newTasks(ctx)
			if isNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}
		}
		advanced, err := w.advance(ctx, pending)
		if isNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		pending = advanced

		_, err = cluster.Get(ctx, w.client, w.clusterID)
		if isNotFound(err) {
			return true, nil
		}

		return false, err
	})
}

// list returns every task of the cluster in the waiter scope. It takes them
// in one call: the API orders tasks only by started_at, which tasks of one
// request share, so LIMIT/OFFSET pages could skip or repeat a task.
func (w *taskWaiter) list(ctx context.Context) ([]mksclient.Task, error) {
	// The API turns limit 0 into no LIMIT.
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

// advance refreshes the pending tasks and returns the ones still running.
func (w *taskWaiter) advance(ctx context.Context, pending []mksclient.Task) ([]mksclient.Task, error) {
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
		default:
			// INQUEUE, INPROGRESS, and a status newer than the client keep polling.
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

// callAndWait runs a mutating call on the cluster and waits for the
// cluster tasks it created.
func callAndWait(ctx context.Context, client *mksv2.ServiceClient, clusterID string, call func() error) error {
	waiter, err := newTaskWaiter(ctx, client, clusterID)
	if err != nil {
		return err
	}

	err = call()
	if err != nil {
		return err
	}

	return waiter.Wait(ctx)
}

// poll calls check every pollInterval until it reports done, fails, or ctx
// ends. A transient failure keeps waiting; any other error ends the wait.
func poll(ctx context.Context, check func() (bool, error)) error {
	for {
		done, err := check()
		if done {
			return err
		}
		if err != nil && !isTransient(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout while waiting: %w", ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

// isTransient reports whether a failed check may succeed when repeated:
// a 4xx other than 429 is a real answer and ends the wait, the rest does
// not, since the cluster keeps changing on the service side.
func isTransient(err error) bool {
	var mksErr *mksclient.MKSError
	if errors.As(err, &mksErr) {
		return mksErr.StatusCode == http.StatusTooManyRequests || mksErr.StatusCode >= 500
	}

	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || errors.As(err, new(net.Error))
}

// valueOrZero dereferences an optional field of an API model.
func valueOrZero[T any](v *T) T {
	if v == nil {
		var zero T

		return zero
	}

	return *v
}
