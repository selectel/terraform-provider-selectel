package filestorage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/shareaccess"
)

const (
	ruleActive = "active"
	ruleError  = "error"
)

func (r *shareAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan shareAccessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	duration, d := plan.Timeouts.Create(ctx, accessTimeouts.Create)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	region := plan.Region.ValueString()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	shareID := plan.ShareID.ValueString()
	// The fork accepts a read-only CIFS rule and then fails to apply it.
	target, _, err := share.Get(ctx, client, shareID)
	if err != nil {
		resp.Diagnostics.Append(apiError("Read file storage", region, err))
		return
	}
	if strings.EqualFold(target.ShareProto, protoCIFS) && plan.AccessLevel.ValueString() != levelRW {
		resp.Diagnostics.AddAttributeError(path.Root("access_level"), "Unsupported access level",
			fmt.Sprintf("File storage %s uses CIFS, which supports only access_level = %q.", shareID, levelRW))

		return
	}
	token, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError("Generate creation token", err.Error())
		return
	}
	rule, _, err := shareaccess.Allow(ctx, client, shareID, shareaccess.AllowOpts{
		AccessTo:    plan.AccessTo.ValueString(),
		AccessLevel: plan.AccessLevel.ValueString(),
		Metadata:    map[string]string{createTokenKey: token},
	})
	if err != nil && createRejected(err) {
		resp.Diagnostics.Append(apiError("Create file storage access rule", region, err))
		return
	}
	id := ""
	if err == nil {
		id = rule.ID
	}
	if id == "" {
		// The request must never be repeated: the rule may exist. Only a rule
		// with this token is ours; an equal rule without it belongs to someone else.
		cause := err
		if cause == nil {
			cause = fmt.Errorf("the response has no access rule ID")
		}
		id, err = r.awaitRuleByToken(ctx, client, shareID, token)
		if err != nil || id == "" {
			plan.ID = types.StringValue(pendingPrefix + token)
			plan.State = types.StringValue("")
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			if err != nil {
				cause = fmt.Errorf("%w; %w", cause, err)
			}
			resp.Diagnostics.AddError("Access rule creation outcome is uncertain",
				uncertainCreateDetail("access rule", token, cause))

			return
		}
	}
	// Publish identity before waiting, so an error state or a timeout leaves a
	// rule that the next apply replaces.
	plan.ID = types.StringValue(id)
	plan.State = types.StringValue("")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	active, err := r.waitRule(ctx, client, id, []string{ruleActive})
	if err != nil {
		resp.Diagnostics.AddError("Wait for access rule", err.Error())
		return
	}
	plan.observe(active)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *shareAccessModel) observe(rule *shareaccess.View) {
	m.ID = types.StringValue(rule.ID)
	if rule.ShareID != "" {
		m.ShareID = types.StringValue(rule.ShareID)
	}
	m.AccessTo = types.StringValue(rule.AccessTo)
	m.AccessLevel = types.StringValue(rule.AccessLevel)
	m.State = types.StringValue(rule.State)
}

// findRuleByToken returns the rule of the share created with the token, or an
// empty string while no such rule is listed.
func findRuleByToken(ctx context.Context, client *sdk.Client, shareID, token string) (string, error) {
	rules, _, err := shareaccess.List(ctx, client, shareID, shareaccess.ListOpts{
		Metadata: map[string]string{createTokenKey: token},
	})
	if err != nil {
		return "", fmt.Errorf("search for creation token %s: %w", token, err)
	}
	for _, rule := range rules {
		if rule.Metadata[createTokenKey] == token {
			return rule.ID, nil
		}
	}

	return "", nil
}

func (r *shareAccessResource) awaitRuleByToken(
	ctx context.Context, client *sdk.Client, shareID, token string,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.recoveryWindow)
	defer cancel()
	for {
		id, err := findRuleByToken(ctx, client, shareID, token)
		if id != "" || (err != nil && !isTransient(err)) {
			return id, err
		}
		select {
		case <-ctx.Done():
			return "", nil
		case <-time.After(r.pollInterval):
		}
	}
}

// recoverRuleID resolves a pending:<token> ID to the UUID of the rule created
// with that token; a UUID is returned as is.
func recoverRuleID(ctx context.Context, client *sdk.Client, shareID, id string) (string, error) {
	token, pending := strings.CutPrefix(id, pendingPrefix)
	if !pending {
		return id, nil
	}
	found, err := findRuleByToken(ctx, client, shareID, token)
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf(
			"no access rule with creation token %s is visible yet; check the project and refresh again. "+
				"If no access rule with this token appears, run terraform state rm for this resource",
			token,
		)
	}

	return found, nil
}

// waitRule polls the rule until it reaches the target state; an empty target
// waits for the rule to disappear. error and other states end the wait.
func (r *shareAccessResource) waitRule(
	ctx context.Context, client *sdk.Client, id string, target []string,
) (*shareaccess.View, error) {
	deadline, _ := ctx.Deadline()
	pending := []string{"queued_to_apply", "applying"}
	if len(target) == 0 {
		pending = []string{"queued_to_deny", "denying", ruleActive}
	}
	last := &shareaccess.View{ID: id, State: pending[0]}
	wait := retry.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Timeout:      time.Until(deadline),
		PollInterval: r.pollInterval,
		Refresh: func() (any, string, error) {
			current, _, err := shareaccess.Get(ctx, client, id)
			switch {
			case sdk.IsKind(err, sdk.KindNotFound):
				return nil, "", nil
			case isTransient(err):
				return last, last.State, nil
			case err != nil:
				return nil, "", fmt.Errorf("observe access rule %s: %w", id, err)
			}
			last = current

			return current, current.State, nil
		},
	}
	current, err := wait.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for access rule %s: %w", id, err)
	}
	if current == nil {
		return nil, nil
	}

	return current.(*shareaccess.View), nil
}

func (r *shareAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state shareAccessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	region := state.Region.ValueString()
	client, err := r.client(ctx, state.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	id, err := recoverRuleID(ctx, client, state.ShareID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(apiError("Recover access rule", region, err))
		return
	}
	// The fork deletes the rules of a deleted share, so a missing share also
	// reads as a missing rule.
	rule, _, err := shareaccess.Get(ctx, client, id)
	if sdk.IsKind(err, sdk.KindNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(apiError("Read access rule", region, err))
		return
	}
	state.observe(rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only stores new timeouts: every other argument replaces the rule.
func (r *shareAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior shareAccessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.State = prior.State
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *shareAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state shareAccessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	duration, d := state.Timeouts.Delete(ctx, accessTimeouts.Delete)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	region := state.Region.ValueString()
	client, err := r.client(ctx, state.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	shareID := state.ShareID.ValueString()
	id, err := recoverRuleID(ctx, client, shareID, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(apiError("Recover access rule", region, err))
		return
	}
	if _, err := shareaccess.Deny(ctx, client, shareID, id); err != nil {
		if sdk.IsKind(err, sdk.KindNotFound) {
			return
		}
		resp.Diagnostics.Append(apiError("Revoke access rule", region, err))

		return
	}
	if _, err := r.waitRule(ctx, client, id, nil); err != nil {
		resp.Diagnostics.AddError("Wait for revoked access rule", err.Error())
	}
}

func (r *shareAccessResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	shareID, accessID, ok := strings.Cut(req.ID, "/")
	if !ok || !uuidPattern.MatchString(shareID) || !uuidPattern.MatchString(accessID) {
		resp.Diagnostics.AddError("Invalid access rule import ID",
			fmt.Sprintf("Expected <share_id>/<access_id> with two UUIDs, got %q.", req.ID))

		return
	}
	project, region, diags := r.importScope()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), accessID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), shareID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
}
