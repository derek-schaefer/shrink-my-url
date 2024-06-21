package shrinkmyurl

// Panics if the given error is not nil.
func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
