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

// FromDef dereferences a pointer.
// It returns def if the pointer is nil.
func FromDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}

	return *v
}
