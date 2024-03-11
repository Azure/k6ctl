package stdlib

// Ptr returns a pointer to the value v.
func Ptr[T any](v T) *T {
	return &v
}

// ValOrZero returns the value of v if it's not nil, and a zero value otherwise.
func ValOrZero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
