package serialize

import (
	"encoding/json"
	"math/rand"
	"testing"
)

func genRandomIntegers() []int64 {
	var out []int64
	for i := 0; i < 10; i++ {
		out = append(out, rand.Int63())
	}
	return out
}

func genRandomStrings() map[string]string {
	return map[string]string{
		"abc":         "1234511",
		"some file":   "中文",
		"some escape": "some \" escape \n \t \fthing",
	}
}

func BenchmarkInt64sMarshalJSON(b *testing.B) {
	var data = genRandomIntegers()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

func BenchmarkInt64sMarshalBinary(b *testing.B) {
	var data = genRandomIntegers()
	_x := Int64List(data)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = _x.MarshalBinary()
	}
}

func BenchmarkInt64sMarshalProtobuf(b *testing.B) {
	var data = genRandomIntegers()
	_x := Int64List(data)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = _x.MarshalProto()
	}
}

func BenchmarkStringMapMarshalJSON(b *testing.B) {
	var data = genRandomStrings()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

func BenchmarkStringMapMarshalProtobuf(b *testing.B) {
	var data = genRandomStrings()
	_x := StringMap(data)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = _x.MarshalProto()
	}
}
