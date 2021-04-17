package json

import (
	"fmt"
	"github.com/jxskiss/gopkg/bbp"
)

var pool bbp.Pool

func MarshalStringMapUnordered(strMap map[string]string) ([]byte, error) {
	buf := pool.Get()
	defer pool.Put(buf)

	buf.B = appendStringMapUnordered(buf.B, strMap)
	return buf.Copy(), nil
}

func AppendStringMapUnordered(buf []byte, strMap map[string]string) ([]byte, error) {
	buf = appendStringMapUnordered(buf, strMap)
	return buf, nil
}

func appendStringMapUnordered(buf []byte, strMap map[string]string) []byte {
	if strMap == nil {
		return append(buf, nullJSON...)
	}
	size := len(strMap)
	if size == 0 {
		return append(buf, emptyObject...)
	}
	idx := 0
	buf = append(buf, leftWING)
	for k, v := range strMap {
		buf = grow(buf, 2+len(k)+len(v))
		if idx++; idx > 1 {
			buf = append(buf, comma)
		}
		buf, _ = appendString(buf, k)
		buf = append(buf, colon)
		buf, _ = appendString(buf, v)
	}
	buf = append(buf, rightWING)
	return buf
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

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(buf).
func grow(buf []byte, n int) []byte {
	c, l := cap(buf), len(buf)
	if c-l >= n {
		return buf
	}
	c1 := 2 * c
	for x := l + n; c1 < x; {
		c1 *= 2
	}
	newbuf := make([]byte, l, c1)
	copy(newbuf, buf)
	return newbuf
}
