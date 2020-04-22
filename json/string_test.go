package json

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var testStringMap = map[string]string{
	"id":         "1234567",
	"first_name": "Jeanette",
	"last_name":  "Penddreth",
	"email":      "jpenddreth0@census.gov",
	"gender":     "Female",
	"ip_address": "26.58.193.2",
	"html_tag":   "<html></html>",
	"chinese":    "北京欢迎你！",
	`a:\b":"\"c`: `d\"e:f`,
}

func TestMarshalStringMap(t *testing.T) {
	strMap := testStringMap

	_, err := json.Marshal(strMap)
	require.Nil(t, err)

	buf2, err := MarshalStringMap(strMap)
	require.Nil(t, err)
	var got map[string]string
	err = json.Unmarshal(buf2, &got)
	require.Nil(t, err)
	assert.Equal(t, strMap, got)
}

func BenchmarkMarshalStringMap(b *testing.B) {
	strMap := testStringMap
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalStringMap(strMap)
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
