package kvutil

import (
	"testing"
)

func Test_Key(t *testing.T) {
	var km KeyFactory
	key := km.NewKey("abc:def:%d:%s")
	want := "abc:def:1234567:x0BtEadepz6L"

	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_Key: got=%v want=%v", got, want)
	}
}

func Test_Key_CornerCase(t *testing.T) {
	cases := []struct {
		key  string
		args []any
		want string
	}{
		{
			key:  "%v:blah:%v",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "1234567:blah:x0BtEadepz6L",
		},
		{
			key:  "%v%v:blah:",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "1234567x0BtEadepz6L:blah:",
		},
		{
			key:  "%v:%v:blah",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "1234567:x0BtEadepz6L:blah",
		},
		{
			key:  "blah:%v%v",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "blah:1234567x0BtEadepz6L",
		},
	}

	var km KeyFactory
	for _, c := range cases {
		key := km.NewKey(c.key)
		got := key(c.args...)

		if got != c.want {
			t.Errorf("failed Test_Key_CornerCase: key=%v got=%v want=%v", c.key, got, c.want)
		}
	}
}

func Test_Key_CurlyBrace(t *testing.T) {
	cases := []struct {
		key  string
		args []any
		want string
	}{
		{
			key:  "{}:blah:{}",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "1234567:blah:x0BtEadepz6L",
		},
		{
			key:  "{}{}:blah:",
			args: []any{uint32(1234567), "x0BtEadepz6L"},
			want: "1234567x0BtEadepz6L:blah:",
		},
		{
			key:  "{}:{}:blah",
			args: []any{1234567, "x0BtEadepz6L"},
			want: "1234567:x0BtEadepz6L:blah",
		},
		{
			key:  "blah:{}{}",
			args: []any{uint64(1234567), "x0BtEadepz6L"},
			want: "blah:1234567x0BtEadepz6L",
		},
	}

	var kf KeyFactory
	for _, c := range cases {
		key := kf.NewKey(c.key)
		got := key(c.args...)

		if got != c.want {
			t.Errorf("failed Test_Key_CurlyBrace: key=%v got=%v want=%v", c.key, got, c.want)
		}
	}
}

func Test_Key_NamedArgs(t *testing.T) {
	var km KeyFactory
	key := km.NewKey("abc:{some_id}:{dummy}")
	want := "abc:1234567:x0BtEadepz6L"
	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_Key_NamedArgs: got=%v, want=%v", got, want)
	}
}

func Test_Key_WithArgNames(t *testing.T) {
	var km KeyFactory

	t.Run("sprintfKey", func(t *testing.T) {
		key := km.NewKey("{{some_id_1}_foo_bar}:{some_id_2}:%v", "some_id_1", "some_id_2")
		want := "{111_foo_bar}:222:333"
		got := key(111, 222, 333)
		if got != want {
			t.Errorf("failed Test_Key_WithArgNames")
		}
	})

	t.Run("builderKey", func(t *testing.T) {
		key := km.NewKey("{{some_id_1}_foo_bar}:{some_id_2}", "some_id_1", "some_id_2")
		want := "{111_foo_bar}:222"
		got := key(111, 222)
		if got != want {
			t.Errorf("failed Test_Key_WithArgNames: got= %v, want= %v", got, want)
		}
	})
}

func Test_Key_UnmatchedArgCount(t *testing.T) {
	var km KeyFactory

	key1 := km.NewKey("abc:{some_id}:{arg2}:{dummy1}:{dummy2}")
	got1 := key1(1234567, "x0BtEadepz6L")
	want1 := "abc:1234567:x0BtEadepz6L:{dummy1}:{dummy2}"
	if got1 != want1 {
		t.Errorf("failed Test_Key_UnmatchedArgCount: got1=%v, want1=%v", got1, want1)
	}

	key2 := km.NewKey("{{some_id_1}_foo_bar_count}:{arg2}:{dummy1}:{dummy2}",
		"some_id_1", "arg2")
	got2 := key2(1234567)
	want2 := "{1234567_foo_bar_count}:{arg2}:{dummy1}:{dummy2}"
	if got2 != want2 {
		t.Errorf("failed Test_Key_UnmatchedArgCount: got2=%v, want2=%v", got2, want2)
	}
}

func Test_SetPrefix(t *testing.T) {
	var km KeyFactory
	km.SetPrefix("some_prefix:")

	key := km.NewKey("abc:def:%d:%s")
	want := "some_prefix:abc:def:1234567:x0BtEadepz6L"
	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_SetPrefix: got=%v want=%v", got, want)
	}
}

var benchmarkData = []struct {
	format   string
	argNames []string
	args     []any
}{
	{"abc:{some_id}:{dummy}", nil,
		[]any{1234567, "x0BtEadepz6L"}},
	{"{}:blah:{}", nil,
		[]any{1234567, "x0BtEadepz6L"}},
	{"{{some_id_1}_foo_bar_count}:{some_id_2}",
		[]string{"some_id_1", "some_id_2"},
		[]any{1234567, "x0BtEadepz6L"}},
}

var benchmarkSprintfKeys []Key

func init() {
	km := KeyFactory{prefix: "my_some_prefix"}
	for i := 0; i < len(benchmarkData); i++ {
		x := benchmarkData[i]
		benchmarkSprintfKeys = append(benchmarkSprintfKeys, km.NewKey(x.format, x.argNames...))
	}
}

func Benchmark_Key_Sprintf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(benchmarkData); j++ {
			_ = benchmarkSprintfKeys[j](benchmarkData[j].args...)
		}
	}
}

func Benchmark_Key_Sprintf_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < len(benchmarkData); j++ {
				_ = benchmarkSprintfKeys[j](benchmarkData[j].args...)
			}
		}
	})
}
