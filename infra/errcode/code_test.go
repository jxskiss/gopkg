package errcode

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestIs(t *testing.T) {
	reg := New()
	code1 := reg.Register(100001, "test code1")

	detailsErr := code1.AddDetails("dummy detail")
	assert.True(t, Is(detailsErr, code1))

	wrapErr1 := &testErrWrapper1{error: detailsErr}
	assert.True(t, Is(detailsErr, code1))

	wrapErr2 := &testErrWrapper2{error: detailsErr}
	assert.True(t, Is(wrapErr2, code1))

	wrapErr3 := &testErrWrapper2{error: wrapErr1}
	assert.True(t, Is(wrapErr3, code1))

	wrapErr4 := &testErrWrapper1{error: wrapErr2}
	assert.True(t, Is(wrapErr4, code1))
}

func TestErrorsCompatibility(t *testing.T) {
	reg := New()
	code1 := reg.Register(100001, "test code1")

	detailsErr := code1.AddDetails("dummy detail")
	assert.True(t, errors.Is(detailsErr, code1))

	wrapErr1 := &testErrWrapper2{error: detailsErr}
	assert.True(t, errors.Is(wrapErr1, code1))

	wrapErr2 := &testErrWrapper2{error: wrapErr1}
	assert.True(t, errors.Is(wrapErr2, code1))
}

type testErrWrapper1 struct {
	error
	extra interface{}
}

func (e *testErrWrapper1) Cause() error {
	return e.error
}

type testErrWrapper2 struct {
	error
	extra interface{}
}

func (e *testErrWrapper2) Unwrap() error {
	return e.error
}
