// ptr provides functions to convert from/to pointers.
package ptr

// To converts a value to pointer.
func To[T any](v T) *T {
	return &v
}

// From dereferences a pointer.
func From[T any](v *T) T {
	return *v
}
