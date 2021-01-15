package json

import (
	"encoding/json"
	"testing"
)

func BenchmarkMarshalStringMap(b *testing.B) {
	strMap := testStringMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalFast(strMap)
	}
}

func BenchmarkMarshalStringMap_JSONIter(b *testing.B) {
	strMap := testStringMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stdcfg.Marshal(strMap)
	}
}

func BenchmarkMarshalStringMap_Standard(b *testing.B) {
	strMap := testStringMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(strMap)
	}
}

func BenchmarkMarshalStringInterfaceMap(b *testing.B) {
	strMap := testStringInterfaceMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalFast(strMap)
	}
}

func BenchmarkMarshalStringInterfaceMap_JSONIter(b *testing.B) {
	strMap := testStringInterfaceMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stdcfg.Marshal(strMap)
	}
}

func BenchmarkMarshalStringInterfaceMap_Standard(b *testing.B) {
	strMap := testStringInterfaceMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(strMap)
	}
}

func BenchmarkUnmarshalStringMap(b *testing.B) {
	strMap := testStringMap
	data, _ := json.Marshal(strMap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tmp map[string]string
		_ = Unmarshal(data, &tmp)
	}
}

func BenchmarkUnmarshalStringMap_JSONIter(b *testing.B) {
	strMap := testStringMap
	data, _ := json.Marshal(strMap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tmp map[string]string
		_ = stdcfg.Unmarshal(data, &tmp)
	}
}

func BenchmarkUnmarshalStringMap_Standard(b *testing.B) {
	strMap := testStringMap
	data, _ := json.Marshal(strMap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tmp map[string]string
		_ = json.Unmarshal(data, &tmp)
	}
}
