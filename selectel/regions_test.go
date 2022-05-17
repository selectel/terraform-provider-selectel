package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
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

func TestValidateRegionOk(t *testing.T) {
	validRegions := []string{
		ru1Region,
		ru2Region,
		ru3Region,
		ru7Region,
		ru8Region,
		ru9Region,
		uz1Region,
		nl1Region,
	}

	for _, region := range validRegions {
		err := validateRegion(region)
		assert.NoError(t, err)
	}
}

func TestValidateRegionErr(t *testing.T) {
	region := "unknown region"

	expected := "region is invalid: unknown region"
	actual := validateRegion(region)

	assert.Error(t, actual)
	assert.EqualError(t, actual, expected)
}
