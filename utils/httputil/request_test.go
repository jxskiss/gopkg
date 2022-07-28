package httputil

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(testHttpHandler))
	defer s.Close()
	data := "test Do Request"
	req, _ := http.NewRequest("POST", s.URL, strings.NewReader(data))

	var respText []byte
	var status int
	var err error

	_, respText, status, err = Do(&Request{
		Req:          req,
		Timeout:      time.Second,
		DumpRequest:  true,
		DumpResponse: true,
	})

	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(respText), data)
}

func testHttpHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	w.Write(dump)
}
