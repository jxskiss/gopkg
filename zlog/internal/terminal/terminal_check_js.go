//go:build js

package terminal

func isTerminal(fd int) bool {
	return false
}
