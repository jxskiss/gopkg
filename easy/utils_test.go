package easy

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testObject struct {
	A int
	B string
}

func TestSetDefault(t *testing.T) {
	intValues := []interface{}{int(1), int32(1), uint16(1), uint64(1)}
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

var prettyTestWant = strings.TrimSpace(`
{
    "1": 123,
    "b": "<html>"
}`)

func TestPretty(t *testing.T) {
	test := map[string]interface{}{
		"1": 123,
		"b": "<html>",
	}
	jsonString := JSON(test)
	assert.Equal(t, `{"1":123,"b":"<html>"}`, jsonString)

	got1 := Pretty(test)
	assert.Equal(t, prettyTestWant, got1)

	got2 := Pretty(jsonString)
	assert.Equal(t, prettyTestWant, got2)

	test3 := []byte("<fff> not a json object")
	got3 := Pretty(test3)
	assert.Equal(t, string(test3), got3)

	test4 := make([]byte, 16)
	rand.Read(test4)
	got4 := Pretty(test4)
	assert.Equal(t, "<pretty: non-printable bytes>", got4)
}

func TestCaller(t *testing.T) {
	name, file, line := Caller(0)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/utils_test.go", file)
	assert.Equal(t, 75, line)
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

func TestLazyJSON(t *testing.T) {
	var x = &testObject{A: 123, B: "abc"}
	got1 := JSON(x)
	got2 := fmt.Sprintf("%v", LazyJSON(x))
	assert.Equal(t, got1, got2)
}
