//go:build !go1.18
// +build !go1.18

package forceexport

type functab struct {
	entryoff uintptr
	funcoff  uintptr
}
