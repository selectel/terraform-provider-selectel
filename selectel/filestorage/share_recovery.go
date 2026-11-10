package filestorage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
)

const pendingPrefix = "pending:"

// createRejected reports whether the API refused the create request, so no
// share exists. Only a 4xx with a structured API fault is a refusal: a
// proxy page, a rate limit, a server error or a lost answer may hide a share
// that the API accepted.
func createRejected(err error) bool {
	var apiErr *sdk.Error
	if !errors.As(err, &apiErr) || !apiErr.StructuredFault {
		return false
	}

	return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != http.StatusTooManyRequests
}

// findShareByToken returns the UUID of the share created with the token, or
// an empty string while no such share is listed.
func findShareByToken(ctx context.Context, client *sdk.Client, token string) (string, error) {
	shares, err := share.ListDetail(ctx, client, share.ListOpts{Metadata: map[string]string{createTokenKey: token}})
	if err != nil {
		return "", fmt.Errorf("search for creation token %s: %w", token, err)
	}
	found := ""
	for _, candidate := range shares {
		if candidate.Metadata[createTokenKey] != token {
			continue
		}
		if found != "" {
			return "", fmt.Errorf("multiple file storages have creation token %s; import the intended UUID manually", token)
		}
		found = candidate.ID
	}

	return found, nil
}

// awaitShareByToken repeats the search until the share appears, the window
// ends or a search fails.
func awaitShareByToken(
	ctx context.Context, client *sdk.Client, token string, interval, window time.Duration,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, window)
	defer cancel()
	for {
		id, err := findShareByToken(ctx, client, token)
		if id != "" || (err != nil && !isTransient(err)) {
			return id, err
		}
		select {
		case <-ctx.Done():
			return "", nil
		case <-time.After(interval):
		}
	}
}

// recoverShareID resolves a pending:<token> ID to the UUID of the share
// created with that token; a UUID is returned as is.
func recoverShareID(ctx context.Context, client *sdk.Client, id string) (string, error) {
	token, pending := strings.CutPrefix(id, pendingPrefix)
	if !pending {
		return id, nil
	}
	found, err := findShareByToken(ctx, client, token)
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf(
			"no file storage with creation token %s is visible yet; check the project and refresh again. "+
				"If no file storage with this token appears, run terraform state rm for this resource",
			token,
		)
	}

	return found, nil
}

// uncertainCreateDetail explains a pending object, such as a file storage or
// an access rule, after a create request without a definite answer.
func uncertainCreateDetail(noun, token string, cause error) string {
	return fmt.Sprintf(
		"The API gave no definite answer to the create request (%v), and no %s with creation token %s "+
			"appeared. The resource is kept in state as %s%s, and refresh and destroy fail until the %s appears. "+
			"When it appears, the next refresh replaces the temporary ID with its UUID, and the next apply "+
			"replaces the %s; to keep it, run terraform untaint for this resource. If no %s with this token "+
			"appears, run terraform state rm for this resource, and the next apply creates it again.",
		cause, noun, token, pendingPrefix, token, noun, noun, noun,
	)
}
