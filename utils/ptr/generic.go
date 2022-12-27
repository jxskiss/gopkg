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
