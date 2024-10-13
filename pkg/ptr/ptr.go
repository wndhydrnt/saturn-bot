package ptr

func To[T any](v T) *T {
	return &v
}

func From[T any](v *T) T {
	return *v
}
