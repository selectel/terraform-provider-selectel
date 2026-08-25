package convert

func ToStringSlice(v []any) []string {
	var s []string
	for _, val := range v {
		if strVal, ok := val.(string); ok {
			s = append(s, strVal)
		}
	}

	return s
}
