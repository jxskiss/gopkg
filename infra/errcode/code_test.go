package errcode

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCode(t *testing.T) {
	reg := New()
	dummy1 := reg.Register(100001, "")
	dummy2 := reg.RegisterReserved(100002, "dummy2")

	assert.Equal(t, dummy1.Error(), "[100001] unknown")
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

func TestCodeFormat(t *testing.T) {
	reg := New()
	badReqErr := reg.Register(100400, "bad request")

	t.Run("no details", func(t *testing.T) {
		err := func() error {
			return badReqErr
		}()

		got1 := fmt.Sprintf("%v", err)
		assert.Equal(t, "[100400] bad request", got1)

		got2 := fmt.Sprintf("%+v", err)
		assert.Equal(t, "[100400] bad request", got2)
	})

	t.Run("with details", func(t *testing.T) {
		err := func() error {
			return badReqErr.AddDetails("test detail", 12345)
		}()

		got1 := fmt.Sprintf("%v", err)
		assert.Equal(t, "[100400] bad request", got1)

		got2 := fmt.Sprintf("%+v", err)
		t.Log(got2)
		assert.Equal(t, "[100400] bad request\ndetails:\n -  test detail\n -  12345", got2)
	})
}

func TestDetails(t *testing.T) {
	reg := New()
	testErr := reg.Register(1001, "test error")

	t.Run("not ErrCode", func(t *testing.T) {
		err := errors.New("not an ErrCode")
		got := Details(err)
		assert.Nil(t, got)
	})

	t.Run("no details", func(t *testing.T) {
		err := testErr
		got := Details(err)
		assert.Nil(t, got)
	})

	t.Run("with details", func(t *testing.T) {
		err := testErr.AddDetails("test detail", 12345)
		got := Details(err)
		require.Len(t, got, 2)
		assert.Equal(t, "test detail", got[0])
		assert.Equal(t, 12345, got[1])
	})

	t.Run("wrapped", func(t *testing.T) {
		err := func() error {
			err1 := testErr.AddDetails("test detail", 12345)
			return fmt.Errorf("wrap message: %w", err1)
		}()
		got := Details(err)
		require.Len(t, got, 2)
		assert.Equal(t, "test detail", got[0])
		assert.Equal(t, 12345, got[1])
	})
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
	extra any
}

func (e *testErrWrapper1) Cause() error {
	return e.error
}

type testErrWrapper2 struct {
	error
	extra any
}

func (e *testErrWrapper2) Unwrap() error {
	return e.error
}
