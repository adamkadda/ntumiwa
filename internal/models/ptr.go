package models

// Helpful for quickly obtaining the address to anything.
// It's kinda annoying that this is how we have to do it.
func ptr[T any](v T) *T { return &v }
