package ezhttp

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"a": 1234,
		"b": "abcd",
	}
	JSON(w, 200, data)

	result := w.Result()
	defer result.Body.Close()
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, contentTypeJSON, result.Header.Get(hdrContentTypeKey))

	body, _ := io.ReadAll(result.Body)
	assert.Contains(t, string(body), `"a":1234`)
	assert.Contains(t, string(body), `"b":"abcd"`)
}

func TestJSONHumanFriendly(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[any]any{
		1234:   "a",
		"abcd": "b",
	}
	JSONHumanFriendly(w, 500, data)

	result := w.Result()
	defer result.Body.Close()
	assert.Equal(t, 500, result.StatusCode)
	assert.Equal(t, contentTypeJSON, result.Header.Get(hdrContentTypeKey))

	body, _ := io.ReadAll(result.Body)
	assert.Contains(t, string(body), "{\n    \"")
	assert.Contains(t, string(body), "\"\n}")
	assert.Contains(t, string(body), `"1234": "a"`)
	assert.Contains(t, string(body), `"abcd": "b"`)
}
