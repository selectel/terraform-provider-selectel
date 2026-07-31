package selectel

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
)

// Error classification for the networking resources.
//
// Every predicate classifies by the typed error class the SDK reports, never by
// matching a substring of the message, so a changed message in Neutron or in the
// SDK cannot silently turn one outcome into another.

// networkingIsNotFound reports whether the API answered "does not exist".
//
// The SDK documents that a not-found class returned by update or delete may be a
// Neutron policy failure masked as absence, so on those paths this predicate
// must be paired with a confirming read - see networkingCheckDeletedOnDelete.
func networkingIsNotFound(err error) bool {
	return vpc.IsErrorClass(err, vpc.ErrorClassNotFound)
}

// networkingIsForbidden reports whether the caller was denied access. A denial
// means the object may well exist, so it must never be read as absence.
func networkingIsForbidden(err error) bool {
	return vpc.IsErrorClass(err, vpc.ErrorClassForbidden) ||
		vpc.IsErrorClass(err, vpc.ErrorClassNotAuthenticated)
}

// networkingIsConflict reports whether the API refused the request because of a
// conflicting or dependent state, for example a network that still has ports.
func networkingIsConflict(err error) bool {
	return vpc.IsErrorClass(err, vpc.ErrorClassConflict)
}

// networkingIsBadRequest reports whether the API rejected the request body.
// Server-side validation the provider cannot reproduce locally - the
// allowed_router_zones allowlist, the extra-route CIDR exclusion, duplicate
// rules or routes - surfaces this way.
func networkingIsBadRequest(err error) bool {
	return vpc.IsErrorClass(err, vpc.ErrorClassBadRequest)
}

// networkingIsServerError reports whether the failure came from the server side
// and may succeed later.
func networkingIsServerError(err error) bool {
	return vpc.IsErrorClass(err, vpc.ErrorClassServer)
}

// networkingIsTransportError reports whether the request never produced an HTTP
// response: a timeout, a broken connection or any other network-level failure.
func networkingIsTransportError(err error) bool {
	if err == nil {
		return false
	}

	var transportErr *vpc.TransportError

	return errors.As(err, &transportErr)
}

// networkingCheckDeleted handles the outcome of a Read: a not-found means the
// object is gone, so the resource leaves the state without a diagnostic error,
// while any other failure - denied access, conflict, server error, transport
// failure - leaves the state untouched and is reported as a diagnostic error.
//
// It is deliberately not used for Create or Update: a not-found there means the
// request was rejected, not that the managed object is absent.
func networkingCheckDeleted(d *schema.ResourceData, err error, object string) diag.Diagnostics {
	if err == nil {
		return nil
	}

	if networkingIsNotFound(err) {
		d.SetId("")

		return nil
	}

	return diag.FromErr(errGettingObject(object, d.Id(), err))
}

// networkingCheckDeletedOnDelete handles the outcome of a Delete.
//
// A not-found from Delete alone does not prove the object is gone: Neutron can
// mask a policy failure as absence. The confirm function therefore performs a
// read of the same object and returns its error, and Delete succeeds only when
// that read also reports not-found. A read that finds the object, or that fails
// for any other reason, is reported as a diagnostic error so the resource is not
// dropped from the state while it still exists remotely.
func networkingCheckDeletedOnDelete(
	d *schema.ResourceData, err error, object string, confirm func() error,
) diag.Diagnostics {
	if err == nil {
		return nil
	}

	if !networkingIsNotFound(err) {
		return diag.FromErr(errDeletingObject(object, d.Id(), err))
	}

	confirmErr := confirm()

	switch {
	case confirmErr == nil:
		// The object is still there, so the not-found was not absence.
		return diag.FromErr(fmt.Errorf(
			"error deleting %s '%s': the api reported it as absent, but it still exists, "+
				"which usually means the delete was denied by policy: %w", object, d.Id(), err))
	case networkingIsNotFound(confirmErr):
		// Confirmed absent, so the delete is idempotently successful.
		return nil
	default:
		return diag.FromErr(fmt.Errorf(
			"error deleting %s '%s': the api reported it as absent, and confirming that failed: %w",
			object, d.Id(), confirmErr))
	}
}
