package base62

import (
	"math"
)

func FormatUint(num uint64) []byte {
	dst := make([]byte, 0)
	return AppendUint(dst, num)
}

func AppendUint(dst []byte, num uint64) []byte {
	if num == 0 {
		dst = append(dst, encodeStd[0])
		return dst
	}
	for num > 0 {
		r := num % base
		dst = append(dst, encodeStd[r])
		num /= base
	}
	return dst
}

func ParseUint(src []byte) (uint64, error) {
	decTable := stdEncoding.decodeMap

	var ret uint64
	for i, c := range src {
		x := decTable[c]
		if x == 0xFF {
			return 0, CorruptInputError(i)
		}
		ret += uint64(x) * uint64(math.Pow(base, float64(i)))
	}
	return ret, nil
}
