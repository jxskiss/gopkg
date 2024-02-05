//go:build appengine

package terminal

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return true
}
