package logid

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testV1Gen = NewV1Gen()

func TestV1Gen(t *testing.T) {
	now := time.Now().UnixMilli()
	machineID := string(testV1Gen.(*v1Gen).machineID[:])

	gotLogId := testV1Gen.Gen()
	assert.Len(t, gotLogId, v1Length)

	info := Decode(gotLogId)
	assert.True(t, info.Valid())
	assert.Equal(t, byte(v1Version), info.Version())
	assert.True(t, math.Abs(float64(info.(V1Info).Time().UnixMilli()-now)) <= 1)
	assert.Equal(t, machineID, info.(V1Info).MachineID())
	assert.Contains(t, info.String(), machineID)

	gotLogId2 := testV1Gen.Gen()
	info2 := Decode(gotLogId2)
	assert.True(t, info2.Valid())
	assert.NotEqual(t, info.(V1Info).Random(), info2.(V1Info).Random())
	assert.Equal(t, info.(*v1Info).machineID, info2.(*v1Info).machineID)

	info3 := Decode("1DAOLKWQALEKRALK")
	assert.False(t, info3.Valid())
	assert.Equal(t, "0|invalid", info3.String())
}

func BenchmarkV1Gen(b *testing.B) {
	gen := NewV1Gen()
	id1 := gen.Gen()

	b.Run("generate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = gen.Gen()
		}
	})

	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Decode(id1)
		}
	})
}
