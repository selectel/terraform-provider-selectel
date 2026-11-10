package images

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/images-go/pkg/v1"
	"github.com/selectel/images-go/pkg/v1/image"
)

// untaintHint follows a wait that timed out: the Image service may still
// finish the data transfer.
const untaintHint = "The image stays in the state and is marked as tainted. If the image becomes active later, run terraform untaint for this resource to keep it."

// importTimeoutHint follows a download request that got no answer in time.
// The Image service answers it before the download starts, so a slow answer
// points at the URL rather than at the image size.
const importTimeoutHint = "The API did not accept the download request in time. Check that the URL responds quickly and is reachable from the Selectel network. The image stays in the state and is marked as tainted, so the next apply replaces it."

// uncertainCreateHint follows a create request without a definite answer.
const uncertainCreateHint = "The API may have registered the image. If an image named %q is in the queued status, delete it: Terraform does not track it."

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	duration, d := plan.Timeouts.Create(ctx, imageTimeouts.Create)
	resp.Diagnostics.Append(d...)
	tags, d := setStrings(ctx, plan.Tags)
	resp.Diagnostics.Append(d...)
	properties, d := mapStrings(ctx, plan.Properties)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	// Without a local file, the Image service downloads the data from the URL.
	// Open the file before registering the image, so a wrong path leaves
	// nothing behind.
	download := plan.LocalFilePath.IsNull()
	var file *os.File
	if !download {
		var err error
		file, err = os.Open(plan.LocalFilePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("local_file_path"), "Open image file", err.Error())
			return
		}
		defer file.Close()
	}
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	// The image is protected only once its data is in place: a protected
	// image that failed to become active could not be deleted.
	unprotected, hidden := false, plan.Hidden.ValueBool()
	created, _, err := image.Create(ctx, client, image.CreateOpts{
		Name:            plan.Name.ValueString(),
		Visibility:      plan.Visibility.ValueString(),
		Protected:       &unprotected,
		Hidden:          &hidden,
		ContainerFormat: plan.ContainerFormat.ValueString(),
		DiskFormat:      plan.DiskFormat.ValueString(),
		MinDisk:         int(plan.MinDiskGB.ValueInt64()),
		MinRAM:          int(plan.MinRAMMB.ValueInt64()),
		Tags:            tags,
		Properties:      properties,
	})
	if err != nil {
		detail := err.Error()
		// The Image service makes a new image for each request, so a repeated
		// request cannot resolve the outcome.
		if createOutcomeUncertain(err) {
			detail += "\n\n" + fmt.Sprintf(uncertainCreateHint, plan.Name.ValueString())
		}
		resp.Diagnostics.AddError("Create image", detail)

		return
	}
	// Publish identity before the data transfer: from here on, a failure
	// keeps the registered image in the state, and Terraform marks it as
	// tainted.
	registered := plan
	resp.Diagnostics.Append(registered.observe(ctx, created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &registered)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var uploaded string
	if download {
		_, err = image.Import(ctx, client, created.ID, plan.ImageSourceURL.ValueString())
	} else {
		uploaded, err = uploadFile(ctx, client, created.ID, file)
	}
	if err != nil {
		detail := err.Error()
		if download && sdk.IsKind(err, sdk.KindTimeout) {
			detail += "\n\n" + importTimeoutHint
		}
		resp.Diagnostics.AddError("Transfer image data", detail)

		return
	}
	current, err := waitImageActive(ctx, client, created.ID, r.pollInterval, download)
	if err != nil {
		detail := err.Error()
		if timedOut(err) {
			detail += "\n\n" + untaintHint
		}
		resp.Diagnostics.AddError("Wait for active image", detail)

		return
	}
	if !download {
		if err := verifyUpload(current, uploaded); err != nil {
			resp.Diagnostics.AddError("Verify uploaded image", err.Error())
			return
		}
	}
	// The data transfer may change what the create request set: the Image
	// service derives min_disk from the data and overwrites some properties
	// during a download. The same patch enables the protection. No property
	// was managed before the create request, so the patch removes none.
	ops, d := imagePatch(ctx, types.MapNull(types.StringType), plan, current)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(ops) > 0 {
		current, _, err = image.Update(ctx, client, created.ID, ops)
		if err != nil {
			resp.Diagnostics.AddError("Apply image settings", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(plan.observe(ctx, current)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// uploadFile streams the file into the image and returns the SHA-512 digest
// of the bytes that it sent.
func uploadFile(ctx context.Context, client *sdk.Client, id string, file *os.File) (string, error) {
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("read image file: %w", err)
	}
	digest := sha512.New()
	if _, err := image.Upload(ctx, client, id, io.TeeReader(file, digest), info.Size()); err != nil {
		return "", fmt.Errorf("upload %s: %w", file.Name(), err)
	}

	return hex.EncodeToString(digest.Sum(nil)), nil
}

// verifyUpload compares the digest of the sent bytes with the digest that the
// Image service calculated from the stored data. A mismatch means that the
// data changed on the way, for example, the file was modified during the upload.
// The service uses SHA-512 by default (hashing_algorithm of glance-api.conf).
func verifyUpload(current *image.View, uploaded string) error {
	if current.HashAlgo != "sha512" {
		return fmt.Errorf("the API reported a %q checksum instead of sha512", current.HashAlgo)
	}
	if current.HashValue != uploaded {
		return fmt.Errorf(
			"the SHA-512 checksum of image %s is %s, but the uploaded data has %s; replace the image to upload the file again",
			current.ID, current.HashValue, uploaded,
		)
	}

	return nil
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	current, _, err := image.Get(ctx, client, state.ID.ValueString())
	// The project reads its own image, so 404 means that the image is gone
	// and not that the access is missing.
	if imageGone(current, err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read image", err.Error())
		return
	}
	resp.Diagnostics.Append(state.observe(ctx, current)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	id := prior.ID.ValueString()
	current, _, err := image.Get(ctx, client, id)
	if err != nil {
		resp.Diagnostics.AddError("Read image before update", err.Error())
		return
	}
	ops, d := imagePatch(ctx, prior.Properties, plan, current)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A change of the sources alone updates only the state.
	if len(ops) > 0 {
		updated, _, err := image.Update(ctx, client, id, ops)
		if err != nil {
			resp.Diagnostics.AddError("Update image", err.Error())
			// The Image service applies a patch as a whole, so a failed patch
			// changed nothing and the image read before it is still current.
			resp.Diagnostics.Append(prior.observe(ctx, current)...)
			resp.Diagnostics.Append(resp.State.Set(ctx, &prior)...)

			return
		}
		current = updated
	}
	resp.Diagnostics.Append(plan.observe(ctx, current)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// imagePatch is the JSON Patch that makes the image match the plan. An
// unknown min_disk_gb is left to the Image service. previouslyManaged are the
// properties that the prior state owned: a key dropped from the plan is
// removed from the image.
func imagePatch(
	ctx context.Context,
	previouslyManaged types.Map,
	plan imageModel,
	current *image.View,
) ([]image.PatchOp, diag.Diagnostics) {
	var ops []image.PatchOp
	for _, field := range []struct {
		path    string
		changed bool
		value   any
	}{
		{"/name", plan.Name.ValueString() != current.Name, plan.Name.ValueString()},
		{"/visibility", plan.Visibility.ValueString() != current.Visibility, plan.Visibility.ValueString()},
		{"/protected", plan.Protected.ValueBool() != current.Protected, plan.Protected.ValueBool()},
		{"/os_hidden", plan.Hidden.ValueBool() != current.Hidden, plan.Hidden.ValueBool()},
		{"/min_ram", plan.MinRAMMB.ValueInt64() != int64(current.MinRAM), plan.MinRAMMB.ValueInt64()},
		{
			"/min_disk",
			!plan.MinDiskGB.IsUnknown() && plan.MinDiskGB.ValueInt64() != int64(current.MinDisk),
			plan.MinDiskGB.ValueInt64(),
		},
	} {
		if field.changed {
			ops = append(ops, image.PatchOp{Op: image.OpReplace, Path: field.path, Value: field.value})
		}
	}
	var diags diag.Diagnostics
	tags, d := setStrings(ctx, plan.Tags)
	diags.Append(d...)
	if !slices.Equal(slices.Sorted(slices.Values(tags)), slices.Sorted(slices.Values(current.Tags))) {
		ops = append(ops, image.PatchOp{Op: image.OpReplace, Path: "/tags", Value: tags})
	}
	planned, d := mapStrings(ctx, plan.Properties)
	diags.Append(d...)
	for _, key := range slices.Sorted(maps.Keys(planned)) {
		remote, exists := current.Properties[key]
		switch {
		case !exists:
			ops = append(ops, image.PatchOp{Op: image.OpAdd, Path: propertyPath(key), Value: planned[key]})
		case remote != planned[key]:
			ops = append(ops, image.PatchOp{Op: image.OpReplace, Path: propertyPath(key), Value: planned[key]})
		}
	}
	// Removing a missing property fails with a conflict, so only the keys
	// that the image still has are removed.
	for _, key := range slices.Sorted(maps.Keys(previouslyManaged.Elements())) {
		_, kept := planned[key]
		_, exists := current.Properties[key]
		if !kept && exists {
			ops = append(ops, image.PatchOp{Op: image.OpRemove, Path: propertyPath(key)})
		}
	}

	return ops, diags
}

// propertyPath is the JSON Patch path of a property: a JSON Pointer (RFC 6901)
// in which "~" and "/" are escaped. The Image service splits the path on "/",
// so an unescaped "/" in a key would address another property.
func propertyPath(key string) string {
	return "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(key)
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	duration, d := state.Timeouts.Delete(ctx, imageTimeouts.Delete)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	id := state.ID.ValueString()
	current, _, err := image.Get(ctx, client, id)
	if imageGone(current, err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read image before deletion", err.Error())
		return
	}
	if current.Protected {
		resp.Diagnostics.AddError(
			"Image is protected",
			"Set protected = false and apply this change on its own, then destroy or replace the image. The image stays in the state.",
		)

		return
	}
	if _, err := image.Delete(ctx, client, id); err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
		resp.Diagnostics.AddError("Delete image", err.Error())
		return
	}
	if err := waitImageDeleted(ctx, client, id, r.pollInterval); err != nil {
		resp.Diagnostics.AddError("Wait for deleted image", err.Error())
	}
}
