package easy

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseTemplates(t *testing.T) {
	var files []string
	add := func(name string, text []byte) error {
		log.Println("template:", name)
		files = append(files, name)
		return nil
	}
	rootDir := "./"
	rePattern := `.+\.go`
	err := parseTemplates(rootDir, rePattern, add)
	assert.Nil(t, err)
	assert.True(t, len(files) > 0)
}
