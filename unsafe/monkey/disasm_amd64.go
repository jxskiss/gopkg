package monkey

import (
	"fmt"

	"golang.org/x/arch/x86/x86asm"
)

func disassemble(buf []byte, required int) int {
	var pos int
	var err error
	var inst x86asm.Inst

	for pos < required {
		inst, err = x86asm.Decode(buf[pos:], 64)
		if err != nil {
			panic(fmt.Sprintf("monkey: %v", err))
		}
		if inst.Op == x86asm.RET {
			panic("monkey: function is too short to patch")
		}
		pos += inst.Len
	}
	return pos
}
