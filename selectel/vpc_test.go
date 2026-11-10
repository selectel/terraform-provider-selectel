package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

const testVPCObject = "network"

func newTestAPIError(status int, class vpc.ErrorClass) error {
	return &vpc.APIError{
		StatusCode: status,
		Type:       "NeutronError",
		Message:    "test",
		Class:      class,
	}
}

func newTestTransportError() error {
	return &vpc.TransportError{Err: errors.New("dial tcp: i/o timeout")}
}

func newVPCTestData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
	d.SetId("2a1b3c4d-5e6f-7081-92a3-b4c5d6e7f809")

	return d
}

func TestVPCCheckDeleted(t *testing.T) {
	t.Run("not found on Read removes the resource from state without an error", func(t *testing.T) {
		d := newVPCTestData(t)

		diags := vpcCheckDeleted(d, newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound), testVPCObject)

		assert.False(t, diags.HasError())
		assert.Empty(t, d.Id())
	})

	for _, c := range []struct {
		name string
		err  error
	}{
		{name: "forbidden", err: newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden)},
		{name: "server error", err: newTestAPIError(http.StatusInternalServerError, vpc.ErrorClassServer)},
		{name: "transport failure", err: newTestTransportError()},
	} {
		t.Run(c.name+" on Read keeps the resource in state and reports an error", func(t *testing.T) {
			d := newVPCTestData(t)
			id := d.Id()

			diags := vpcCheckDeleted(d, c.err, testVPCObject)

			require.True(t, diags.HasError())
			assert.Equal(t, id, d.Id())
		})
	}

	t.Run("no error leaves the resource untouched", func(t *testing.T) {
		d := newVPCTestData(t)
		id := d.Id()

		diags := vpcCheckDeleted(d, nil, testVPCObject)

		assert.False(t, diags.HasError())
		assert.Equal(t, id, d.Id())
	})
}

func vpcV2StatusSequence(statuses ...string) vpcV2StatusRefreshFunc {
	i := 0

	return func(context.Context) (string, error) {
		if i >= len(statuses) {
			return statuses[len(statuses)-1], nil
		}
		status := statuses[i]
		i++

		return status, nil
	}
}

func TestWaitForVPCV2Status(t *testing.T) {
	const (
		interval = time.Millisecond
		timeout  = 50 * time.Millisecond
	)
	pending := []string{vpcV2StatusDown}

	t.Run("already settled", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusActive))
		require.NoError(t, err)
		assert.Equal(t, vpcV2StatusActive, status)
	})

	t.Run("pending then active", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, vpcV2StatusDown, vpcV2StatusActive))
		require.NoError(t, err)
		assert.Equal(t, vpcV2StatusActive, status)
	})

	t.Run("unknown status counts as settled", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, "N/A"))
		require.NoError(t, err)
		assert.Equal(t, "N/A", status)
	})

	t.Run("error status", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown, vpcV2StatusError))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "object id is in ERROR status")
		assert.Equal(t, vpcV2StatusError, status)
	})

	t.Run("timeout keeps last status", func(t *testing.T) {
		status, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, vpcV2StatusSequence(vpcV2StatusDown))
		require.ErrorIs(t, err, errVPCV2StatusTimeout)
		assert.Equal(t, vpcV2StatusDown, status)
	})

	t.Run("refresh error", func(t *testing.T) {
		refreshErr := errors.New("boom")
		_, err := waitForVPCV2Status(context.Background(), "object", "id", pending,
			interval, timeout, func(context.Context) (string, error) { return "", refreshErr })
		require.ErrorIs(t, err, refreshErr)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := waitForVPCV2Status(ctx, "object", "id", pending,
			time.Hour, time.Hour, vpcV2StatusSequence(vpcV2StatusDown))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func vpcTagsV2TestSet(tags ...string) *schema.Set {
	values := make([]any, len(tags))
	for i, tag := range tags {
		values[i] = tag
	}

	return schema.NewSet(schema.HashString, values)
}

func TestFlattenVPCTagsV2(t *testing.T) {
	t.Run("keeps tags that are present on the object", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
			"tags": []any{"managed"},
		})

		flattenVPCTagsV2(d, []string{"managed", "foreign"})

		assert.ElementsMatch(t, []any{"managed", "foreign"}, d.Get("all_tags").(*schema.Set).List())
		assert.ElementsMatch(t, []any{"managed"}, d.Get("tags").(*schema.Set).List())
	})

	t.Run("drops tags removed outside terraform", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
			"tags": []any{"managed", "removed-outside"},
		})

		flattenVPCTagsV2(d, []string{"managed", "foreign"})

		assert.ElementsMatch(t, []any{"managed"}, d.Get("tags").(*schema.Set).List())
	})
}

func TestExpandVPCUpdateTagsV2(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
		"tags": []any{"managed"},
	})
	d.Set("all_tags", []string{"foreign"})

	assert.ElementsMatch(t, []string{"managed", "foreign"}, expandVPCUpdateTagsV2(d))
}

type vpcTagsV2AccOperations interface {
	Get(ctx context.Context) ([]string, error)
	Add(ctx context.Context, tag string) error
}

func testAccCheckVPCTagsV2AddOutOfBand(
	operations func(*terraform.State) (vpcTagsV2AccOperations, error), tag string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ops, err := operations(s)
		if err != nil {
			return err
		}

		return ops.Add(context.Background(), tag)
	}
}

func testAccCheckVPCTagsV2Match(
	operations func(*terraform.State) (vpcTagsV2AccOperations, error), expected []string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ops, err := operations(s)
		if err != nil {
			return err
		}

		actual, err := ops.Get(context.Background())
		if err != nil {
			return err
		}

		actual = slices.Clone(actual)
		want := slices.Clone(expected)
		slices.Sort(actual)
		slices.Sort(want)

		if !slices.Equal(actual, want) {
			return fmt.Errorf("object has tags %v, want %v", actual, want)
		}

		return nil
	}
}

func TestMergeVPCTagsV2(t *testing.T) {
	t.Run("keeps foreign tags on managed tag change", func(t *testing.T) {
		merged := mergeVPCTagsV2(
			vpcTagsV2TestSet("one", "two", "foreign"),
			vpcTagsV2TestSet("one", "two"),
			vpcTagsV2TestSet("one", "three"),
		)

		assert.ElementsMatch(t, []string{"one", "three", "foreign"}, merged)
	})

	t.Run("keeps only foreign tags when all managed tags are removed", func(t *testing.T) {
		merged := mergeVPCTagsV2(
			vpcTagsV2TestSet("one", "foreign"),
			vpcTagsV2TestSet("one"),
			vpcTagsV2TestSet(),
		)

		assert.ElementsMatch(t, []string{"foreign"}, merged)
	})
}

func TestDeleteVPCV2WithRetry(t *testing.T) {
	const (
		interval = time.Millisecond
		timeout  = 50 * time.Millisecond
	)
	conflict := newTestAPIError(http.StatusConflict, vpc.ErrorClassConflict)
	notFound := newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound)

	t.Run("deleted at once", func(t *testing.T) {
		calls := 0
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, vpcV2ConflictIsTransient,
			func(context.Context) error { calls++; return nil })
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("already gone counts as deleted", func(t *testing.T) {
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, vpcV2ConflictIsTransient,
			func(context.Context) error { return notFound })
		require.NoError(t, err)
	})

	t.Run("conflict is retried until deleted", func(t *testing.T) {
		calls := 0
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, vpcV2ConflictIsTransient,
			func(context.Context) error {
				calls++
				if calls < 3 {
					return conflict
				}

				return nil
			})
		require.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("conflict until timeout", func(t *testing.T) {
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, vpcV2ConflictIsTransient,
			func(context.Context) error { return conflict })
		require.ErrorIs(t, err, errVPCV2DeleteTimeout)
		require.ErrorIs(t, err, conflict)
		assert.Contains(t, err.Error(), "object id is still in use after 50ms")
	})

	t.Run("other error is returned at once", func(t *testing.T) {
		calls := 0
		forbidden := newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden)
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, vpcV2ConflictIsTransient,
			func(context.Context) error { calls++; return forbidden })
		require.ErrorIs(t, err, forbidden)
		assert.Equal(t, 1, calls)
	})

	t.Run("conflict the predicate calls permanent is returned at once", func(t *testing.T) {
		calls := 0
		never := func(error) bool { return false }
		err := deleteVPCV2WithRetry(context.Background(), "object", "id", interval, timeout, never,
			func(context.Context) error { calls++; return conflict })
		require.ErrorIs(t, err, conflict)
		require.NotErrorIs(t, err, errVPCV2DeleteTimeout)
		assert.Equal(t, 1, calls)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := deleteVPCV2WithRetry(ctx, "object", "id", time.Hour, time.Hour, vpcV2ConflictIsTransient,
			func(context.Context) error { return conflict })
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestWaitForVPCFloatingIPV2Active(t *testing.T) {
	newClient := func(t *testing.T, statuses ...string) *vpc.Client {
		t.Helper()
		i := 0
		client, err := vpc.NewClient(vpc.Config{
			Endpoint: "https://vpc.example.test",
			Token:    "token",
			HTTPClient: &http.Client{
				Transport: httptest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "/v2.0/floatingips/fip-1", req.URL.Path)
					status := statuses[min(i, len(statuses)-1)]
					i++

					return httptest.NewFakeResponse(http.StatusOK,
						fmt.Sprintf(`{"floatingip": {"id": "fip-1", "status": %q}}`, status)), nil
				}),
			},
		})
		require.NoError(t, err)

		return client
	}

	t.Run("active", func(t *testing.T) {
		diags := waitForVPCFloatingIPV2Active(context.Background(), newClient(t, vpcV2StatusDown, vpcV2StatusActive),
			"fip-1", time.Millisecond, 50*time.Millisecond)
		assert.Empty(t, diags)
	})

	t.Run("still down after timeout is a warning", func(t *testing.T) {
		diags := waitForVPCFloatingIPV2Active(context.Background(), newClient(t, vpcV2StatusDown),
			"fip-1", time.Millisecond, 20*time.Millisecond)
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "fip-1 is still DOWN")
	})

	t.Run("error status", func(t *testing.T) {
		diags := waitForVPCFloatingIPV2Active(context.Background(), newClient(t, vpcV2StatusError),
			"fip-1", time.Millisecond, 20*time.Millisecond)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, "fip-1 is in ERROR status")
	})
}

func TestVPCV2ConflictIsTransient(t *testing.T) {
	assert.True(t, vpcV2ConflictIsTransient(newTestAPIError(http.StatusConflict, vpc.ErrorClassConflict)))
	assert.False(t, vpcV2ConflictIsTransient(newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound)))
	assert.False(t, vpcV2ConflictIsTransient(newTestTransportError()))
}

func planVPCV2ResourceDiff(t *testing.T, r *schema.Resource, values map[string]cty.Value) error {
	t.Helper()

	block := r.CoreConfigSchema()
	configValue, err := block.CoerceValue(cty.ObjectVal(values))
	require.NoError(t, err)

	// Terraform passes the raw configuration through the prior state, also on create.
	state := &terraform.InstanceState{RawConfig: configValue}
	_, err = r.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(configValue, block), nil)

	return err
}
