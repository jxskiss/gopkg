package obscure

import (
	"bytes"
	"testing"
)

func Test_Base64_Encode(t *testing.T) {
	src := []byte("0123456789abcdefghij")
	dst := B64Encode(src)
	want := []byte("urJfurMxysXVWFglAb75NEj+pwgp")

	if !bytes.Equal(want, dst) {
		t.Errorf("Test_Encode failed: got=%v want=%v", string(dst), string(want))
	}
}

func Test_Base64_Decode(t *testing.T) {
	src := []byte("urJfurMxysXVWFglAb75NEj+pwgp")
	dst, _ := B64Decode(src)
	want := []byte("0123456789abcdefghij")

	if !bytes.Equal(want, dst) {
		t.Errorf("Test_Decode failed: got=%v want=%v", string(dst), string(want))
	}
}

func Test_SetBase64Table(t *testing.T) {
	table := Table{
		"UkRY/Ax_-Vq+EmirMDTHvSN0jBXClZc4oyG89bWuOh675f1gPdtJeQKzp2FIansL", // 0
		"SJTi7EOGvW1hDtaM0Yk-Pcj3nwF+ub4LyVp_forqgxdC8NlmIZBQe6UsKz92A5X/", // 1
		"2ZHItnA1PsXrxDT7hRcpVq8kSGi3fNLb6W9JB/-md4FvEuaCl5oyOMKwYe_zjU+g", // 2
		"Yt8bNRKWxZmjs2z1Bu/gTaS4XyhOVe6wP3-n7fEJcd9ACprUIHk5DGF_+Mq0lLov", // 3
		"dnNe0DLFIlj+OTcw1uYtCmpg2zxK/Py-QkoVM4a9sBhrRHW8Avq_Zif3XS67UbJ5", // 4
		"hnvAuIO0lZSi87j_XxVDMsBHzCYKebG1pykdP3Rw9+qU2W4gta5N6LEofm-TrFcQ", // 5
		"wLHBfbvDmFRYquCSiQ-az94AphG560r7OZygt12NKlVXJ+Td8I_konE/MPxj3sWe", // 6
		"ye/Ks5MIX8j6qBzxwgcvJfm7Q3UCkZSpRGlV2AanFYhrDut+i_L14-TEo9PHWdOb", // 7
		"Vo6IkEsjBdJO5/_Ma2vrbxiAW+ZRw8H3tpmlycD1hL-gGqXCNUnF0TuQ79feSPYK", // 8
		"GrAxeX+93c_B5YgZUzaJ7IEuNqDMkwfFVWj4SKLdmvbHPh-Os/n0RplQyo21t6i8", // 9
		"EVq6XrC18dQ+GlZ2FRgM05HJWm_LnTBhfSpyDcwusxezjPbiaO7kIt/o9-AUNvKY", // 10
		"I2YLsXgdHBq5ivW1TGeVS_jbDUOoczm-xpy7kE6+h03CPMwrl9NJR/fF4ZKQ8nta", // 11
		"0vP+8UA5ELI6RwNXuTksQZWhe4qKirtC23MdnHjmOa/7z1-gYb9ylFBVcGfJopS_", // 12
	}
	SetBase64Table(table)

	src := []byte("0123456789abcdefghij")
	dst := B64Encode(src)

	got, err := B64Decode(dst)
	if err != nil {
		t.Errorf("failed decode base64: err= %v", err)
	}
	if !bytes.EqualFold(src, got) {
		t.Errorf("Test_SetBase64Table failed: got=%v src=%v", string(got), string(src))
	}
}

func Benchmark_Base64_Encode(bb *testing.B) {
	src := []byte("0123456789abcdefghij")
	for i := 0; i < bb.N; i++ {
		_ = B64Encode(src)
	}
}

func Benchmark_Base64_Decode(bb *testing.B) {
	src := []byte("Xtjr6t0O2fNXD8Z/wcUy.5KJShZS")
	for i := 0; i < bb.N; i++ {
		_, _ = B64Decode(src)
	}
}
