package sqlutil

import (
	"database/sql"
	"database/sql/driver"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// Bitmap represents a bitmap value, it implements sql/driver.Valuer and sql.Scanner.
// Bitmap provides Get, Set and Clear methods to manipulate the bitmap value.
type Bitmap[T constraints.Integer] struct {
	val *T
}

// NewBitmap returns a new bitmap value.
func NewBitmap[T constraints.Integer](val *T) Bitmap[T] {
	return Bitmap[T]{val: val}
}

// Value implements driver.Valuer interface.
func (b Bitmap[T]) Value() (driver.Value, error) {
	var out int64
	if b.val != nil {
		out = int64(*b.val)
	}
	return out, nil
}

// Scan implements sql.Scanner interface.
func (b *Bitmap[T]) Scan(src interface{}) error {
	var tmp sql.NullInt64
	err := tmp.Scan(src)
	if err == nil {
		val := T(tmp.Int64)
		b.val = &val
	}
	return err
}

// Get returns whether mask is set in the bitmap.
func (b Bitmap[T]) Get(mask int) bool {
	if b.val == nil {
		return false
	}
	return *b.val&T(mask) != 0
}

// Set sets mask to the bitmap.
func (b *Bitmap[T]) Set(mask int) {
	if b.val == nil {
		var zero T
		b.val = &zero
	}
	*b.val |= T(mask)
}

// Clear clears mask from the bitmap.
func (b *Bitmap[T]) Clear(mask int) {
	if b.val != nil {
		*b.val &= ^T(mask)
	}
}

// Underlying returns the underlying integer value.
func (b Bitmap[T]) Underlying() T {
	if b.val != nil {
		return *b.val
	}
	return 0
}
