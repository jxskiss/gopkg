package ptr

// Copy returns a shallow copy of the given pointer of any type.
// If p is nil, it returns nil.
func Copy[T any](p *T) (ret *T) {
	if p != nil {
		x := *p
		ret = &x
	}
	return
}

// Deref returns the value pointed by pointer p.
// If p is nil, it returns zero value of type T.
func Deref[T any](p *T) (ret T) {
	if p != nil {
		ret = *p
	}
	return
}

// Ptr returns copies v of any type and returns a pointer.
func Ptr[T any](v T) *T {
	return &v
}

// NotZero returns a pointer to v if v is not zero value of its type,
// else it returns nil.
func NotZero[T comparable](v T) (ret *T) {
	var zero T
	if v != zero {
		ret = &v
	}
	return
}
