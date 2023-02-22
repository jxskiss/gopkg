package monkey

func branchInto(to uintptr) []byte {
	return []byte{
		0x48, 0xba,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // mov rdx, to
		0xff, 0x22,     // jmp DWORD PTR [rdx]
	}
}

//nolint:unused
func branchTo(to uintptr) []byte {
	return []byte{
		0x48, 0xba,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // mov rdx, to
		0xff, 0xe2,     // jmp DWORD [rdx]
	}
}
