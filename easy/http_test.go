package easy

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"
)

func TestSingleJoin(t *testing.T) {
	text := []string{"a","b..", "..c"}
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

type testObject struct {
	A int    `xml:"a" json:"a"`
	B string `xml:"b" json:"b"`
}

func TestJSONToReader(t *testing.T) {
	data := testObject{
		A: 1,
		B: "2",
	}
	r, err := JSONToReader(data)
	require.Nil(t, err)

	buf, _ := ioutil.ReadAll(r)
	want := []byte(`{"a":1,"b":"2"}`)
	assert.Equal(t, want, buf)
}

func TestDecodeJSON(t *testing.T) {
	var data testObject
	r := bytes.NewBufferString(`{"a":1,"b":"2"}`)
	err := DecodeJSON(r, &data)
	require.Nil(t, err)

	want := testObject{A: 1, B: "2"}
	assert.Equal(t, want, data)
}

func TestXMLToReader(t *testing.T) {
	var data = testObject{
		A: 123,
		B: "456",
	}
	r, err := XMLToReader(data)
	require.Nil(t, err)

	buf, _ := ioutil.ReadAll(r)
	want := []byte(`<testObject><a>123</a><b>456</b></testObject>`)
	assert.Equal(t, want, buf)
}

func TestDecodeXML(t *testing.T) {
	var data testObject
	r := bytes.NewBufferString(`<testObject><a>123</a><b>456</b></testObject>`)
	err := DecodeXML(r, &data)
	require.Nil(t, err)

	want := testObject{A: 123, B: "456"}
	assert.Equal(t, want, data)
}

func BenchmarkSlashJoin(b *testing.B) {
	path := []string{"/a", "b", "c.png"}
	for i := 0; i < b.N; i++ {
		_ = SlashJoin(path...)
	}
}

func TestDoRequest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(testHttpHandler))
	defer s.Close()
	data := "test DoRequest"
	req, _ := http.NewRequest("POST", s.URL, strings.NewReader(data))

	var respText []byte
	var status int
	var err error

	logbuf := CopyStdLog(func() {
		respText, status, err = DoRequest(&Request{
			Req:          req,
			Timeout:      time.Second,
			DumpRequest:  true,
			DumpResponse: true,
		})
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(respText), data)
	count := bytes.Count(logbuf, []byte(data))
	assert.Equal(t, 2, count)
}

func testHttpHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	w.Write(dump)
}
