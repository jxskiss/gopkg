//go:build gc && !go1.18

package forceexport

type functab struct {
	entryoff uintptr
	funcoff  uintptr
}
