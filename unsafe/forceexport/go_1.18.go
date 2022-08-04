//go:build gc && go1.18 && !go1.20

package forceexport

type functab struct {
	entryoff uint32 // relative to runtime.text
	funcoff  uint32
}
