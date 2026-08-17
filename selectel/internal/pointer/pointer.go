package pointer

func To[T any](value T) *T {
	return &value
}
