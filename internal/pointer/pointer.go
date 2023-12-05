package pointer

// To returns a pointer to the given value. This is useful when setting a value to a struct field that is a pointer.
func To[T any](v T) *T {
	return &v
}
