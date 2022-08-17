package logid

import (
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testIP = "10.1.2.3"
var testV2Gen = NewV2Gen(net.ParseIP(testIP))

func TestV2Gen(t *testing.T) {
	now := time.Now().UnixMilli()

	gotLogId := testV2Gen.Gen()
	assert.Len(t, gotLogId, v2Length)

	info := Decode(gotLogId)
	assert.True(t, info.Valid())
	assert.Equal(t, "2", info.Version())
	assert.True(t, math.Abs(float64(info.Time().UnixMilli()-now)) <= 1)
	assert.Equal(t, testIP, info.IP().String())

	gotLogId2 := testV2Gen.Gen()
	info2 := Decode(gotLogId2)
	assert.True(t, info2.Valid())
	assert.NotEqual(t, info.Random(), info2.Random())

	info3 := Decode("2daolkwqalekralk")
	assert.False(t, info3.Valid())
	assert.Equal(t, "0|invalid", info3.String())
}
