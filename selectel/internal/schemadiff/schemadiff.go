// Package schemadiff holds CustomizeDiff helpers shared by the resources.
package schemadiff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// RecomputeOnChange marks a computed set as unknown in the plan whenever one of
// the attributes it is derived from changes.
//
// SDK v2 keeps the planned value of a computed set through an apply: when the
// plan already knows the set, the value the provider reads back after the update
// is dropped and the state keeps the stale set until the next refresh. Making the
// set unknown in the plan lets the value from Read land in the state right away.
func RecomputeOnChange(computed string, triggers ...string) schema.CustomizeDiffFunc {
	return func(_ context.Context, d *schema.ResourceDiff, _ any) error {
		for _, key := range triggers {
			if d.HasChange(key) {
				return d.SetNewComputed(computed)
			}
		}

		return nil
	}
}
