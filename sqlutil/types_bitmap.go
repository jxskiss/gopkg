package sqlutil

import (
	"database/sql"
	"database/sql/driver"
)

// Bitmap represents a bitmap value, it implements sql/driver.Valuer and sql.Scanner.
// Bitmap provides Get, Set and Clear methods to manipulate the bitmap value.
type Bitmap int

// Value implements driver.Valuer interface.
func (b Bitmap) Value() (driver.Value, error) {
	return int64(b), nil
}

// Scan implements sql.Scanner interface.
func (b *Bitmap) Scan(src interface{}) error {
	var tmp sql.NullInt64
	err := tmp.Scan(src)
	if err == nil {
		*b = Bitmap(tmp.Int64)
	}
	return err
}

func (b Bitmap) Get(mask int) bool {
	return int(b)&mask != 0
}

func (b *Bitmap) Set(mask int) {
	*b |= Bitmap(mask)
}

func (b *Bitmap) Clear(mask int) {
	*b &= ^Bitmap(mask)
}
