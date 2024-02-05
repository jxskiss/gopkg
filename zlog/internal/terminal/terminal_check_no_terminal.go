//go:build js || nacl || plan9

package terminal

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return false
}
