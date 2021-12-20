package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	featureGatesKey         = "feature_gates"
	admissionControllersKey = "admission_controllers"
)

func getSetAsStrings(d *schema.ResourceData, key string) ([]string, error) {
	val, ok := d.GetOk(key)
	if !ok {
		return []string{}, nil
	}

	set, ok := val.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("%q is not 'Set' at schema", key)
	}

	list := set.List()
	result := make([]string, len(list))
	for i, item := range list {
		val, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%q item '%v' is not a string", key, item)
		}
		result[i] = val
	}

	return result, nil
}
