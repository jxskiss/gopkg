package logid

import (
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestV2Gen(t *testing.T) {
	testCases := []struct {
		name       string
		useUTC     bool
		decodeFunc func(string) Info
	}{
		{"use Local", false, Decode},
		{"use UTC", true, Decode},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testIP := "10.1.2.3"
			testV2Gen := NewV2Gen(net.ParseIP(testIP))
			if tc.useUTC {
				testV2Gen.UseUTC()
			}

			now := time.Now().UnixMilli()

			gotLogId := testV2Gen.Gen()
			assert.Len(t, gotLogId, v2IPv4Length)

			info := tc.decodeFunc(gotLogId)
			assert.True(t, info.Valid())
			assert.Equal(t, byte(v2Version), info.Version())
			assert.True(t, math.Abs(float64(info.(V2Info).Time().UnixMilli()-now)) <= 1)
			assert.Equal(t, testIP, info.(V2Info).IP().String())

			gotLogId2 := testV2Gen.Gen()
			info2 := tc.decodeFunc(gotLogId2)
			assert.True(t, info2.Valid())
			assert.NotEqual(t, info.(V2Info).Random(), info2.(V2Info).Random())

			info3 := tc.decodeFunc("2DAOLKWQALEKRALK")
			assert.False(t, info3.Valid())
			assert.Equal(t, "0|invalid", info3.String())
		})
	}
}

func BenchmarkV2Gen(b *testing.B) {
	gen := NewV2Gen(nil)
	id2 := gen.Gen()

	b.Run("generate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = gen.Gen()
		}
	})

	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Decode(id2)
		}
	})
}
