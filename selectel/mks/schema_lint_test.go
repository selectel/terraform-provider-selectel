package mks

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The cluster resource is not registered in the provider yet, so
// `make schemadoc` cannot lint it; this test keeps the schema lint-clean.
func TestClusterSchemaLint(t *testing.T) {
	r := &clusterResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	assert.Empty(t, schemadoc.Lint(schema.Schema{
		Attributes: schemaResp.Schema.Attributes, Blocks: schemaResp.Schema.Blocks,
	}))
	require.False(t, schemaResp.Diagnostics.HasError())
}
