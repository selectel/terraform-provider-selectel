package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
