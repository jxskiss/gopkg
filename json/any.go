package json

import (
	"github.com/json-iterator/go"
	"strconv"
	"strings"
)

type Any = jsoniter.Any

func Get(data []byte, path ...interface{}) Any {
	return stdcfg.Get(data, path...)
}

func GetByDot(data []byte, path string) Any {
	return stdcfg.Get(data, splitDotPath(path)...)
}

func splitDotPath(path string) []interface{} {
	parts := strings.Split(path, ".")
	out := make([]interface{}, 0, len(parts))
	for _, s := range parts {
		switch {
		case isDigits(s):
			idx, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, int(idx))
		case s == "*":
			out = append(out, '*')
		default:
			out = append(out, s)
		}
	}
	return out
}

func isDigits(s string) bool {
	for _, x := range s {
		if !('0' <= x && x <= '9') {
			return false
		}
	}
	return true
}
