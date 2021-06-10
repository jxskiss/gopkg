package errcode

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCode(t *testing.T) {
	reg := New()
	dummy1 := reg.Register(100001, "")
	dummy2 := reg.RegisterReserved(100002, "dummy2")

	assert.Equal(t, dummy1.Error(), "[100001] (no message)")
	assert.Equal(t, dummy2.Error(), "[100002] dummy2")

	got1 := func() error { return dummy1 }()
	got2 := func() error { return dummy2 }()
	assert.True(t, Is(got1, dummy1))
	assert.True(t, Is(got2, dummy2))
	assert.True(t, IsErrCode(got1))
	assert.True(t, IsErrCode(got2))

	assert.False(t, Is(nil, dummy1))
	assert.False(t, Is(errors.New("dummy1"), dummy1))
	assert.False(t, IsErrCode(nil))
	assert.False(t, IsErrCode(errors.New("dummy1")))

	json1, _ := dummy1.MarshalJSON()
	json2, _ := dummy2.MarshalJSON()
	assert.Equal(t, []byte(`{"code":100001}`), json1)
	assert.Equal(t, []byte(`{"code":100002,"message":"dummy2"}`), json2)
}
