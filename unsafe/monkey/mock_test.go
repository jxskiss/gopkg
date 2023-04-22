package monkey_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/monkey"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/testpkg"
)

func _T(y, d int) time.Time {
	return time.Date(y, time.January, d, 0, 0, 0, 0, time.UTC)
}

func TestMock(t *testing.T) {
	t.Run("function", func(t *testing.T) {
		before := time.Now()
		monkey.AutoUnpatch(func() {
			monkey.Mock(time.Now).To(func() time.Time { return _T(2000, 1) }).Build()
			during := time.Now()
			assert.Equal(t, _T(2000, 1), during)

			monkey.Mock(time.Now).Return(_T(2001, 1)).Build()
			during = time.Now()
			assert.Equal(t, _T(2001, 1), during)

			monkey.AutoUnpatch(func() {
				monkey.Mock(time.Now).To(func() time.Time { return _T(2002, 1) }).Build()
				during := time.Now()
				assert.Equal(t, _T(2002, 1), during)

				monkey.Mock(time.Now).Return(_T(2003, 1)).Build()
				during = time.Now()
				assert.Equal(t, _T(2003, 1), during)
			})
			assert.Equal(t, time.Now(), _T(2001, 1))
		})
		assert.True(t, time.Now().After(before))
	})

	t.Run("method", func(t *testing.T) {
		i := &f{}
		monkey.AutoUnpatch(func() {
			assert.Equal(t, 0, i.Value())

			monkey.Mock((*f).Value).Return(1).Build()
			assert.Equal(t, 1, i.Value())

			monkey.Mock().Method(i, "Value").To(func(*f) int { return 2 }).Build()
			assert.Equal(t, 2, i.Value())

			monkey.AutoUnpatch(func() {
				monkey.Mock((*f).Value).Return(3).Build()
				assert.Equal(t, 3, i.Value())

				monkey.Mock().Method(i, "Value").To(func(*f) int { return 4 }).Build()
				assert.Equal(t, 4, i.Value())
			})
			assert.Equal(t, 2, i.Value())
		})

		obj := testpkg.NewTestObj()
		assert.Equal(t, 0, obj.Value())
		monkey.AutoUnpatch(func() {
			monkey.Mock().Method(obj, "Value").Return(1).Build()
			assert.Equal(t, 1, obj.Value())
		})
		assert.Equal(t, 0, obj.Value())
	})

	t.Run("byName", func(t *testing.T) {
		testpkg_a := "github.com/jxskiss/gopkg/v2/unsafe/monkey/testpkg.a"
		assert.Equal(t, "testpkg.a", testpkg.A())

		monkey.AutoUnpatch(func() {
			var sig func() string
			monkey.Mock().ByName(testpkg_a, sig).Return("TestMock / byName").Build()
			assert.Equal(t, "TestMock / byName", testpkg.A())

			mockFn := func() string {
				return "mock fn 1"
			}
			p1 := monkey.Mock().ByName(testpkg_a, mockFn).To(mockFn).Build()
			assert.Equal(t, "mock fn 1", testpkg.A())

			p1.Delete()
			assert.Equal(t, "TestMock / byName", testpkg.A())
		})
		assert.Equal(t, "testpkg.a", testpkg.A())
	})
}
