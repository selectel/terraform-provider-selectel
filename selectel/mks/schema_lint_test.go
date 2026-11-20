package mks

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

func TestClusterSchemaLint(t *testing.T) {
	r := &clusterResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	assert.Empty(t, schemadoc.Lint(schema.Schema{
		Attributes: schemaResp.Schema.Attributes, Blocks: schemaResp.Schema.Blocks,
	}))
	require.False(t, schemaResp.Diagnostics.HasError())
}
