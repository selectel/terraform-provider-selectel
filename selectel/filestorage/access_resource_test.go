package filestorage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ruleID = "55555555-5555-4555-8555-555555555555"

// fakeRules serves the access rules of one share and records every request.
type fakeRules struct {
	t  *testing.T
	mu sync.Mutex

	proto     string
	shareGone bool
	rule      map[string]any // nil when the rule does not exist
	foreign   []map[string]any
	states    []string // successive states of the rule; the last repeats
	failAllow *answer
	lostAllow *answer // the rule is created, but the answer is lost
	failGet   *answer
	// transientGets is the number of rule reads answered with a transient failure.
	transientGets int
	allowCalls    int
	requests      []string
}

func newFakeRules(t *testing.T) *fakeRules {
	t.Helper()

	return &fakeRules{t: t, proto: protoNFS}
}

func (f *fakeRules) resource() *shareAccessResource {
	server := httptest.NewServer(http.HandlerFunc(f.serve))
	f.t.Cleanup(server.Close)
	client, err := sdk.NewClient(sdk.Config{Endpoint: server.URL, Token: "token"})
	require.NoError(f.t, err)

	return &shareAccessResource{
		connection:     connection{config: testConfig{client: client, project: "project", region: "ru-9"}},
		pollInterval:   time.Millisecond,
		recoveryWindow: 50 * time.Millisecond,
	}
}

func (f *fakeRules) writes() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var writes []string
	for _, r := range f.requests {
		if strings.HasPrefix(r, http.MethodPost) {
			writes = append(writes, r)
		}
	}

	return writes
}

func (f *fakeRules) serve(w http.ResponseWriter, req *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	body, _ := io.ReadAll(req.Body)
	key := req.Method + " " + req.URL.Path
	w.Header().Set("Content-Type", "application/json")
	switch key {
	case "GET /shares/" + shareID:
		f.requests = append(f.requests, key)
		if f.shareGone {
			notFound(w)
			return
		}
		reply(f.t, w, http.StatusOK, map[string]any{"share": map[string]any{"id": shareID, "share_proto": f.proto}})
	case "POST /shares/" + shareID + "/action":
		f.action(w, body)
	case "GET /share-access-rules/" + ruleID:
		f.requests = append(f.requests, key)
		f.get(w)
	case "GET /share-access-rules":
		f.requests = append(f.requests, key)
		f.list(w, req)
	default:
		f.t.Errorf("unexpected request %s", key)
		w.WriteHeader(http.StatusTeapot)
	}
}

func (f *fakeRules) action(w http.ResponseWriter, body []byte) {
	var action struct {
		Allow *struct {
			AccessType  string            `json:"access_type"`
			AccessTo    string            `json:"access_to"`
			AccessLevel string            `json:"access_level"`
			Metadata    map[string]string `json:"metadata"`
		} `json:"allow_access"`
		Deny *struct {
			AccessID string `json:"access_id"`
		} `json:"deny_access"`
	}
	require.NoError(f.t, json.Unmarshal(body, &action))
	switch {
	case action.Allow != nil:
		f.requests = append(f.requests, "POST allow_access")
		f.allowCalls++
		assert.Equal(f.t, "ip", action.Allow.AccessType)
		if f.failAllow != nil {
			w.WriteHeader(f.failAllow.status)
			_, _ = io.WriteString(w, f.failAllow.body)

			return
		}
		f.rule = map[string]any{
			"id": ruleID, "share_id": shareID, "access_type": action.Allow.AccessType,
			"access_to": action.Allow.AccessTo, "access_level": action.Allow.AccessLevel,
			"state": "queued_to_apply", "metadata": action.Allow.Metadata,
		}
		if f.lostAllow != nil {
			w.WriteHeader(f.lostAllow.status)
			_, _ = io.WriteString(w, f.lostAllow.body)

			return
		}
		reply(f.t, w, http.StatusOK, map[string]any{"access": f.rule})
	case action.Deny != nil:
		f.requests = append(f.requests, "POST deny_access")
		if f.shareGone || f.rule == nil || action.Deny.AccessID != ruleID {
			notFound(w)
			return
		}
		f.rule["state"] = "queued_to_deny"
		f.states = []string{"denying", "deleted"}
		w.WriteHeader(http.StatusAccepted)
	}
}

func (f *fakeRules) get(w http.ResponseWriter) {
	if f.failGet != nil {
		w.WriteHeader(f.failGet.status)
		_, _ = io.WriteString(w, f.failGet.body)

		return
	}
	if f.transientGets > 0 {
		f.transientGets--
		w.WriteHeader(unavailable503.status)
		_, _ = io.WriteString(w, unavailable503.body)

		return
	}
	if f.rule != nil && len(f.states) > 0 {
		f.rule["state"] = f.states[0]
		if len(f.states) > 1 {
			f.states = f.states[1:]
		}
	}
	if f.rule == nil || f.shareGone || f.rule["state"] == "deleted" {
		f.rule = nil
		notFound(w)

		return
	}
	reply(f.t, w, http.StatusOK, map[string]any{"access": f.rule})
}

func (f *fakeRules) list(w http.ResponseWriter, req *http.Request) {
	assert.Equal(f.t, shareID, req.URL.Query().Get("share_id"))
	var filter map[string]string
	require.NoError(f.t, json.Unmarshal([]byte(req.URL.Query().Get("metadata")), &filter))
	rules := []any{}
	for _, rule := range append([]map[string]any{f.rule}, f.foreign...) {
		if rule == nil {
			continue
		}
		metadata, _ := rule["metadata"].(map[string]string)
		if metadata[createTokenKey] == filter[createTokenKey] {
			rules = append(rules, rule)
		}
	}
	reply(f.t, w, http.StatusOK, map[string]any{"access_list": rules})
}

func reply(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(w).Encode(value))
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, `{"itemNotFound": {"code": 404, "message": "Not found."}}`)
}

func accessState() shareAccessModel {
	return shareAccessModel{
		ID:          types.StringValue(ruleID),
		ProjectID:   types.StringValue("project"),
		Region:      types.StringValue("ru-9"),
		ShareID:     types.StringValue(shareID),
		AccessTo:    types.StringValue("10.0.0.0/24"),
		AccessLevel: types.StringValue(levelRW),
		State:       types.StringValue(ruleActive),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(
			map[string]attr.Type{"create": types.StringType, "delete": types.StringType},
			map[string]attr.Value{"create": types.StringValue("2s"), "delete": types.StringValue("2s")},
		)},
	}
}

func accessPlan() shareAccessModel {
	m := accessState()
	m.ID = types.StringUnknown()
	m.State = types.StringUnknown()

	return m
}

func accessStateOf(t *testing.T, r *shareAccessResource, m shareAccessModel) tfsdk.State {
	t.Helper()
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	state := tfsdk.State{Schema: s.Schema}
	require.False(t, state.Set(context.Background(), &m).HasError())

	return state
}

func readAccessModel(t *testing.T, state tfsdk.State) shareAccessModel {
	t.Helper()
	var m shareAccessModel
	require.False(t, state.Get(context.Background(), &m).HasError())

	return m
}

func createAccess(t *testing.T, r *shareAccessResource, plan shareAccessModel) (shareAccessModel, bool, diag.Diagnostics) {
	t.Helper()
	empty := accessStateOf(t, r, accessState())
	empty.Raw = tftypes.NewValue(empty.Raw.Type(), nil)
	resp := resource.CreateResponse{State: empty}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan(accessStateOf(t, r, plan))}, &resp)
	if resp.State.Raw.IsNull() {
		return shareAccessModel{}, false, resp.Diagnostics
	}

	return readAccessModel(t, resp.State), true, resp.Diagnostics
}

func TestAccessTargetValidation(t *testing.T) {
	for value, valid := range map[string]bool{
		"10.0.0.10":      true,
		"10.0.0.0/24":    true,
		"0.0.0.0/0":      true,
		"192.168.5.0/24": true,
		"10.0.0.5/24":    false,
		"fd00::1":        false,
		"fd00::/64":      false,
		"storage.local":  false,
		"10.0.0.256":     false,
	} {
		t.Run(value, func(t *testing.T) {
			var resp validatorResponse
			accessTargetValidator().ValidateString(context.Background(), stringRequest(value), &resp.StringResponse)
			assert.Equal(t, !valid, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

func TestAccessCreate(t *testing.T) {
	t.Run("waits for active", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.states = []string{"applying", ruleActive}
		r := fake.resource()

		got, ok, diags := createAccess(t, r, accessPlan())

		require.False(t, diags.HasError(), "%v", diags)
		require.True(t, ok)
		assert.Equal(t, ruleID, got.ID.ValueString())
		assert.Equal(t, ruleActive, got.State.ValueString())
		assert.NotEmpty(t, fake.rule["metadata"].(map[string]string)[createTokenKey])
	})

	t.Run("a transient read failure does not end the wait", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.states = []string{"applying", ruleActive}
		fake.transientGets = 1
		r := fake.resource()

		got, ok, diags := createAccess(t, r, accessPlan())

		require.False(t, diags.HasError(), "%v", diags)
		require.True(t, ok)
		assert.Equal(t, ruleID, got.ID.ValueString())
		assert.Equal(t, ruleActive, got.State.ValueString())
		assert.Equal(t, 1, fake.allowCalls)
	})

	t.Run("read only CIFS is rejected before allow_access", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.proto = "CIFS"
		r := fake.resource()
		plan := accessPlan()
		plan.AccessLevel = types.StringValue(levelRO)

		_, ok, diags := createAccess(t, r, plan)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), `access_level = "rw"`)
		assert.False(t, ok)
		assert.Zero(t, fake.allowCalls)
	})

	t.Run("error state keeps the UUID", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.states = []string{"applying", ruleError}
		r := fake.resource()

		got, ok, diags := createAccess(t, r, accessPlan())

		require.True(t, diags.HasError())
		require.True(t, ok)
		assert.Equal(t, ruleID, got.ID.ValueString())
	})

	t.Run("a duplicate is rejected without a search", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.failAllow = &answer{
			http.StatusBadRequest,
			`{"badRequest": {"code": 400, "message": "Share access ip:10.0.0.0/24 exists."}}`,
		}
		fake.foreign = []map[string]any{{"id": "foreign", "access_to": "10.0.0.0/24", "metadata": map[string]string{}}}
		r := fake.resource()

		_, ok, diags := createAccess(t, r, accessPlan())

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "exists")
		assert.False(t, ok)
		assert.NotContains(t, fake.requests, "GET /share-access-rules")
	})

	for name, lost := range map[string]answer{
		"server error":     {http.StatusInternalServerError, ""},
		"rate limit":       {http.StatusTooManyRequests, `{"overLimit": {"code": 429, "message": "slow down"}}`},
		"unstructured 4xx": {http.StatusBadRequest, proxyPage},
	} {
		t.Run(name+" finds the rule by token", func(t *testing.T) {
			fake := newFakeRules(t)
			fake.lostAllow = &lost
			fake.states = []string{ruleActive}
			r := fake.resource()

			got, ok, diags := createAccess(t, r, accessPlan())

			require.False(t, diags.HasError(), "%v", diags)
			require.True(t, ok)
			assert.Equal(t, ruleID, got.ID.ValueString())
			assert.Equal(t, 1, fake.allowCalls)
		})
	}

	t.Run("a rule without the token is not adopted", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.failAllow = &answer{http.StatusBadGateway, proxyPage}
		fake.foreign = []map[string]any{{
			"id": "foreign", "access_to": "10.0.0.0/24", "access_level": "rw", "metadata": map[string]string{},
		}}
		r := fake.resource()

		got, ok, diags := createAccess(t, r, accessPlan())

		require.True(t, diags.HasError())
		require.True(t, ok)
		token, pending := strings.CutPrefix(got.ID.ValueString(), pendingPrefix)
		require.True(t, pending)
		detail := diags.Errors()[0].Detail()
		assert.Contains(t, detail, token)
		assert.Contains(t, detail, "terraform untaint")
		assert.Contains(t, detail, "terraform state rm")
		assert.Equal(t, 1, fake.allowCalls)
	})
}

func readAccess(t *testing.T, r *shareAccessResource, prior shareAccessModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	state := accessStateOf(t, r, prior)
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)

	return resp.State, resp.Diagnostics
}

func deleteAccess(t *testing.T, r *shareAccessResource, prior shareAccessModel) diag.Diagnostics {
	t.Helper()
	state := accessStateOf(t, r, prior)
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)

	return resp.Diagnostics
}

func existingRule(token string) map[string]any {
	return map[string]any{
		"id": ruleID, "share_id": shareID, "access_type": "ip", "access_to": "10.0.0.7",
		"access_level": "ro", "state": ruleActive, "metadata": map[string]string{createTokenKey: token},
	}
}

func TestAccessRead(t *testing.T) {
	t.Run("refreshes the rule", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.rule = existingRule("x")
		r := fake.resource()

		state, diags := readAccess(t, r, accessState())

		require.False(t, diags.HasError(), "%v", diags)
		got := readAccessModel(t, state)
		assert.Equal(t, "10.0.0.7", got.AccessTo.ValueString())
		assert.Equal(t, levelRO, got.AccessLevel.ValueString())
	})

	t.Run("a rule removed outside Terraform leaves the state", func(t *testing.T) {
		fake := newFakeRules(t)
		r := fake.resource()

		state, diags := readAccess(t, r, accessState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.True(t, state.Raw.IsNull())
	})

	for name, failure := range map[string]answer{
		"forbidden":    {http.StatusForbidden, `{"forbidden": {"code": 403, "message": "Policy does not allow."}}`},
		"server error": {http.StatusInternalServerError, ""},
	} {
		t.Run(name+" keeps the state", func(t *testing.T) {
			fake := newFakeRules(t)
			fake.failGet = &failure
			r := fake.resource()

			state, diags := readAccess(t, r, accessState())

			require.True(t, diags.HasError())
			assert.False(t, state.Raw.IsNull())
		})
	}

	t.Run("pending is replaced by the rule with the token", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.rule = existingRule("token-1")
		r := fake.resource()
		prior := accessState()
		prior.ID = types.StringValue(pendingPrefix + "token-1")

		state, diags := readAccess(t, r, prior)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, ruleID, readAccessModel(t, state).ID.ValueString())
	})

	t.Run("pending without the rule keeps the state", func(t *testing.T) {
		fake := newFakeRules(t)
		r := fake.resource()
		prior := accessState()
		prior.ID = types.StringValue(pendingPrefix + "token-1")

		state, diags := readAccess(t, r, prior)

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "token-1")
		assert.Equal(t, pendingPrefix+"token-1", readAccessModel(t, state).ID.ValueString())
	})
}

func TestAccessDelete(t *testing.T) {
	t.Run("revokes and waits", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.rule = existingRule("x")
		r := fake.resource()

		diags := deleteAccess(t, r, accessState())

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, []string{"POST deny_access"}, fake.writes())
		assert.Nil(t, fake.rule)
	})

	t.Run("a missing rule is revoked", func(t *testing.T) {
		fake := newFakeRules(t)
		r := fake.resource()

		require.False(t, deleteAccess(t, r, accessState()).HasError())
	})

	t.Run("a missing share is revoked", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.rule = existingRule("x")
		fake.shareGone = true
		r := fake.resource()

		require.False(t, deleteAccess(t, r, accessState()).HasError())
	})

	t.Run("pending revokes the rule found by token", func(t *testing.T) {
		fake := newFakeRules(t)
		fake.rule = existingRule("token-1")
		r := fake.resource()
		prior := accessState()
		prior.ID = types.StringValue(pendingPrefix + "token-1")

		diags := deleteAccess(t, r, prior)

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, []string{"POST deny_access"}, fake.writes())
	})
}

func TestAccessUpdateStoresTimeouts(t *testing.T) {
	r := &shareAccessResource{}
	prior := accessState()
	plan := prior
	plan.State = types.StringUnknown()
	plan.Timeouts = timeouts.Value{Object: types.ObjectValueMust(
		map[string]attr.Type{"create": types.StringType, "delete": types.StringType},
		map[string]attr.Value{"create": types.StringValue("15m"), "delete": types.StringValue("20m")},
	)}
	planned := tfsdk.Plan(accessStateOf(t, r, plan))
	resp := resource.UpdateResponse{State: tfsdk.State(planned)}

	r.Update(context.Background(), resource.UpdateRequest{State: accessStateOf(t, r, prior), Plan: planned}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	got := readAccessModel(t, resp.State)
	assert.Equal(t, ruleActive, got.State.ValueString())
	assert.Equal(t, plan.Timeouts, got.Timeouts)
}

func TestAccessImport(t *testing.T) {
	for name, tc := range map[string]struct {
		id, project, region, detail string
	}{
		"one part":    {ruleID, "project", "ru-9", "<share_id>/<access_id>"},
		"not UUIDs":   {"share/rule", "project", "ru-9", "<share_id>/<access_id>"},
		"no project":  {shareID + "/" + ruleID, "", "ru-9", "INFRA_PROJECT_ID"},
		"no region":   {shareID + "/" + ruleID, "project", "", "INFRA_REGION"},
		"valid scope": {shareID + "/" + ruleID, "project", "ru-9", ""},
	} {
		t.Run(name, func(t *testing.T) {
			r := &shareAccessResource{connection: connection{config: testConfig{project: tc.project, region: tc.region}}}
			empty := accessStateOf(t, r, accessState())
			empty.Raw = tftypes.NewValue(empty.Raw.Type(), nil)
			resp := resource.ImportStateResponse{State: empty}

			r.ImportState(context.Background(), resource.ImportStateRequest{ID: tc.id}, &resp)

			if tc.detail == "" {
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				got := readAccessModel(t, resp.State)
				assert.Equal(t, ruleID, got.ID.ValueString())
				assert.Equal(t, shareID, got.ShareID.ValueString())

				return
			}
			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.detail)
		})
	}
}
