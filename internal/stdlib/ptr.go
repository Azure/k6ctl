package stdlib

// Ptr returns a pointer to the value v.
func Ptr[T any](v T) *T {
	return &v
}
