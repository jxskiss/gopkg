package logid

import (
	"encoding/hex"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestV1Gen(t *testing.T) {
	gen := &v1Gen{}
	now := time.Now().UnixMilli()

	gotLogId := gen.Gen()
	assert.Len(t, gotLogId, v1Length)

	info := Decode(gotLogId)
	assert.True(t, info.Valid())
	assert.Equal(t, "1", info.Version())
	assert.True(t, math.Abs(float64(info.Time().UnixMilli()-now)) <= 1)
	assert.Equal(t, localIPStr, hex.EncodeToString(info.IP().To16()))

	gotLogId2 := gen.Gen()
	info2 := Decode(gotLogId2)
	assert.True(t, info2.Valid())
	assert.NotEqual(t, info.Random(), info2.Random())
}
