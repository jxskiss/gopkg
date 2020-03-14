package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []map[string]interface{}{
		{
			"value": 123,
			"want":  "123",
		},
		{
			"value": "456",
			"want":  `"456"`,
		},
		{
			"value": simple{"ABC"},
			"want":  `{"A":"ABC"}`,
		},
		{
			"value": "<html></html>",
			"want":  `"<html></html>"`,
		},
	}
	for _, test := range tests {
		x := JSON(test["value"])
		assert.Equal(t, test["want"], x)
	}
}

func TestCaller(t *testing.T) {
	name, file, line := Caller(1)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/log_test.go", file)
	assert.Equal(t, 34, line)
}
