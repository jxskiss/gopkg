package easy

import "testing"

func TestSlashJoin(t *testing.T) {
	path1 := []string{"/a", "b", "c.png"}
	want1 := "/a/b/c.png"
	if got1 := SlashJoin(path1...); got1 != want1 {
		t.Errorf("unexpectetd slash join result: %v", got1)
	}

	path2 := []string{"/a/", "b/", "/c.png"}
	want2 := "/a/b/c.png"
	if got2 := SlashJoin(path2...); got2 != want2 {
		t.Errorf("unexpected slash join result: %v", got2)
	}
}

func BenchmarkSlashJoin(b *testing.B) {
	path := []string{"/a", "b", "c.png"}
	for i := 0; i < b.N; i++ {
		_ = SlashJoin(path...)
	}
}
