package monkey

import (
	"fmt"
	"time"

	"github.com/jxskiss/gopkg/v2/unsafe/monkey/testpkg"
)

func MyTimeFunc(year, month, day, hour, min, sec, nsec int) time.Time {
	_ = time.Now()
	return time.Date(year, time.Month(month), day, hour, min, sec, nsec, time.UTC)
}

func mock333(year, month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(2003, time.March, 3, 3, 3, 3, 3, time.UTC)
}

func Example() {
	beforePatch := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
	fmt.Printf("original: %v\n", beforePatch)

	patch := PatchFunc(MyTimeFunc, func(year, month, day, hour, min, sec, nsec int) time.Time {
		return time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC)
	})
	duringPatch := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
	fmt.Printf("patched: %v\n", duringPatch)

	patch.Delete()
	unpatched := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
	fmt.Printf("unpatched: %v\n", unpatched)

	fmt.Printf("before AutoUnpatch\n")

	AutoUnpatch(func() {
		tmp := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, before mock: %v\n", tmp)

		// We can use `Return` to return specified values.
		Mock().Target(MyTimeFunc).
			Return(time.Date(2001, time.January, 1, 1, 1, 1, 1, time.UTC)).
			Build()
		tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, after mock: %v\n", tmp)

		var innerFunc = func(year, month, day, hour, min, sec, nsec int) time.Time {
			return time.Date(2002, time.February, 2, 2, 2, 2, 2, time.UTC)
		}

		AutoUnpatch(func() {
			tmp := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
			fmt.Printf("inner AutoUnpatch, before mock: %v\n", tmp)

			// We can also use `To` to specify a replacement function.
			Mock().Target(MyTimeFunc).To(innerFunc).Build()
			tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
			fmt.Printf("inner AutoUnpatch, after mock: %v\n", tmp)
		})

		tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, after inner: %v\n", tmp)

		// In AutoUnpatch, we can also use the Patch* functions.

		// Static function works.
		PatchFunc(MyTimeFunc, mock333)
		tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, after inner mock333: %v\n", tmp)

		// Closure also works.
		var fakeTime = time.Date(2004, time.April, 4, 4, 4, 4, 4, time.UTC)
		var mock444 = func(year, month, day, hour, min, sec, nsec int) time.Time {
			return fakeTime
		}
		PatchFunc(MyTimeFunc, mock444)
		tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, after inner mock444: %v\n", tmp)

		// We can also patch a function by name.
		Mock().ByName("github.com/jxskiss/gopkg/v2/unsafe/monkey.MyTimeFunc",
			(func(year, month, day, hour, min, sec, nsec int) time.Time)(nil)).
			Return(time.Date(2005, time.May, 5, 5, 5, 5, 5, time.UTC)).
			Build()
		tmp = MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
		fmt.Printf("with AutoUnpatch, ByName: %v\n", tmp)
	})

	afterAutoUnpatch := MyTimeFunc(2021, 2, 3, 4, 5, 6, 7)
	fmt.Printf("after AutoUnpatch: %v\n", afterAutoUnpatch)

	// Output:
	// original: 2021-02-03 04:05:06.000000007 +0000 UTC
	// patched: 2001-01-01 01:01:01.000000001 +0000 UTC
	// unpatched: 2021-02-03 04:05:06.000000007 +0000 UTC
	// before AutoUnpatch
	// with AutoUnpatch, before mock: 2021-02-03 04:05:06.000000007 +0000 UTC
	// with AutoUnpatch, after mock: 2001-01-01 01:01:01.000000001 +0000 UTC
	// inner AutoUnpatch, before mock: 2001-01-01 01:01:01.000000001 +0000 UTC
	// inner AutoUnpatch, after mock: 2002-02-02 02:02:02.000000002 +0000 UTC
	// with AutoUnpatch, after inner: 2001-01-01 01:01:01.000000001 +0000 UTC
	// with AutoUnpatch, after inner mock333: 2003-03-03 03:03:03.000000003 +0000 UTC
	// with AutoUnpatch, after inner mock444: 2004-04-04 04:04:04.000000004 +0000 UTC
	// with AutoUnpatch, ByName: 2005-05-05 05:05:05.000000005 +0000 UTC
	// after AutoUnpatch: 2021-02-03 04:05:06.000000007 +0000 UTC
}

func ExamplePatchVar() {
	var someVar = 1234
	fmt.Printf("original: %v\n", someVar)

	patch := PatchVar(&someVar, 2345)
	fmt.Printf("patched: %v\n", someVar)

	patch.Delete()
	fmt.Printf("unpatched: %v\n", someVar)

	fmt.Printf("before AutoUnpatch\n")

	AutoUnpatch(func() {
		fmt.Printf("with AutoUnpatch, before patch: %v\n", someVar)

		PatchVar(&someVar, 3456)
		fmt.Printf("with AutoUnpatch, after patch: %v\n", someVar)

		AutoUnpatch(func() {
			fmt.Printf("inner AutoUnpatch, before patch: %v\n", someVar)

			PatchVar(&someVar, 4567)
			fmt.Printf("inner AutoUnpatch, after patch: %v\n", someVar)
		})

		fmt.Printf("with AutoUnpatch, after inner: %v\n", someVar)

		// Patch again.
		PatchVar(&someVar, 5678)
		fmt.Printf("with AutoUnpatch, patch again: %v\n", someVar)
	})

	fmt.Printf("after AutoUnpatch: %v\n", someVar)

	// Output:
	// original: 1234
	// patched: 2345
	// unpatched: 1234
	// before AutoUnpatch
	// with AutoUnpatch, before patch: 1234
	// with AutoUnpatch, after patch: 3456
	// inner AutoUnpatch, before patch: 3456
	// inner AutoUnpatch, after patch: 4567
	// with AutoUnpatch, after inner: 3456
	// with AutoUnpatch, patch again: 5678
	// after AutoUnpatch: 1234
}

func ExamplePatch_Origin() {
	var someVar = 1234

	AutoUnpatch(func() {
		fmt.Printf("before patch, testpkg.A: %v\n", testpkg.A())
		fmt.Printf("before patch, someVar: %v\n", someVar)

		patch1 := Mock().Target(testpkg.A).Return("ExamplePatch_Origin").Build()
		patch2 := PatchVar(&someVar, 5678)
		_ = patch1
		_ = patch2

		fmt.Printf("after patch, testpkg.A: %v\n", testpkg.A())
		fmt.Printf("after patch, someVar: %v\n", someVar)
	})

	// Output:
	// before patch, testpkg.A: testpkg.a
	// before patch, someVar: 1234
	// after patch, testpkg.A: ExamplePatch_Origin
	// after patch, someVar: 5678
}
