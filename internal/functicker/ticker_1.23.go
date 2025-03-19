//go:build go1.23

package functicker

import (
	"reflect"
	"time"
)

// From Go 1.23, time.Ticker and time.Timer are defined as different structs,
// though the underlying data is still same, we have to use unsafe trick
// to cast the timer pointer.
//
// !!! WE MUST ASSERT THE DEFINITION OF TICKER AND TIMER MATCHES !!!

var (
	typBool         = reflect.TypeOf(false)
	typTimeRecvChan = reflect.TypeOf((<-chan time.Time)(nil))
)

func init() {
	errMessage := "functicker: time.Ticker and time.Timer definition not match"
	timer := reflect.ValueOf(time.Timer{})
	ticker := reflect.ValueOf(time.Ticker{})
	if timer.Type().NumField() != 2 ||
		timer.Type().Field(0).Name != "C" || timer.Field(0).Type() != typTimeRecvChan ||
		timer.Type().Field(1).Name != "initTimer" || timer.Field(1).Type() != typBool {
		panic(errMessage)
	}
	if ticker.Type().NumField() != 2 ||
		ticker.Type().Field(0).Name != "C" || ticker.Field(0).Type() != typTimeRecvChan ||
		ticker.Type().Field(1).Name != "initTicker" || ticker.Field(1).Type() != typBool {
		panic(errMessage)
	}
}
