package ezhttp

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHttpHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	w.Write(dump)
}

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

	require.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(respText), data)
}

func TestDiscardResponseBody(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(testHttpHandler))
	defer s.Close()
	data := "test DiscardResponseBody"

	_, respText, status, err := Do(&Request{
		Method:              http.MethodGet,
		URL:                 s.URL,
		Body:                data,
		DiscardResponseBody: true,
	})

	require.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "", string(respText))
}

func TestMergeRequest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(testHttpHandler))
	defer s.Close()

	t.Run("build request", func(t *testing.T) {
		reqURL := s.URL + "?k1=v1"
		req := &Request{
			URL: reqURL,
			Params: map[string]any{
				"k2": "v2",
				"k3": 345,
			},
			Headers: map[string]string{
				"Content-Type":  "application/test",
				"x-my-header-1": "test_header",
			},
		}

		_, respText, status, err := Do(req)
		require.Nil(t, err)
		assert.Equal(t, 200, status)
		assert.Contains(t, string(respText), "?k1=v1&k2=v2&k3=345")
		assert.Contains(t, string(respText), "Content-Type: application/test")
		assert.Contains(t, string(respText), "X-My-Header-1: test_header")
	})

	t.Run("prepared request", func(t *testing.T) {
		reqURL := s.URL + "?k1=v1"
		httpReq, _ := http.NewRequest("GET", reqURL, nil)
		req := &Request{
			Req: httpReq,
			Params: map[string]string{
				"k2": "v2",
				"k3": "345",
			},
			Headers: map[string]string{
				"Content-Type":  "application/test",
				"x-my-header-1": "test_header",
			},
		}

		_, respText, status, err := Do(req)
		require.Nil(t, err)
		assert.Equal(t, 200, status)
		assert.Contains(t, string(respText), "?k1=v1&k2=v2&k3=345")
		assert.Contains(t, string(respText), "Content-Type: application/test")
		assert.Contains(t, string(respText), "X-My-Header-1: test_header")
	})

	t.Run("with trailing ampersand", func(t *testing.T) {
		reqURL := s.URL + "?k1=v1&"
		httpReq, _ := http.NewRequest("GET", reqURL, nil)
		req := &Request{
			Req: httpReq,
			Params: map[string][]string{
				"k2": {"v2"},
				"k3": {"345"},
			},
		}

		_, respText, status, err := Do(req)
		require.Nil(t, err)
		assert.Equal(t, 200, status)
		assert.Contains(t, string(respText), "?k1=v1&k2=v2&k3=345")
	})
}

func TestRequestSetBasicAuth(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(testHttpHandler))
	defer s.Close()

	t.Run("build request", func(t *testing.T) {
		req := (&Request{
			URL: s.URL,
		}).SetBasicAuth("user", "pass")

		_, respText, status, err := Do(req)
		require.Nil(t, err)
		assert.Equal(t, 200, status)
		assert.Contains(t, string(respText), "Authorization: ")
	})

	t.Run("prepared request", func(t *testing.T) {
		httpReq, _ := http.NewRequest("GET", s.URL, nil)
		req := &Request{
			Req: httpReq,
		}
		req.SetBasicAuth("user", "pass")

		_, respText, status, err := Do(req)
		require.Nil(t, err)
		assert.Equal(t, 200, status)
		assert.Contains(t, string(respText), "Authorization: ")
	})
}
