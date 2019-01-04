package selvpc

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestExpandResellV2Regions(t *testing.T) {
	r := resourceResellKeypairV2()
	d := r.TestResourceData()
	d.SetId("1")
	regions := []interface{}{"ru-1", "ru-2", "ru-3"}
	d.Set("regions", regions)

	expected := []string{"ru-1", "ru-2", "ru-3"}

	actual := expandResellV2Regions(d.Get("regions").(*schema.Set))

	assert.ElementsMatch(t, expected, actual)
}
