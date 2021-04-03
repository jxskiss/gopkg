package monkey

import (
	"github.com/jxskiss/gopkg/reflectx"
)

func buildJmpDirective(to uintptr) []byte {
	if reflectx.PtrBitSize == 32 {
		return _buildJmpDirective_x86(to)
	}
	return _buildJmpDirective_amd64(to)
}

func _buildJmpDirective_x86(to uintptr) []byte {
	return []byte{
		0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24), // mov edx, to
		0xFF, 0x22,     // jmp DWORD PTR [edx]
	}
}

func _buildJmpDirective_amd64(to uintptr) []byte {
	return []byte{
		0x48, 0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // movabs rdx, to
		0xFF, 0x22,     // jmp QWORD PTR [rdx]
	}
}
