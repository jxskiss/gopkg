package vdutil

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGreaterThanZero(t *testing.T) {
	var vdErr *ValidationError

	_, err := Validate(context.Background(),
		GreaterThanZero("testVar", 100, false))
	require.Nil(t, err)

	_, err = Validate(context.Background(),
		GreaterThanZero("testVar", -100, false))
	require.NotNil(t, err)
	assert.True(t, errors.As(err, &vdErr))
	assert.Contains(t, err.Error(), "testVar: value -100 <= 0")

	got1, err := Validate(context.Background(),
		GreaterThanZero("testVar", int64(100), true))
	require.Nil(t, err)
	assert.Equal(t, int64(100), got1.Data.GetInt("testVar"))

	got2, err := Validate(context.Background(),
		GreaterThanZero("testVar", "100", true))
	require.Nil(t, err)
	assert.Equal(t, int64(100), got2.Data.GetInt("testVar"))

	got3, err := Validate(context.Background(),
		GreaterThanZero("testVar", "0", true))
	require.NotNil(t, err)
	require.NotNil(t, got3)
	assert.True(t, errors.As(err, &vdErr))
	assert.Contains(t, err.Error(), "testVar: value 0 <= 0")

	got4, err := Validate(context.Background(),
		GreaterThanZero("testVar", "xyz", true))
	require.NotNil(t, err)
	require.NotNil(t, got4)
	assert.True(t, errors.As(err, &vdErr))
	assert.Contains(t, err.Error(), "testVar: value xyz is not integer")
}

func TestAllElementsGreaterThanZero(t *testing.T) {
	t.Run("not integer", func(t *testing.T) {
		slice := []string{"1", "xyz", "2"}
		got, err := Validate(context.Background(),
			AllElementsGreaterThanZero("testVar", slice, true))
		require.NotNil(t, err)
		require.NotNil(t, got)
		assert.Contains(t, err.Error(), "slice has non-integer element")
	})

	t.Run("valid", func(t *testing.T) {
		slice := []int{1, 2, 3}
		got, err := Validate(context.Background(),
			AllElementsGreaterThanZero("testVar", slice, true))
		require.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, got.Data.MustGet("testVar"))
	})

	t.Run("invalid", func(t *testing.T) {
		slice := []uint8{1, 2, 3, 0}
		got, err := Validate(context.Background(),
			AllElementsGreaterThanZero("testVar", slice, true))
		require.NotNil(t, err)
		require.NotNil(t, got)
		assert.Contains(t, err.Error(), "slice element 0 <= 0")
	})
}

func TestAllElementsNotZero(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		slice := []int32{1, 2}
		_, err := Validate(context.Background(),
			AllElementsNotZero("testVar", slice))
		require.Nil(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		slice := []string{"abc", ""}
		_, err := Validate(context.Background(),
			AllElementsNotZero("testVar", slice))
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "slice has zero element")
	})
}

func TestLessThanOrEqual(t *testing.T) {
	var vdErr *ValidationError

	_, err := Validate(context.Background(),
		LessThanOrEqual("testVar", 20, 20))
	require.Nil(t, err)

	_, err = Validate(context.Background(),
		LessThanOrEqual("testVar", 20, 25))
	require.NotNil(t, err)
	assert.True(t, errors.As(err, &vdErr))
	assert.Contains(t, err.Error(), "testVar: value 25 > 20")
}

func TestInRange(t *testing.T) {
	got1, err := Validate(context.Background(),
		InRange("count", 1, 20, 15))
	require.Nil(t, err)
	_ = got1

	got2, err := Validate(context.Background(),
		InRange("count", 1, 20, 100))
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "count: value 100 is not in range [1, 20]")
	_ = got2
}

func TestInRangeMode(t *testing.T) {
	testData := []struct {
		Name     string
		Mode     RangeMode
		Min      int
		Max      int
		Value    int
		ErrIsNil bool
		ErrMsg   string
	}{
		{"testVar", GtAndLte, 1, 20, 15, true, ""},
		{"testVar", GtAndLte, 1, 20, 1, false, "testVar: value 1 is not in range (1, 20]"},
		{"testVar", GtAndLte, 1, 20, 20, true, ""},

		{"testVar", GtAndLt, 1, 20, 15, true, ""},
		{"testVar", GtAndLt, 1, 20, 1, false, "testVar: value 1 is not in range (1, 20)"},
		{"testVar", GtAndLt, 1, 20, 20, false, "testVar: value 20 is not in range (1, 20)"},

		{"testVar", GteAndLte, 1, 20, 15, true, ""},
		{"testVar", GteAndLte, 1, 20, 1, true, ""},
		{"testVar", GteAndLte, 1, 20, 20, true, ""},

		{"testVar", GteAndLt, 1, 20, 15, true, ""},
		{"testVar", GteAndLt, 1, 20, 1, true, ""},
		{"testVar", GteAndLt, 1, 20, 20, false, "testVar: value 20 is not in range [1, 20)"},
	}

	var vdErr *ValidationError
	for _, c := range testData {
		_, err := Validate(context.Background(),
			InRangeMode(c.Name, c.Mode, c.Min, c.Max, c.Value))
		if c.ErrIsNil {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.True(t, errors.As(err, &vdErr))
			assert.Contains(t, err.Error(), c.ErrMsg)
		}
	}
}

func TestParseStringSliceToInt64Slice(t *testing.T) {
	got1, err := Validate(context.Background(),
		ParseStrsToInt64Slice("entityIDs", []string{"1", "2", "3"}))
	assert.Nil(t, err)
	assert.Equal(t, []int64{1, 2, 3}, got1.Data.GetInt64s("entityIDs"))
}

func TestParseStringSliceToInt64Map(t *testing.T) {
	got1, err := Validate(context.Background(),
		ParseStrsToInt64Map("entityIDs", []string{"1", "2", "3"}))
	assert.Nil(t, err)
	assert.Equal(t, map[int64]bool{1: true, 2: true, 3: true}, got1.Data.MustGet("entityIDs").(map[int64]bool))
}

func TestNotNil(t *testing.T) {
	notNilValues := []any{
		1,
		map[int]int{},
		[]int{},
		&Result{},
		GreaterThanZero("", 1234, false),
		Rule(GreaterThanZero("", 1234, false)),
	}
	for _, x := range notNilValues {
		_, err := Validate(context.Background(), NotNil("testVar", x))
		assert.Nil(t, err)
	}

	var vdErr *ValidationError
	nilValues := []any{
		nil,
		(*int)(nil),
		(map[int]int)(nil),
		([]int)(nil),
		(*Result)(nil),
		RuleFunc(nil),
		Rule(nil),
	}
	for _, x := range nilValues {
		_, err := Validate(context.Background(), NotNil("testVar", x))
		assert.NotNil(t, err)
		assert.True(t, errors.As(err, &vdErr))
		assert.Contains(t, err.Error(), "value is nil")
	}
}
