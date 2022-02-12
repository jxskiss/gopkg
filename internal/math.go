package internal

func NextPowerOfTwo(x uint) uint {
	if x <= 1 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32

	return x + 1
}
