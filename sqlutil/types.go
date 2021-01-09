package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"github.com/jxskiss/gopkg/json"
	"github.com/jxskiss/gopkg/serialize"
)

// Bitmap represents a bitmap value, it implements sql/driver.Valuer and sql.Scanner.
// Bitmap provides Get, Set and Clear methods to manipulate the bitmap value.
type Bitmap int

func (b Bitmap) Value() (driver.Value, error) {
	return int64(b), nil
}

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

var (
	emptyArray  = []byte{'[', ']'}
	emptyObject = []byte{'{', '}'}
	zeroBytes   = []byte{}
)

// JSONInt32s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONInt32s []int32

func (p JSONInt32s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONInt32s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONInt64s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONInt64s []int64

func (p JSONInt64s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONInt64s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONStrings represents a string array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONStrings []string

func (p JSONStrings) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONStrings) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONStringMap represents a map[string]string value, it implements
// sql/driver.Valuer and sql.Scanner. It uses JSON to do serialization.
type JSONStringMap map[string]string

func (p JSONStringMap) Value() (driver.Value, error) {
	if p == nil {
		return emptyObject, nil
	}
	return json.Marshal(p)
}

func (p *JSONStringMap) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONDict represents a map[string]interface{} value, it implements
// sql/driver.Valuer and sql.Scanner. It uses JSON to do serialization.
type JSONDict map[string]interface{}

func (p JSONDict) Value() (driver.Value, error) {
	if p == nil {
		return emptyObject, nil
	}
	return json.Marshal(p)
}

func (p *JSONDict) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// PBInt32s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBInt32s []int32

func (p PBInt32s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.Int32List(p).MarshalProto()
}

func (p *PBInt32s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.Int32List)(p).UnmarshalProto(data)
	}
}

// PBInt64s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBInt64s []int64

func (p PBInt64s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.Int64List(p).MarshalProto()
}

func (p *PBInt64s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.Int64List)(p).UnmarshalProto(data)
	}
}

// PBStrings represents a string array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBStrings []string

func (p PBStrings) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.StringList(p).MarshalProto()
}

func (p *PBStrings) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.StringList)(p).UnmarshalProto(data)
	}
}

// PBStringMap represents a map[string]string value, it implements
// sql/driver.Valuer and sql.Scanner. It uses binary serialization format (protobuf).
type PBStringMap map[string]string

func (p PBStringMap) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.StringMap(p).MarshalProto()
}

func (p *PBStringMap) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.StringMap)(p).UnmarshalProto(data)
	}
}
