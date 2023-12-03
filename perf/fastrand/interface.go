package fastrand

type randInterface interface {
	Uint64() uint64
	Float64() float64
}

type globalImpl struct{}

func (globalImpl) Uint64() uint64   { return Uint64() }
func (globalImpl) Float64() float64 { return Float64() }
