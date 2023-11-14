package validat

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/perf/lru"
)

func TestMatchRegexp(t *testing.T) {
	ctx := context.Background()
	vdErr := &ValidationError{}

	t.Run("string pattern", func(t *testing.T) {
		pattern := `\w+\s+\d+`

		got1, err := MatchRegexp("testVar", pattern, "abc 123").Validate(ctx, nil)
		require.Nil(t, err)
		assert.Equal(t, "abc 123", got1)

		_, err = MatchRegexp("testVar", pattern, "abc123").Validate(ctx, nil)
		require.NotNil(t, err)
		assert.True(t, errors.As(err, &vdErr))
		assert.Equal(t, `testVar: value "abc123" not match regexp`, err.Error())
	})

	t.Run("invalid pattern", func(t *testing.T) {
		pattern := `(\w+\s+\d+`
		_, err := MatchRegexp("testVar", pattern, "abc 123").Validate(ctx, nil)
		require.NotNil(t, err)
		assert.False(t, errors.As(err, &vdErr))
	})

	t.Run("regexp pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`\w+\s+\d+`)

		got1, err := MatchRegexp("testVar", pattern, "abc 123").Validate(ctx, nil)
		require.Nil(t, err)
		assert.Equal(t, "abc 123", got1)

		_, err = MatchRegexp("testVar", pattern, "abc123").Validate(ctx, nil)
		require.NotNil(t, err)
		assert.True(t, errors.As(err, &vdErr))
		assert.Equal(t, `testVar: value "abc123" not match regexp`, err.Error())
	})

	t.Run("cache enabled", func(t *testing.T) {
		cache := lru.NewCache[string, *regexp.Regexp](100)
		EnableRegexpCache(cache)

		pattern := `\w+\s+\d+`

		got1, err := MatchRegexp("testVar", pattern, "abc 123").Validate(ctx, nil)
		require.Nil(t, err)
		assert.Equal(t, "abc 123", got1)

		_, err = MatchRegexp("testVar", pattern, "abc123").Validate(ctx, nil)
		require.NotNil(t, err)
		assert.True(t, errors.As(err, &vdErr))
	})
}
