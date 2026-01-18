package config

// IsZero reports whether the provided configuration struct is empty.
// This can be used to detect missing or unset configuration.
func IsZero[T comparable](cfg T) bool {
	var zero T
	return cfg == zero
}
