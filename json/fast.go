package json

import (
	"fmt"
	"github.com/valyala/bytebufferpool"
)

var pool bytebufferpool.Pool

func MarshalStringMapUnordered(strMap map[string]string) ([]byte, error) {
	buf := pool.Get()
	defer pool.Put(buf)

	var err error
	buf.B, err = AppendStringMapUnordered(buf.B, strMap)
	if err != nil {
		return nil, err
	}
	out := make([]byte, buf.Len())
	copy(out, buf.B)
	return out, nil
}

func AppendStringMapUnordered(buf []byte, strMap map[string]string) ([]byte, error) {
	if strMap == nil {
		return append(buf, nullJSON...), nil
	}
	size := len(strMap)
	if size == 0 {
		return append(buf, emptyObject...), nil
	}
	idx := 0
	buf = append(buf, leftWING)
	for k, v := range strMap {
		buf, _ = appendString(buf, k)
		buf = append(buf, colon)
		buf, _ = appendString(buf, v)
		if idx++; idx < size {
			buf = append(buf, comma)
		}
	}
	buf = append(buf, rightWING)
	return buf, nil
}

func UnmarshalStringMap(data []byte, dst *map[string]string) error {
	size := len(data)
	buf := make([]byte, size)
	copy(buf, data)

	var lastIdx = size - 1
	var idx = 0
	c, idx, err := nextToken(buf, idx, lastIdx)
	if err != nil {
		return err
	}
	var isNull bool
	if idx, isNull = checkNull(c, buf, idx, lastIdx); isNull {
		*dst = nil
		return nil
	}
	if c != leftWING || buf[lastIdx] != rightWING {
		return fmt.Errorf("json: UnmarshalStringMap: invalid json string")
	}

	*dst = make(map[string]string)
	if ch, _, _ := nextToken(buf, idx, lastIdx); ch == rightWING {
		return nil
	}
	for ; c == comma || c == leftWING; c, idx, err = nextToken(buf, idx, lastIdx) {
		if err != nil {
			return fmt.Errorf("json: UnmarshalStringMap: %v", err)
		}
		var key, val string
		key, idx, err = readString(buf, idx, lastIdx)
		if err != nil {
			return fmt.Errorf("json: UnmarshalStringMap: %v", err)
		}
		c, idx, err = nextToken(buf, idx, lastIdx)
		if err != nil {
			return fmt.Errorf("json: UnmarshalStringMap: %v", err)
		}
		if c != ':' {
			err := "expects ':' after object field, but found " + string(c)
			return fmt.Errorf("json: UnmarshalStringMap: %v", err)
		}
		val, idx, err = readString(buf, idx, lastIdx)
		if err != nil {
			return fmt.Errorf("json: UnmarshalStringMap: %v", err)
		}
		(*dst)[key] = val
	}
	return nil
}
