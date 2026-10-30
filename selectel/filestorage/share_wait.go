package filestorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/message"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/sharenetwork"
)

const (
	statusAvailable      = "available"
	statusCreating       = "creating"
	statusExtending      = "extending"
	statusDeleting       = "deleting"
	statusError          = "error"
	statusErrorDeleting  = "error_deleting"
	statusExtendingError = "extending_error"
)

// waitShareAvailable waits for a created share. A status outside creating and
// available, such as error, ends the wait.
func waitShareAvailable(ctx context.Context, client *sdk.Client, id string, interval time.Duration) (*share.View, error) {
	current, err := waitShare(ctx, client, id, interval, []string{statusCreating}, []string{statusAvailable}, nil)
	if err != nil {
		return nil, err
	}

	return current.(*share.View), nil
}

// waitShareExtended waits until the share is available at the new size. The
// fork sets extending before it accepts the request, so available at the old
// size means the scheduler rejected the extension; that read and
// extending_error end the wait.
func waitShareExtended(
	ctx context.Context, client *sdk.Client, id string, size int, interval time.Duration,
) (*share.View, error) {
	status := func(current *share.View) string {
		if current.Status == statusAvailable && current.Size < size {
			return fmt.Sprintf("%s at size %d", statusAvailable, current.Size)
		}

		return current.Status
	}
	current, err := waitShare(ctx, client, id, interval, []string{statusExtending}, []string{statusAvailable}, status)
	if err != nil {
		return nil, err
	}

	return current.(*share.View), nil
}

// waitShareDeleted waits until the share is gone. error_deleting is outside
// pending, so it ends the wait with the status.
func waitShareDeleted(ctx context.Context, client *sdk.Client, id string, interval time.Duration) error {
	_, err := waitShare(ctx, client, id, interval, []string{statusDeleting, statusAvailable, statusError}, nil, nil)

	return err
}

// waitShare polls the share status until the operation deadline, which every
// caller sets. An empty target waits for the share to disappear. status maps a
// read share to its wait state; nil uses the API status.
func waitShare(
	ctx context.Context,
	client *sdk.Client,
	id string,
	interval time.Duration,
	pending, target []string,
	status func(*share.View) string,
) (any, error) {
	if status == nil {
		status = func(current *share.View) string { return current.Status }
	}
	deadline, _ := ctx.Deadline()
	// The last observed share answers a poll that failed with a transient
	// error: a nil result would count as a missing share.
	last := &share.View{ID: id, Status: pending[0]}
	wait := retry.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Timeout:      time.Until(deadline),
		PollInterval: interval,
		Refresh: func() (any, string, error) {
			current, _, err := share.Get(ctx, client, id)
			switch {
			case sdk.IsKind(err, sdk.KindNotFound):
				return nil, "", nil
			case isTransient(err):
				return last, status(last), nil
			case err != nil:
				return nil, "", fmt.Errorf("observe file storage %s: %w", id, err)
			}
			last = current

			return current, status(current), nil
		},
	}
	current, err := wait.WaitForStateContext(ctx)
	var unexpected *retry.UnexpectedStateError
	switch {
	case errors.As(err, &unexpected):
		return nil, shareFailure(ctx, client, id, unexpected.State)
	case err != nil:
		return nil, fmt.Errorf("wait for file storage %s: %w", id, err)
	}

	return current, nil
}

// shareFailure reports a share that ended a wait in a failed status. The share
// itself carries no reason: the API records it as a user message, which is
// added when it can be read.
func shareFailure(ctx context.Context, client *sdk.Client, id, status string) error {
	failure := fmt.Errorf("file storage %s entered status %s", id, status)
	messages, _, err := message.List(ctx, client, message.ListOpts{ResourceID: id})
	if err != nil || len(messages) == 0 {
		return failure
	}

	return fmt.Errorf("%w: %s", failure, messages[0].UserMessage)
}

// waitShareNetworkDeleted waits until the fork removes the service share
// network after the share is gone.
func waitShareNetworkDeleted(ctx context.Context, client *sdk.Client, id string, interval time.Duration) error {
	deadline, _ := ctx.Deadline()
	wait := retry.StateChangeConf{
		Pending:      []string{"present"},
		Target:       nil,
		Timeout:      time.Until(deadline),
		PollInterval: interval,
		Refresh: func() (any, string, error) {
			_, _, err := sharenetwork.Get(ctx, client, id)
			switch {
			case sdk.IsKind(err, sdk.KindNotFound):
				return nil, "", nil
			case err != nil && !isTransient(err):
				return nil, "", fmt.Errorf("observe share network %s: %w", id, err)
			}

			return id, "present", nil
		},
	}
	if _, err := wait.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("wait for share network %s: %w", id, err)
	}

	return nil
}
