package ptr

func Copy[T any](p *T) *T {
	if p == nil {
		return nil
	}
	x := *p
	return &x
}

func Deref[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

func Ptr[T any](v T) *T {
	return &v
}
