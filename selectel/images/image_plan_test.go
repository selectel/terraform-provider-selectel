package images

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestSourceReplaced(t *testing.T) {
	url, file := types.StringValue("https://example.com/a.img"), types.StringValue("/tmp/a.img")
	null := types.StringNull()
	for name, tc := range map[string]struct {
		priorFile, priorURL, planFile, planURL types.String
		replaced                               bool
	}{
		"same source":                  {null, url, null, url, false},
		"url switched to file":         {null, url, file, null, true},
		"source set on imported image": {null, null, file, null, false},
		"source removed":               {file, null, null, null, false},
		"other file":                   {file, null, types.StringValue("/tmp/b.img"), null, true},
		"other url":                    {null, url, null, types.StringValue("https://example.com/b.img"), true},
		"file not known yet":           {file, null, types.StringUnknown(), null, true},
	} {
		t.Run(name, func(t *testing.T) {
			prior, plan := planned(), planned()
			prior.LocalFilePath, prior.ImageSourceURL = tc.priorFile, tc.priorURL
			plan.LocalFilePath, plan.ImageSourceURL = tc.planFile, tc.planURL
			assert.Equal(t, tc.replaced, sourceReplaced(prior, plan))
			assert.Equal(t, tc.replaced, hasCreationChange(prior, plan))
		})
	}
}

func TestValidatePropertyKeys(t *testing.T) {
	r := testImageResource(t, nil)
	config := planned()
	config.Properties = types.MapValueMust(types.StringType, map[string]attr.Value{
		"os_distro": types.StringValue("ubuntu"), "OS_Distro": types.StringValue("ubuntu"),
		"min_disk": types.StringValue("1"), "os_glance_x": types.StringValue("1"), "empty": types.StringNull(),
	})
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
		Config: tfsdk.Config(imageState(t, r, config)),
	}, &resp)

	summaries := make([]string, 0, len(resp.Diagnostics.Errors()))
	for _, d := range resp.Diagnostics.Errors() {
		summaries = append(summaries, d.Summary()+": "+d.Detail())
	}
	assert.ElementsMatch(t, []string{
		"Duplicate property key: The key differs from \"OS_Distro\" only in letter case.",
		"Reserved property key: The key is an image field. Set the min_disk_gb argument instead.",
		"Reserved property key: The os_glance prefix is reserved by the API.",
		"Null property value: Property values must be strings.",
	}, summaries)
}

func TestPlanRequiresImageData(t *testing.T) {
	r := testImageResource(t, nil)
	imported := planned()
	imported.ID = types.StringValue(imageID)
	for name, tc := range map[string]struct {
		create  bool
		change  func(*imageModel)
		missing bool
	}{
		"create":             {true, func(*imageModel) {}, true},
		"create from a file": {true, func(m *imageModel) { m.LocalFilePath = types.StringValue("/tmp/a.raw") }, false},
		"create from a URL":  {true, func(m *imageModel) { m.ImageSourceURL = types.StringValue("https://example.com/a.raw") }, false},
		// The source may come from another resource; apply checks it then.
		"create, source not known yet": {true, func(m *imageModel) { m.ImageSourceURL = types.StringUnknown() }, false},
		// An unknown project replaces the image, as RequiresReplace does.
		"replacement, project not known yet": {false, func(m *imageModel) { m.ProjectID = types.StringUnknown() }, true},
		"update of imported image":           {false, func(m *imageModel) { m.Name = types.StringValue("other") }, false},
		"replacement":                        {false, func(m *imageModel) { m.DiskFormat = types.StringValue("qcow2") }, true},
	} {
		t.Run(name, func(t *testing.T) {
			plan := imported
			tc.change(&plan)
			state := imageState(t, r, imported)
			if tc.create {
				state.Raw = tftypes.NewValue(state.Schema.Type().TerraformType(context.Background()), nil)
			}
			var resp resource.ModifyPlanResponse
			r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
				Plan: tfsdk.Plan(imageState(t, r, plan)), State: state,
			}, &resp)
			assert.Equal(t, tc.missing, resp.Diagnostics.HasError())
		})
	}
}
