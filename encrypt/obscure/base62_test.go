package obscure

import (
	"bytes"
	"testing"
)

func Test_Base62_Encode(t *testing.T) {
	src := []byte("0123456789abcdefghij")
	dst := B62Encode(src)
	want := []byte("uvy18nslsP5DMmy67aMB5qrQrCx6")

	if !bytes.Equal(want, dst) {
		t.Errorf("Test_Encode failed: got=%v want=%v", string(dst), string(want))
	}
}

func Test_Base62_Decode(t *testing.T) {
	src := []byte("uvy18nslsP5DMmy67aMB5qrQrCx6")
	dst, _ := B62Decode(src)
	want := []byte("0123456789abcdefghij")

	if !bytes.Equal(want, dst) {
		t.Errorf("Test_Decode failed: got=%v want=%v", string(dst), string(want))
	}
}

func Test_SetBase62Table(t *testing.T) {
	table := Table{
		"zorED0PmSCeKbVLwgyIlR38UHi1u5pNZ9FBXhc4OfdJMQW76jtqYxnTkasvG2A", // 0
		"XNmKERFZ67QAkq3UY5stIbJxDleB9OHdGjMcao1PSwTLzhVn8r4yfC2Wigvu0p", // 1
		"hyPtOofRWq4J6cKeSLdQ1sGiMvN2kZlzFIAga8509nu3VxUjE7bwrYXpCDmBTH", // 2
		"gqFQ0ZYTcvO6WH2XsaBKPpol3t1uCwy9e8jUrL7dESiIkfMxAG4hDRnJzbmV5N", // 3
		"izV4geYSCqZAkulJ1sIrv7tpxW0aKM8bQBFd6Rc2ynTPLNj3Ohwo59UEXmDfHG", // 4
		"wLxKSXkVU9Nl5GMzYtBhPAoC2J6umTHI3yepqg71f8QvrWnsRjdia0bZEcD4OF", // 5
		"wGBRDdPckSF4My83mIsjoW1N7tqlg0eAnO6LpYv5CEhxK9uUZrXi2TzJaVbfHQ", // 6
		"C2UVljE7pNO8Ro1ecb9wy4Zd0nYGHLWSKPTJit6xMhAk3FuXfQIBrsgDaq5vmz", // 7
		"usIonOALieafl9Jbgd0YGDC6y37BUXqmKPSQ2rtM4jV5h1xpFRv8ZcEHzwNkTW", // 8
		"nPsK0hTNYMJ4XxOd3AFoZBgSvH918QLG2aWcjVEukftCqIw5imR7D6bUpreyzl", // 9
		"SeLPfzljd02Ar5QZJ1O4RXEa6kcuhV8stiN7IDFqKHBWTbxyCYMmwo93gUnpGv", // 10
		"RSqCHp1u9hLi53cDOZyrd78TxBoWseflPXg6FUQaVj4zk0ENIKGntAY2JbwMmv", // 11
		"7wryYJtuU2SD6F4BNQ0blcZeChAWv1aIxz3LPKd8OjGTHomqEVipknX9sgfMR5", // 12
	}
	SetBase62Table(table)

	src := []byte("0123456789abcdefghij")
	dst := B62Encode(src)

	got, err := B62Decode(dst)
	if err != nil {
		t.Errorf("failed decode base62: err= %v", err)
	}
	if !bytes.EqualFold(src, got) {
		t.Errorf("Test_SetBase64Table failed: got=%v src=%v", string(got), string(src))
	}
}

func Benchmark_Base62_Encode(bb *testing.B) {
	src := []byte("0123456789abcdefghij")
	for i := 0; i < bb.N; i++ {
		_ = B62Encode(src)
	}
}

func Benchmark_Base62_Decode(bb *testing.B) {
	src := []byte("uvy18nslsP5DMmy67aMB5qrQrCx6")
	for i := 0; i < bb.N; i++ {
		_, _ = B62Decode(src)
	}
}
