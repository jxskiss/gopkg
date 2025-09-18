package easy

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testObject struct {
	A int
	B string
}

func TestSetDefault(t *testing.T) {
	intValues := []any{int(1), int32(1), uint16(1), uint64(1)}
	for _, value := range intValues {
		var tmp int16
		SetDefault(&tmp, value)
		assert.Equal(t, int16(1), tmp)
	}

	var ptr *testObject
	var tmp = &testObject{A: 1, B: "b"}
	SetDefault(&ptr, tmp)
	assert.Equal(t, testObject{A: 1, B: "b"}, *ptr)
	assert.Equal(t, tmp, ptr)
}

func TestSetDefault_ShouldPanic(t *testing.T) {
	var ptr *testObject
	var tmp = &testObject{A: 1, B: "b"}

	err := Safe(func() {
		SetDefault(ptr, tmp)
	})()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SetDefault")
	assert.Contains(t, err.Error(), "must be a non-nil pointer")
}

func TestCaller(t *testing.T) {
	name, file, line := Caller(0)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/utils_test.go", file)
	assert.Equal(t, 44, line)
}

func TestCallerName(t *testing.T) {
	name := CallerName()
	assert.Equal(t, "easy.TestCallerName", name)
}

func TestSingleJoin(t *testing.T) {
	text := []string{"a", "b..", "..c"}
	got := SingleJoin("..", text...)
	want := "a..b..c"
	assert.Equal(t, want, got)
}

func TestSlashJoin(t *testing.T) {
	got0 := SlashJoin()
	assert.Equal(t, "", got0)

	path1 := []string{"/a", "b", "c.png"}
	want1 := "/a/b/c.png"
	got1 := SlashJoin(path1...)
	assert.Equal(t, want1, got1)

	path2 := []string{"/a/", "b/", "/c.png"}
	want2 := "/a/b/c.png"
	got2 := SlashJoin(path2...)
	assert.Equal(t, want2, got2)
}

func TestWithTimeout(t *testing.T) {
	t.Run("no timeout", func(t *testing.T) {
		err := WithTimeout("test", time.Second, func() error {
			return nil
		})
		assert.Nil(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		err := WithTimeout("test", 100*time.Millisecond, func() error {
			time.Sleep(time.Second)
			return nil
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "operation test timeout")
		t.Log(err.Error())
	})

	t.Run("error", func(t *testing.T) {
		err := WithTimeout("test", 100*time.Millisecond, func() error {
			return errors.New("test error")
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "test error")
		t.Log(err.Error())
	})

	t.Run("panic", func(t *testing.T) {
		err := WithTimeout("test", 100*time.Millisecond, func() error {
			panic("test panic")
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "test panic")
		t.Log(err.Error())
	})
}
