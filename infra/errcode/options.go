package errcode

func (p *Registry) applyOptions(options ...Option) {
	for _, o := range options {
		if o.applyRegistry != nil {
			o.applyRegistry(p)
		}
	}
}

// An Option customizes the behavior of Registry.
type Option struct {
	applyRegistry func(*Registry)
}

// WithReserved returns an option to make a Registry to reserve some codes.
// Calling Register with a reserved code causes a panic.
// Reserved code can be registered by calling RegisterReserved.
func WithReserved(fn func(code int32) bool) Option {
	return Option{
		applyRegistry: func(r *Registry) {
			r.reserve = fn
		},
	}
}
