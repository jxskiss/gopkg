// +build release

package easy

func DEBUG(args ...interface{})  {}
func PRETTY(args ...interface{}) {}
func SPEW(args ...interface{})   {}
func DUMP(args ...interface{})   {}

func DEBUGSkip(skip int, args ...interface{})  {}
func PRETTYSkip(skip int, args ...interface{}) {}
func SPEWSKip(skip int, args ...interface{})   {}
func DUMPSkip(skip int, args ...interface{})   {}
