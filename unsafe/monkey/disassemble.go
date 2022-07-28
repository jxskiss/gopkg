package monkey

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/arch/x86/x86asm"
)

var _go1VersionOnce sync.Once
var _go1MinorVersion int

func getGo1MinorVersion() int {
	_go1VersionOnce.Do(func() {
		version := runtime.Version()
		minorIdx := strings.Index(version, "go1.")
		if minorIdx >= 0 {
			suffix := version[minorIdx+4:]
			splitter := regexp.MustCompile(`[.\-]`)
			parts := splitter.Split(suffix, -1)
			_go1MinorVersion, _ = strconv.Atoi(parts[0])
		}
		if _go1MinorVersion <= 0 {
			panic("monkey: cannot determine the Go version")
		}
	})
	return _go1MinorVersion
}

var branchingOps = map[x86asm.Op]bool{
	x86asm.CALL:   true,
	x86asm.IRET:   true,
	x86asm.IRETD:  true,
	x86asm.IRETQ:  true,
	x86asm.JA:     true,
	x86asm.JAE:    true,
	x86asm.JB:     true,
	x86asm.JBE:    true,
	x86asm.JCXZ:   true,
	x86asm.JE:     true,
	x86asm.JECXZ:  true,
	x86asm.JG:     true,
	x86asm.JGE:    true,
	x86asm.JL:     true,
	x86asm.JLE:    true,
	x86asm.JMP:    true,
	x86asm.JNE:    true,
	x86asm.JNO:    true,
	x86asm.JNP:    true,
	x86asm.JNS:    true,
	x86asm.JO:     true,
	x86asm.JP:     true,
	x86asm.JRCXZ:  true,
	x86asm.JS:     true,
	x86asm.LCALL:  true,
	x86asm.LJMP:   true,
	x86asm.LOOP:   true,
	x86asm.LOOPE:  true,
	x86asm.LOOPNE: true,
	x86asm.LRET:   true,
	x86asm.RET:    true,
}

func disassemble(buf []byte, required int) int {
	var pos int
	var err error
	var inst x86asm.Inst

	if getGo1MinorVersion() >= 17 {
		pos = skipPrologue(buf)
		required += pos
	}

	for pos < required {
		if inst, err = x86asm.Decode(buf[pos:], 64); err != nil {
			panic(err)
		}
		if branchingOps[inst.Op] {
			panic("monkey: function is too short to patch")
		}
		pos += inst.Len
	}
	return pos
}

func skipPrologue(buf []byte) int {
	pos := 0
	for i := 0; i < 3; i++ {
		inst, err := x86asm.Decode(buf[pos:], 64)
		if err != nil {
			panic(err)
		}
		pos += inst.Len
		if inst.Op == x86asm.JBE {
			return pos
		}
	}
	return 0
}
