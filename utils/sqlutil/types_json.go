package sqlutil

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/json"
)

//nolint:unused
var (
	null        = []byte("null")
	emptyObject = []byte("{}")
)

// JSON holds a map[string]any value, it implements
// sql/driver.Valuer and sql.Scanner. It uses JSON to do serialization.
//
// JSON embeds a gemap.Map, thus all methods defined on gemap.Map is also
// available from a JSON instance.
type JSON struct {
	ezmap.Map
}

// Value implements driver.Valuer interface.
func (p JSON) Value() (driver.Value, error) {
	if p.Map == nil {
		return emptyObject, nil
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(p.Map)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Scan implements sql.Scanner interface.
func (p *JSON) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = unsafeheader.StringToBytes(v)
	default:
		return fmt.Errorf("sqlutil: wants []byte/string but got %T", src)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(&p.Map)
}

func (p JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Map)
}

func (p *JSON) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.Map)
}
