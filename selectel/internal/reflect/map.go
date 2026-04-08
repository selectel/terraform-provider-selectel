package reflect

import "reflect"

func IsSetContainsSubset(subset, set map[string]any) bool {
	for k, subsetValue := range subset {
		setValue, ok := set[k]
		if !ok {
			return false
		}

		switch subsetValueTyped := subsetValue.(type) {
		case map[string]any:
			setValueTyped, ok := setValue.(map[string]any)
			if !ok || !IsSetContainsSubset(subsetValueTyped, setValueTyped) {
				return false
			}

		case []any:
			setValueTyped, ok := setValue.([]any)
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

func isArrayContainsSubarray(subarray, array []any) bool {
	for _, subarrayElement := range subarray {
		found := false
		for _, arrayElement := range array {
			switch subarrayElementTyped := subarrayElement.(type) {
			case map[string]any:
				arrayElementTyped, ok := arrayElement.(map[string]any)
				if ok && IsSetContainsSubset(subarrayElementTyped, arrayElementTyped) {
					found = true
					break
				}

			case []any:
				arrayElementTyped, ok := arrayElement.([]any)
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
