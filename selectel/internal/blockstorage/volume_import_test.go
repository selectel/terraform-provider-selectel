package blockstorage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	for _, scenario := range []string{"valid", "invalid id", "missing project", "missing region"} {
		t.Run(scenario, func(t *testing.T) {
			r := testResource(nil)
			id := volumeID
			switch scenario {
			case "invalid id":
				id = "bad"
			case "missing project":
				r.config = testConfig{region: "ru-1"}
			case "missing region":
				r.config = testConfig{project: "project"}
			}
			response := resource.ImportStateResponse{State: tfsdk.State{Schema: stateOf(t, r, model()).Schema}}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &response)
			if scenario != "valid" {
				require.True(t, response.Diagnostics.HasError())
				return
			}
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			state := readModel(t, response.State)
			assert.Equal(t, volumeID, state.ID.ValueString())
			assert.Equal(t, "project", state.ProjectID.ValueString())
			assert.Equal(t, "ru-1", state.Region.ValueString())
			assert.True(t, state.ImageID.IsNull())
			assert.False(t, state.OnlineResize.ValueBool())
		})
	}
}
