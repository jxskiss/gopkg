package linkname

import "unsafe"

// GetPid returns the id of current p.
func GetPid() int {
	pid := Runtime_procPin()
	Runtime_procUnpin()
	return pid
}

//go:linkname sysBigEndian internal.sys.BigEndian
var sysBigEndian bool

// Runtime_readUnaligned32 reads memory pointed by p as a uint32 value.
// It performs the read with a native endianness.
//
// It is copied from runtime.readUnaligned32 but not linked to help inlining.
func Runtime_readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	if sysBigEndian {
		return uint32(q[3]) | uint32(q[2])<<8 | uint32(q[1])<<16 | uint32(q[0])<<24
	}
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

// Runtime_readUnaligned64 reads memory pointed by p as a uint64 value.
// It performs the read with a native endianness.
//
// It is copied from runtime.readUnaligned64 but not linked to help inlining.
func Runtime_readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	if sysBigEndian {
		return uint64(q[7]) | uint64(q[6])<<8 | uint64(q[5])<<16 | uint64(q[4])<<24 |
			uint64(q[3])<<32 | uint64(q[2])<<40 | uint64(q[1])<<48 | uint64(q[0])<<56
	}
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 |
		uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}
