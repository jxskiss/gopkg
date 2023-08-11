package heapx

import (
	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

type Ordered interface {
	constraints.Integer | ~string
}

type MinImpl[T Ordered] []T

func (x MinImpl[T]) Len() int {
	return len(x)
}

func (x MinImpl[T]) Less(i, j int) bool {
	return x[i] < x[j]
}

func (x MinImpl[T]) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x *MinImpl[T]) Push(v any) {
	*x = append(*x, v.(T))
}

func (x *MinImpl[T]) Pop() any {
	old := *x
	item := old[len(old)-1]
	*x = old[:len(old)-1]
	return item
}

type MaxImpl[T Ordered] []T

func (x MaxImpl[T]) Len() int {
	return len(x)
}

func (x MaxImpl[T]) Less(i, j int) bool {
	return x[i] > x[j]
}

func (x MaxImpl[T]) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x *MaxImpl[T]) Push(v any) {
	*x = append(*x, v.(T))
}

func (x *MaxImpl[T]) Pop() any {
	old := *x
	n := len(old)
	item := old[n-1]
	*x = old[:n-1]
	return item
}
