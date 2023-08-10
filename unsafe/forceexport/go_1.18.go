//go:build gc && go1.18 && !go1.22

package forceexport

// functab is a copy type of runtime.functab.
//
//nolint:unused
type functab struct {
	entryoff uint32 // relative to runtime.text
	funcoff  uint32
}
