package reflect

import (
	"reflect"
	"strings"
)

func IsSetContainsSubset(subset map[string]interface{}, set any) bool {
	return match(subset, reflect.ValueOf(set))
}

func match(subset map[string]interface{}, val reflect.Value) bool {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for k, subsetValue := range subset {
		var fieldValue reflect.Value

		switch val.Kind() {
		case reflect.Map:
			fieldValue = val.MapIndex(reflect.ValueOf(k))
			if !fieldValue.IsValid() {
				return false
			}

		case reflect.Struct:
			fieldValue = val.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, k)
			})
			if !fieldValue.IsValid() {
				return false
			}

		default:
			return false
		}

		if !matchValue(subsetValue, fieldValue.Interface()) {
			return false
		}
	}

	return true
}

func matchValue(subsetValue any, setValue any) bool {
	switch subsetValueTyped := subsetValue.(type) {
	case map[string]interface{}:
		return IsSetContainsSubset(subsetValueTyped, setValue)

	case []interface{}:
		setSlice, ok := toSlice(setValue)
		if !ok {
			return false
		}

		return isArrayContainsSubarray(subsetValueTyped, setSlice)

	default:
		return reflect.DeepEqual(subsetValue, setValue)
	}
}

func toSlice(v any) ([]interface{}, bool) {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, false
	}

	var result []interface{}
	for i := 0; i < val.Len(); i++ {
		result = append(result, val.Index(i).Interface())
	}

	return result, true
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
