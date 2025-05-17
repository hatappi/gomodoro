package conv

// ToPointer converts a value to a pointer.
func ToPointer[T any](v T) *T {
	return &v
}
