// +build release

package easy

func DEBUG(args ...interface{}) {}

func DEBUGSkip(skip int, args ...interface{}) {}

func SPEW(args ...interface{}) {}

func DUMP(args ...interface{}) {}
