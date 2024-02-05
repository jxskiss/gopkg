package terminal

import (
	"fmt"
	"io"
)

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Gray
)

type Color uint8

// Format adds the coloring to the given string.
func (c Color) Format(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

func CheckIsTerminal(w io.Writer) bool {
	return checkIfTerminal(w)
}
