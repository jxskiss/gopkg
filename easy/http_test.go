package easy

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

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

func TestJsonToReader(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": "2",
	}
	r, err := JsonToReader(data)
	assert.Nil(t, err)
	buf, _ := ioutil.ReadAll(r)
	want := []byte(`{"a":1,"b":"2"}`)
	assert.Equal(t, want, buf)
}

func TestDecodeJson(t *testing.T) {
	data := make(map[string]interface{})
	r := bytes.NewBufferString(`{"a":1,"b":"2"}`)
	err := DecodeJson(r, &data)
	assert.Nil(t, err)

	want := map[string]interface{}{"a": float64(1), "b": "2"}
	assert.Equal(t, want, data)
}

func BenchmarkSlashJoin(b *testing.B) {
	path := []string{"/a", "b", "c.png"}
	for i := 0; i < b.N; i++ {
		_ = SlashJoin(path...)
	}
}
