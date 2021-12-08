//go:build go1.18
// +build go1.18

package forceexport

type functab struct {
	entryoff uint32 // relative to runtime.text
	funcoff  uint32
}
