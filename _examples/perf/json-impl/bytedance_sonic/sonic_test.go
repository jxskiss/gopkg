package bytedance_sonic

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

func TestSonicImpl(t *testing.T) {
	impl := New(sonic.ConfigDefault, true)
	json.ChangeImpl(impl)

	x := struct {
		A int
		B string
	}{123, "abc"}
	want := `{"A":123,"B":"abc"}`

	t.Run("Marshal", func(t *testing.T) {
		got, err := json.Marshal(x)
		require.Nil(t, err)
		assert.Equal(t, want, string(got))
	})

	t.Run("MarshalToString", func(t *testing.T) {
		got, err := json.MarshalToString(x)
		require.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("MarshalFastest", func(t *testing.T) {
		got, err := json.MarshalFastest(x)
		require.Nil(t, err)
		assert.Equal(t, want, string(got))
	})
}
