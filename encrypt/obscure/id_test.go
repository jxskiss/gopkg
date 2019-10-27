package obscure

import (
	"encoding/json"
	"testing"
)

func Test_ID_String(t *testing.T) {
	x := 6590172069002560793
	want := "UyKHI4iwS81B"

	if got := ID(x).String(); got != want {
		t.Errorf("TestID_String failed: got=%v want=%v", got, want)
	}
}

func Test_ID_Decode(t *testing.T) {
	s := "UyKHI4iwS81B"
	want := int64(6590172069002560793)

	var id ID
	if err := id.Decode(s); err != nil {
		t.Errorf("TestID_Decode failed: err=%v", err)
	}
	if got := int64(id); got != want {
		t.Errorf("TestID_Decode failed: got=%v want=%v", got, want)
	}
}

func Test_ID_MarshalJSON(t *testing.T) {
	x := &testType{ID: 6590172069002560793}
	want := `{"id":"UyKHI4iwS81B"}`

	b, _ := json.Marshal(x)
	got := string(b)
	if got != want {
		t.Errorf("TestID_MarshalJSON failed: got=%v want=%v", got, want)
	}
}

func Test_ID_UnmarshalJSON(t *testing.T) {
	s := `{"id":"UyKHI4iwS81B"}`
	want := int64(6590172069002560793)

	x := &testType{}
	_ = json.Unmarshal([]byte(s), x)
	if got := int64(x.ID); got != want {
		t.Errorf("TestID_UnmarshalJSON failed: got=%v want=%v", got, want)
	}
}

type testType struct {
	ID ID `json:"id,omitempty"`
}

func Benchmark_ID_Encode(bb *testing.B) {
	x := 6590172069002560793

	for i := 0; i < bb.N; i++ {
		_ = ID(x).String()
	}
}

func Benchmark_ID_Decode(bb *testing.B) {
	s := "UyKHI4iwS81B"

	var id ID
	for i := 0; i < bb.N; i++ {
		_ = id.Decode(s)
	}
}
