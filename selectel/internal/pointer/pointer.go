package pointer

//go:fix inline
func To[T any](value T) *T {
	return new(value)
}
