//go:build go1.23

package functicker

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	typBool         = reflect.TypeOf(false)
	typTimeRecvChan = reflect.TypeOf((<-chan time.Time)(nil))
)

func TestTickerDefinition(t *testing.T) {
	timer := reflect.ValueOf(time.Timer{})
	assert.Equal(t, 2, timer.Type().NumField())
	assert.Equal(t, "C", timer.Type().Field(0).Name)
	assert.Equal(t, typTimeRecvChan, timer.Field(0).Type())
	assert.Equal(t, "initTimer", timer.Type().Field(1).Name)
	assert.Equal(t, typBool, timer.Field(1).Type())

	ticker := reflect.ValueOf(time.Ticker{})
	assert.Equal(t, 2, ticker.Type().NumField())
	assert.Equal(t, "C", ticker.Type().Field(0).Name)
	assert.Equal(t, typTimeRecvChan, ticker.Field(0).Type())
	assert.Equal(t, "initTicker", ticker.Type().Field(1).Name)
	assert.Equal(t, typBool, ticker.Field(1).Type())
}
