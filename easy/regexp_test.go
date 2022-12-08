package easy

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchGroups(t *testing.T) {
	text := "Tom saw a cat got a mat."
	re := regexp.MustCompile(`(\w+)\s(?P<action>\w+)\s.*`)

	want1 := map[string][]byte{
		"action": []byte("saw"),
	}
	got1 := MatchGroups(re, []byte(text))
	assert.Equal(t, want1, got1)

	want2 := map[string]string{
		"action": "saw",
	}
	got2 := MatchStringGroups(re, text)
	assert.Equal(t, want2, got2)
}
