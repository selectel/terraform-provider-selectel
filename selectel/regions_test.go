package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

const (
	testRu3Region = "ru-3"
)

func TestExpandVPCV2Regions(t *testing.T) {
	r := resourceVPCKeypairV2()
	d := r.TestResourceData()
	d.SetId("1")
	regions := []interface{}{"ru-1", "ru-2", "ru-3"}
	d.Set("regions", regions)

	expected := []string{"ru-1", "ru-2", "ru-3"}

	actual := expandVPCV2Regions(d.Get("regions").(*schema.Set))

	assert.ElementsMatch(t, expected, actual)
}
