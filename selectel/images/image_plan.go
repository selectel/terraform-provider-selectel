package images

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *imageResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config imageModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	properties := config.Properties.Elements()
	// Keys that differ only in letter case make the Image service fail with
	// an internal error, so the first key of each spelling is remembered.
	byLowercase := map[string]string{}
	for _, key := range slices.Sorted(maps.Keys(properties)) {
		at := path.Root("properties").AtMapKey(key)
		if argument, base := baseFields[key]; base {
			detail := "The key is an image field that the API manages."
			if argument != "" {
				detail = fmt.Sprintf("The key is an image field. Set the %s argument instead.", argument)
			}
			resp.Diagnostics.AddAttributeError(at, "Reserved property key", detail)
		}
		if strings.HasPrefix(key, reservedPropertyPrefix) {
			resp.Diagnostics.AddAttributeError(
				at, "Reserved property key", "The "+reservedPropertyPrefix+" prefix is reserved by the API.",
			)
		}
		if first, exists := byLowercase[strings.ToLower(key)]; exists {
			resp.Diagnostics.AddAttributeError(
				at, "Duplicate property key", fmt.Sprintf("The key differs from %q only in letter case.", first),
			)
		} else {
			byLowercase[strings.ToLower(key)] = key
		}
		if properties[key].IsNull() {
			resp.Diagnostics.AddAttributeError(at, "Null property value", "Property values must be strings.")
		}
	}
}

// ModifyPlan requires the image data whenever Terraform creates the image.
// The schema cannot require a source: an imported image has none.
func (r *imageResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !req.State.Raw.IsNull() {
		var prior imageModel
		resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
		if resp.Diagnostics.HasError() || !hasCreationChange(prior, plan) {
			return
		}
	}
	// A source that is not known yet is not null: apply checks it.
	if !resp.Diagnostics.HasError() && plan.LocalFilePath.IsNull() && plan.ImageSourceURL.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("local_file_path"),
			"Missing image data",
			"Set local_file_path or image_source_url: Terraform creates a new image and needs its data.",
		)
	}
}

// sourceReplaced reports whether a change of the image data source replaces
// the image. The data of an existing image cannot change, so another file or
// URL, or a switch between them, replaces it. Setting a source on an image
// that has none, such as an imported image, or removing the source only
// updates the state.
func sourceReplaced(prior, plan imageModel) bool {
	hadSource := !prior.LocalFilePath.IsNull() || !prior.ImageSourceURL.IsNull()
	hasSource := !plan.LocalFilePath.IsNull() || !plan.ImageSourceURL.IsNull()
	sameSource := prior.LocalFilePath.Equal(plan.LocalFilePath) && prior.ImageSourceURL.Equal(plan.ImageSourceURL)

	return hadSource && hasSource && !sameSource
}

// hasCreationChange reports whether the plan replaces the image. It repeats
// the RequiresReplace rules of the schema, so change both together: Framework
// runs ModifyPlan with an empty RequiresReplace list, and the decisions of the
// attribute plan modifiers are not available here.
func hasCreationChange(prior, plan imageModel) bool {
	for _, change := range []struct{ before, after types.String }{
		{prior.ProjectID, plan.ProjectID},
		{prior.Region, plan.Region},
		{prior.ContainerFormat, plan.ContainerFormat},
		{prior.DiskFormat, plan.DiskFormat},
	} {
		if !change.after.Equal(change.before) {
			return true
		}
	}

	return sourceReplaced(prior, plan)
}
