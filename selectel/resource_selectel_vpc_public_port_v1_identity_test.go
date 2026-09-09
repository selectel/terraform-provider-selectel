package selectel

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceVPCPublicPortV1ImportStateByIdentity(t *testing.T) {
	r := resourceVPCPublicPortV1()
	require.NotNil(t, r.Identity, "resource must declare an identity schema")

	d := schema.TestResourceDataWithIdentityRaw(t, r.SchemaMap(), r.Identity.SchemaMap(), map[string]string{
		"id":         "b311ce58-2658-46b5-b733-7a0f418703f2",
		"project_id": "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		"region":     "ru-6",
	})

	result, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{})
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", result[0].Id())
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", result[0].Get("project_id"))
	assert.Equal(t, "ru-6", result[0].Get("region"))
}

func TestResourceVPCPublicPortV1ImportStateByID(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")

	result, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{
		ProjectID: "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		Region:    "ru-6",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", result[0].Id())
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", result[0].Get("project_id"))
	assert.Equal(t, "ru-6", result[0].Get("region"))
}

func TestResourceVPCPublicPortV1ImportStateByIDWithoutProjectScope(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")

	_, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{})
	assert.Error(t, err)
}

func TestFillVPCPublicPortDataSetsIdentity(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")
	require.NoError(t, d.Set("region", "ru-6"))

	description := "port"
	fillVPCPublicPortData(&publicnetapi.Port{
		ID:          "b311ce58-2658-46b5-b733-7a0f418703f2",
		ProjectID:   "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		Description: &description,
	}, d)

	identity, err := d.Identity()
	require.NoError(t, err)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", identity.Get("id"))
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", identity.Get("project_id"))
	assert.Equal(t, "ru-6", identity.Get("region"))
}
