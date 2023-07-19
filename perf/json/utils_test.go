package json

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAndDump(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "")
	require.Nil(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	data := []int{1, 2, 3}

	err = Dump(tmpFile.Name(), data, "", "  ")
	require.Nil(t, err)

	var got []int64
	err = Load(tmpFile.Name(), &got)
	require.Nil(t, err)
	assert.Equal(t, []int64{1, 2, 3}, got)
}
