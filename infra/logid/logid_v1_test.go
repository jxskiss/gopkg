package logid

import (
	"encoding/hex"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testV1Gen = NewV1Gen()

func TestV1Gen(t *testing.T) {
	now := time.Now().UnixMilli()

	gotLogId := testV1Gen.Gen()
	assert.Len(t, gotLogId, v1Length)

	info := Decode(gotLogId)
	assert.True(t, info.Valid())
	assert.Equal(t, "1", info.Version())
	assert.True(t, math.Abs(float64(info.Time().UnixMilli()-now)) <= 1)
	assert.Equal(t, "", hex.EncodeToString(info.IP().To16()))
	assert.Contains(t, info.String(), string(testV1Gen.(*v1Gen).machineID[:]))

	gotLogId2 := testV1Gen.Gen()
	info2 := Decode(gotLogId2)
	assert.True(t, info2.Valid())
	assert.NotEqual(t, info.Random(), info2.Random())
	assert.Equal(t, info.infoInterface.(*v1Info).machineID, info2.infoInterface.(*v1Info).machineID)

	info3 := Decode("1daolkwqalekralk")
	assert.False(t, info3.Valid())
	assert.Equal(t, "0|invalid", info3.String())
}
