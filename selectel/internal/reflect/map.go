package reflect

import "reflect"

func IsSetContainsSubset(subset, set map[string]interface{}) bool {
	for k, subsetValue := range subset {
		setValue, ok := set[k]
		if !ok {
			return false
		}

		switch subsetValueTyped := subsetValue.(type) {
		case map[string]interface{}:
			setValueTyped, ok := setValue.(map[string]interface{})
			if !ok || !IsSetContainsSubset(subsetValueTyped, setValueTyped) {
				return false
			}

		case []interface{}:
			setValueTyped, ok := setValue.([]interface{})
			if !ok {
				return false
			}
			if !isArrayContainsSubarray(subsetValueTyped, setValueTyped) {
				return false
			}

		default:
			if !reflect.DeepEqual(subsetValue, setValue) {
				return false
			}
		}
	}

	return true
}

func isArrayContainsSubarray(subarray, array []interface{}) bool {
	for _, subarrayElement := range subarray {
		found := false
		for _, arrayElement := range array {
			switch subarrayElementTyped := subarrayElement.(type) {
			case map[string]interface{}:
				arrayElementTyped, ok := arrayElement.(map[string]interface{})
				if ok && IsSetContainsSubset(subarrayElementTyped, arrayElementTyped) {
					found = true
					break
				}

			case []interface{}:
				arrayElementTyped, ok := arrayElement.([]interface{})
				if ok && isArrayContainsSubarray(subarrayElementTyped, arrayElementTyped) {
					found = true
					break
				}

			default:
				if reflect.DeepEqual(subarrayElement, arrayElement) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}
