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
		_, _ = Marshal(strMap)
	}
}

func BenchmarkJSONIterMarshal(b *testing.B) {
	strMap := testStringMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cfg.Marshal(strMap)
	}
}

func BenchmarkStdJSONMarshal(b *testing.B) {
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
		_, _ = Marshal(strMap)
	}
}

func BenchmarkJSONIterStringInterfaceMap(b *testing.B) {
	strMap := testStringInterfaceMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cfg.Marshal(strMap)
	}
}

func BenchmarkStdJSONStringInterfaceMap(b *testing.B) {
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

func BenchmarkJSONIterUnmarshal(b *testing.B) {
	strMap := testStringMap
	data, _ := json.Marshal(strMap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tmp map[string]string
		_ = cfg.Unmarshal(data, &tmp)
	}
}

func BenchmarkStdJSONUnmarshal(b *testing.B) {
	strMap := testStringMap
	data, _ := json.Marshal(strMap)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tmp map[string]string
		_ = json.Unmarshal(data, &tmp)
	}
}
