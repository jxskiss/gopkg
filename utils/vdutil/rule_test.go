package vdutil

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	var vdErr *ValidationError
	err := fmt.Errorf("with message: %w", &ValidationError{Name: "testVar", Err: errors.New("test error")})
	ok := errors.As(err, &vdErr)
	assert.True(t, ok)
	assert.NotNil(t, vdErr)
	assert.Equal(t, err.Error(), "with message: testVar: test error")
	assert.Equal(t, vdErr.Error(), "testVar: test error")
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	got1, err := Validate(ctx,
		GreaterThanZero("var1", "10", true),
		LessThanOrEqual("var2", 100, 200),
	)
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "var2: value 200 > 100")
	assert.EqualValues(t, 10, got1.Data.GetInt("var1"))
	assert.True(t, got1.IsValidationError)

	got2, err := Validate(ctx,
		GreaterThanZero("var1", 10, false),
		LessThanOrEqual("var2", 100, 100),
		ParseStrsToInt64Slice("var3", []string{"123", "456", "789"}),
	)
	require.Nil(t, err)
	assert.Equal(t, []int64{123, 456, 789}, got2.Data.GetSlice("var3"))
	assert.False(t, got2.IsValidationError)

	got3, err := Validate(ctx,
		RuleFunc(func(ctx context.Context, result *Result) (any, error) {
			return nil, errors.New("internal error")
		}))
	require.NotNil(t, err)
	require.False(t, got3.IsValidationError)
}
